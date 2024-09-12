// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// Config configures [Load].
type Config struct {

	// Dir is the directory in which to run the build system's query tool
	// that provides information about the packages.
	// If Dir is empty, the tool is run in the current directory.
	//
	// Package patterns given to [ParsePackages] are relative to this directory.
	Dir string

	// BuildFlags is a list of command-line flags to be passed through to
	// the build system's query tool.
	BuildFlags []string

	// Env is the environment to use when invoking the build system's query tool.
	// If Env is nil, the current environment is used.
	// As in os/exec's Cmd, only the last value in the slice for
	// each environment key is used. To specify the setting of only
	// a few variables, append to the current environment, as in:
	//
	//	opt.Env = append(os.Environ(), "GOOS=plan9", "GOARCH=386")
	//
	Env []string

	// If Tests is set, the loader includes not just the packages
	// matching a particular pattern but also any related test packages,
	// including test-only variants of the package and the test executable.
	//
	// For example, when using the go command, loading "fmt" with Tests=true
	// returns four packages, with IDs "fmt" (the standard package),
	// "fmt [fmt.test]" (the package as compiled for the test),
	// "fmt_test" (the test functions from source files in package fmt_test),
	// and "fmt.test" (the test binary).
	//
	// In build systems with explicit names for tests,
	// setting Tests may have no effect.
	Tests bool

	// TypeChecking enables type checking for utilities like [Bast.ResolveBasicType].
	TypeChecking bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Dir:          ".",
		TypeChecking: true,
	}
}

// Load loads packages specified by pattern and returns a *Bast of it
// or an error.
//
// An optional config configures what is parsed, paths, etc.
// See [Config].
func Load(config *Config, patterns ...string) (bast *Bast, err error) {

	bast = new()
	if bast.config = config; bast.config == nil {
		bast.config = DefaultConfig()
	}

	var mode = packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName
	if config.TypeChecking {
		mode |= packages.NeedTypes | packages.NeedDeps | packages.NeedImports
	}

	var (
		cfg = &packages.Config{
			Mode:       mode,
			Dir:        bast.config.Dir,
			BuildFlags: bast.config.BuildFlags,
			Env:        bast.config.Env,
			Tests:      bast.config.Tests,
		}
		pkgs []*packages.Package
	)

	if pkgs, err = packages.Load(cfg, patterns...); err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var errs []error
			for _, e := range pkg.Errors {
				errs = append(errs, e)
			}
			return nil, fmt.Errorf("parse packages: %w", errors.Join(errs...))
		}
		bast.parsePackage(pkg, bast.packages)
	}

	return
}

// parsePackage parses a package into a bast package, adds it to PackageMap
// keying it by its package path.
func (self *Bast) parsePackage(in *packages.Package, out *PackageMap) {

	var pkg = NewPackage()
	pkg.Name = in.Name

	for idx, file := range in.Syntax {
		self.parseFile(in.CompiledGoFiles[idx], file, pkg.Files)
	}

	pkg.pkg = in
	pkg.Path = in.PkgPath

	out.Put(pkg.Path, pkg)

	return
}

// parseFile parses an ast file parsed from fileName into a bast [File] and
// adds it to [FileMap], keyed by filename.
func (self *Bast) parseFile(fileName string, in *ast.File, out *FileMap) {

	var file = NewFile()
	file.Name = filepath.Base(fileName)
	file.FileName = fileName

	for _, comment := range in.Comments {
		var cg []string
		self.parseCommentGroup(comment, &cg)
		file.Comments = append(file.Comments, cg)
	}

	self.parseCommentGroup(in.Doc, &file.Doc)

	for _, imprt := range in.Imports {
		self.parseImportSpec(imprt, file.Imports)
	}

	for _, d := range in.Decls {
		self.parseDeclarations(d.(ast.Node), file.Declarations)
	}

	out.Put(file.FileName, file)

	return
}

func (self *Bast) parseDeclarations(in ast.Node, out *DeclarationMap) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					self.parseConsts(s, out)
				}
			}
		case token.VAR:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					self.parseVar(s, out)
				}
			}
		case token.TYPE:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Assign != token.NoPos {
						continue
					}
					switch s.Type.(type) {
					case *ast.InterfaceType:
						self.parseInterface(n, s, out)
					case *ast.StructType:
						self.parseStruct(n, s, out)
					case *ast.ArrayType:
						self.parseType(n, s, out)
					case *ast.FuncType:
						self.parseFuncType(n, s, out)
					case *ast.Ident:
						self.parseType(n, s, out)
					case *ast.ChanType:
						self.parseType(n, s, out)
					case *ast.MapType:
						self.parseType(n, s, out)
					case *ast.SelectorExpr:
						self.parseType(n, s, out)
					default:
						fmt.Println(self.printExpr(s))
					}
				}
			}
		}
	case *ast.FuncDecl:
		if n.Recv != nil {
			self.parseMethod(n, out)
		} else {
			self.parseFunc(n, out)
		}
	}
	return
}

// parseCommentGroup a comment group into a string slice, line per entry.
func (self *Bast) parseCommentGroup(in *ast.CommentGroup, out *[]string) {
	if in == nil {
		return
	}
	for _, entry := range in.List {
		*out = append(*out, entry.Text)
	}
	return
}

func (self *Bast) parseImportSpec(in *ast.ImportSpec, out *ImportSpecMap) {
	var val = NewImport()
	if in.Name != nil {
		val.Name = self.printExpr(in.Name)
	}
	val.Path = self.printExpr(in.Path)
	self.parseCommentGroup(in.Doc, &val.Doc)
	out.Put(val.Path, val)
	return
}

func (self *Bast) parseConsts(in *ast.ValueSpec, out *DeclarationMap) {
	var lastType string
	for i := 0; i < len(in.Names); i++ {
		var val = NewConst()
		val.Name = self.printExpr(in.Names[i])
		self.parseCommentGroup(in.Doc, &val.Doc)
		if in.Type != nil {
			val.Type = self.printExpr(in.Type)
			lastType = val.Type
		} else if lastType != "" {
			val.Type = lastType
		}
		if in.Values != nil {
			val.Value = self.printExpr(in.Values[i])
		}
		out.Put(val.Name, val)
	}
}

func (self *Bast) parseVar(in *ast.ValueSpec, out *DeclarationMap) {
	for i := 0; i < len(in.Names); i++ {
		var val = NewVar()
		val.Name = self.printExpr(in.Names[i])
		self.parseCommentGroup(in.Doc, &val.Doc)

		val.Type = self.printExpr(in.Type)
		if in.Values != nil {
			if len(in.Values) == 1 {
				val.Value = self.printExpr(in.Values[0])

			} else {
				val.Value = self.printExpr(in.Values[i])
			}
		}
		out.Put(val.Name, val)
	}
}

func (self *Bast) parseFunc(in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewFunc()
	val.Name = self.printExpr(in.Name)
	self.parseCommentGroup(in.Doc, &val.Doc)
	self.parseFieldList(in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(in.Type.Params, val.Params)
	self.parseFieldList(in.Type.Results, val.Results)
	out.Put(val.Name, val)
}

func (self *Bast) parseMethod(in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewMethod()
	self.parseCommentGroup(in.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)

	if in.Recv != nil {
		val.Receiver = NewField()
		if len(in.Recv.List[0].Names) > 0 {
			val.Receiver.Name = self.printExpr(in.Recv.List[0].Names[0])
		}
		val.Receiver.Type = self.printExpr(in.Recv.List[0].Type)
	}

	self.parseFieldList(in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(in.Type.Params, val.Params)
	self.parseFieldList(in.Type.Results, val.Results)
	out.Put(val.Name, val)
	return
}

func (self *Bast) parseFuncType(g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var val = NewFunc()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	var ft = in.Type.(*ast.FuncType)
	self.parseFieldList(ft.TypeParams, val.TypeParams)
	self.parseFieldList(ft.Params, val.Params)
	self.parseFieldList(ft.Results, val.Results)
	out.Put(val.Name, val)
	return
}

func (self *Bast) parseType(g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var val = NewType()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	val.Type = self.printExpr(in.Type)
	val.IsAlias = in.Assign.IsValid()
	out.Put(val.Name, val)
	return
}

func (self *Bast) parseFieldList(in *ast.FieldList, out *FieldMap) {
	if in == nil {
		return
	}
	for idx, field := range in.List {
		if field.Names == nil {
			var key = fmt.Sprintf("%s [%d]", self.printExpr(field.Type), idx)
			out.Put(key, &Field{
				Type: self.printExpr(field.Type),
			})
			continue
		}
		for _, name := range field.Names {
			var val = NewField()
			self.parseCommentGroup(field.Doc, &val.Doc)
			val.Name = self.printExpr(name)
			val.Type = self.printExpr(field.Type)
			val.Tag = self.printExpr(field.Tag)
			out.Put(val.Name, val)
		}
	}
}

func (self *Bast) parseInterface(g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {

	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	var val = NewInterface()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)

	for _, method := range it.Methods.List {
		if len(method.Names) == 0 {
			var intf = NewField()
			self.parseCommentGroup(method.Doc, &intf.Doc)
			intf.Type = self.printExpr(method.Type)
			intf.Name = intf.Type
			intf.Unnamed = true
			val.Interfaces.Put(intf.Type, intf)
		} else {
			var m = NewMethod()
			self.parseCommentGroup(method.Doc, &m.Doc)
			m.Name = self.printExpr(method.Names[0])
			var ft = method.Type.(*ast.FuncType)
			self.parseFieldList(ft.TypeParams, m.TypeParams)
			self.parseFieldList(ft.Params, m.Params)
			self.parseFieldList(ft.Results, m.Results)
			val.Methods.Put(m.Name, m)
		}
	}

	out.Put(val.Name, val)

	return
}

func (self *Bast) parseStruct(g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}
	var val = NewStruct()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	for _, field := range st.Fields.List {
		self.parseStructField(field, val.Fields)
	}

	out.Put(val.Name, val)

	return
}

func (self *Bast) parseStructField(in *ast.Field, out *FieldMap) {

	var val = NewField()

	if len(in.Names) == 0 {
		self.parseCommentGroup(in.Doc, &val.Doc)
		val.Unnamed = true
		val.Type = self.printExpr(in.Type)
		val.Tag = self.printExpr(in.Tag)
		out.Put(val.Name, val)
		return
	}

	for _, name := range in.Names {
		self.parseCommentGroup(in.Doc, &val.Doc)
		val.Name = self.printExpr(name)
		val.Type = self.printExpr(in.Type)
		val.Tag = self.printExpr(in.Tag)
		out.Put(val.Name, val)
	}

	return
}
