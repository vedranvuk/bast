// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"time"

	"github.com/vedranvuk/strutils"
	"golang.org/x/tools/go/packages"
)

// Config configures [Load].
type Config struct {

	// -- Build system configuration --

	// Dir is the directory in which to run the build system's query tool
	// that provides information about the packages.
	// If Dir is empty, the tool is run in the current directory.
	//
	// Package patterns given to [ParsePackages] are relative to this directory.
	//
	// Default is "." which sets the build dir to current directory.
	Dir string `json:"dir,omitempty"`

	// BuildFlags is a list of command-line flags to be passed through to
	// the build system's query tool.
	BuildFlags []string `json:"buildFlags,omitempty"`

	// Env is the environment to use when invoking the build system's query tool.
	// If Env is nil, the current environment is used.
	// As in os/exec's Cmd, only the last value in the slice for
	// each environment key is used. To specify the setting of only
	// a few variables, append to the current environment, as in:
	//
	//	opt.Env = append(os.Environ(), "GOOS=plan9", "GOARCH=386")
	//
	Env []string `json:"env,omitempty"`

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
	Tests bool `json:"tests,omitempty"`

	// -- Bast configuration --

	// TypeChecking enables type checking to enable few type resolution
	// related utilities like [Bast.ResolveBasicType].
	//
	// Default: true
	TypeChecking bool `json:"typeChecking,omitempty"`

	// TypeCheckingErrors if enabled will return errors during [Load] if
	// typechecking or parsed packages failed.
	//
	// Default: true
	TypeCheckingErrors bool `json:"typeCheckingErrors,omitempty"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Dir:                ".",
		TypeChecking:       true,
		TypeCheckingErrors: true,
	}
}

// Default is an alias for [DefaultConfig].
func Default() *Config { return DefaultConfig() }

// Load loads packages specified by pattern and returns a *Bast of it
// or an error.
//
// An optional config configures the underlying go build system
// and other details. See [Config].
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
		if len(pkg.Errors) > 0 && config.TypeCheckingErrors {
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
	var pkg = NewPackage(in.Name, in.PkgPath, in)
	for idx, file := range in.Syntax {
		self.parseFile(pkg, in.CompiledGoFiles[idx], file, pkg.Files)
	}
	pkg.bast = self
	out.Put(pkg.Path, pkg)
	return
}

// parseFile parses an ast file parsed from fileName into a bast [File] and
// adds it to [FileMap], keyed by filename.
func (self *Bast) parseFile(pkg *Package, fileName string, in *ast.File, out *FileMap) {

	var file = NewFile(pkg, fileName)

	for _, comment := range in.Comments {
		var cg []string
		self.parseCommentGroup(comment, &cg)
		file.Comments = append(file.Comments, cg)
	}

	self.parseCommentGroup(in.Doc, &file.Doc)

	for _, imp := range in.Imports {
		self.parseImportSpec(imp, file.Imports)
	}

	for _, d := range in.Decls {
		self.parseDeclaration(file, d.(ast.Node), file.Declarations)
	}

	out.Put(file.Name, file)

	return
}

// parseDeclaration parses in node into a DeclarationMap out.
func (self *Bast) parseDeclaration(file *File, in ast.Node, out *DeclarationMap) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.VAR:
			self.parseVars(file, n, out)
		case token.CONST:
			self.parseConsts(file, n, out)
		case token.TYPE:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// if s.Assign != token.NoPos {
					// 	continue
					// }
					switch s.Type.(type) {
					case *ast.InterfaceType:
						self.parseInterface(file, n, s, out)
					case *ast.StructType:
						self.parseStruct(file, n, s, out)
					case *ast.ArrayType:
						self.parseType(file, n, s, out)
					case *ast.FuncType:
						self.parseFuncType(file, n, s, out)
					case *ast.Ident:
						self.parseType(file, n, s, out)
					case *ast.ChanType:
						self.parseType(file, n, s, out)
					case *ast.MapType:
						self.parseType(file, n, s, out)
					case *ast.SelectorExpr:
						self.parseType(file, n, s, out)
					default:
						fmt.Println(self.printExpr(s))
					}
				}
			}
		}
	case *ast.FuncDecl:
		if n.Recv != nil {
			self.parseMethod(file, n, out)
		} else {
			self.parseFunc(file, n, out)
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

// parseImportSpec parses import spec into a map keyed by path.
func (self *Bast) parseImportSpec(in *ast.ImportSpec, out *ImportSpecMap) {
	var val = NewImport(
		self.printExpr(in.Name),
		"",
	)
	val.Path, _ = strutils.UnquoteDouble(self.printExpr(in.Path))
	self.parseCommentGroup(in.Doc, &val.Doc)
	out.Put(val.Path, val)
	return
}

// parseVars parses a GenDecl in of vars into a DeclarationMap out.
func (self *Bast) parseVars(file *File, in *ast.GenDecl, out *DeclarationMap) {

	for _, spec := range in.Specs {

		var vspec, ok = spec.(*ast.ValueSpec)
		if !ok {
			return
		}

		for i := 0; i < len(vspec.Names); i++ {

			var val = NewVar(
				file,
				self.printExpr(vspec.Names[i]),
				self.printExpr(vspec.Type),
			)

			self.parseCommentGroup(in.Doc, &val.Doc)
			self.parseCommentGroup(vspec.Doc, &val.Doc)

			if vspec.Values != nil {
				if len(vspec.Values) == 1 {
					val.Value = self.printExpr(vspec.Values[0])
				} else {
					val.Value = self.printExpr(vspec.Values[i])
				}
			}

			out.Put(val.Name, val)
		}
	}
}

// parseVars parses a GenDecl in of consts into a DeclarationMap out.
func (self *Bast) parseConsts(file *File, in *ast.GenDecl, out *DeclarationMap) {
	for _, spec := range in.Specs {

		var vspec, ok = spec.(*ast.ValueSpec)
		if !ok {
			return
		}

		for i := 0; i < len(vspec.Names); i++ {

			var val = NewConst(
				file,
				self.printExpr(vspec.Names[i]),
				self.printExpr(vspec.Type),
			)

			self.parseCommentGroup(in.Doc, &val.Doc)
			self.parseCommentGroup(vspec.Doc, &val.Doc)

			if vspec.Values != nil {
				if len(vspec.Values) == 1 {
					val.Value = self.printExpr(vspec.Values[0])
				} else {
					val.Value = self.printExpr(vspec.Values[i])
				}
			}

			out.Put(val.Name, val)
		}
	}
}

// parseFunc parses in func decl into DeclarationMap out.
func (self *Bast) parseFunc(file *File, in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewFunc(file, self.printExpr(in.Name))
	self.parseCommentGroup(in.Doc, &val.Doc)
	self.parseFieldList(file, in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(file, in.Type.Params, val.Params)
	self.parseFieldList(file, in.Type.Results, val.Results)
	out.Put(val.Name, val)
}

// parseMethod parses in method decl into DeclarationMap out.
func (self *Bast) parseMethod(file *File, in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewMethod(file, self.printExpr(in.Name))
	self.parseCommentGroup(in.Doc, &val.Doc)

	if in.Recv != nil {
		val.Receiver = NewField(file, "")
		if len(in.Recv.List[0].Names) > 0 {
			val.Receiver.Name = self.printExpr(in.Recv.List[0].Names[0])
		}
		// Parse out the bare receiver type name. Exclude star and type params.
		switch t := in.Recv.List[0].Type.(type) {
		case *ast.Ident:
			val.Receiver.Type = t.Name
		case *ast.IndexExpr:
			val.Receiver.Type = self.printExpr(t.X)
		case *ast.StarExpr:
			if p, ok := t.X.(*ast.IndexExpr); ok {
				val.Receiver.Type = self.printExpr(p.X)
			} else {
				val.Receiver.Type = self.printExpr(t.X)
			}
			val.Receiver.Pointer = true
		default:
			time.Sleep(1 * time.Millisecond)
			_ = t
		}
		// val.Receiver.Type = self.printExpr(in.Recv.List[0].Type)
	}

	self.parseFieldList(file, in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(file, in.Type.Params, val.Params)
	self.parseFieldList(file, in.Type.Results, val.Results)
	out.Put(val.Name, val)
	return
}

// parseFuncType parses func type spec in into a DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Bast) parseFuncType(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var val = NewFunc(file, self.printExpr(in.Name))
	self.parseCommentGroup(g.Doc, &val.Doc)
	var ft = in.Type.(*ast.FuncType)
	self.parseFieldList(file, in.TypeParams, val.TypeParams)
	self.parseFieldList(file, ft.Params, val.Params)
	self.parseFieldList(file, ft.Results, val.Results)
	out.Put(val.Name, val)
	return
}

// parseType parses type spec in into a DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Bast) parseType(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var val = NewType(
		file,
		self.printExpr(in.Name),
		self.printExpr(in.Type),
	)
	self.parseCommentGroup(g.Doc, &val.Doc)
	self.parseCommentGroup(in.Doc, &val.Doc)
	self.parseFieldList(file, in.TypeParams, val.TypeParams)
	val.IsAlias = in.Assign.IsValid()
	out.Put(val.Name, val)
	return
}

// parseFieldList parses in field list into FieldMap out.
func (self *Bast) parseFieldList(file *File, in *ast.FieldList, out *FieldMap) {

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
			var val = NewField(file, self.printExpr(name))
			self.parseCommentGroup(field.Doc, &val.Doc)
			val.Type = self.printExpr(field.Type)
			val.Tag = self.printExpr(field.Tag)
			out.Put(val.Name, val)
		}
	}
}

// parseStruct parses a struct declaration in into DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Bast) parseStruct(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {

	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}

	var val = NewStruct(file, self.printExpr(in.Name))
	self.parseCommentGroup(g.Doc, &val.Doc)
	self.parseCommentGroup(in.Doc, &val.Doc)

	for _, field := range st.Fields.List {
		self.parseStructField(file, field, val.Fields)
	}

	self.parseFieldList(file, in.TypeParams, val.TypeParams)

	out.Put(val.Name, val)

	return
}

// parseStructField parses a struct field in into a FieldMap out.
func (self *Bast) parseStructField(file *File, in *ast.Field, out *FieldMap) {

	var val = NewField(file, "")

	// Unnamed field.
	if len(in.Names) == 0 {
		self.parseCommentGroup(in.Doc, &val.Doc)
		val.Unnamed = true
		val.Name = self.printExpr(in.Type)
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

// parseStruct parses an interface declaration in into DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Bast) parseInterface(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {

	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	var val = NewInterface(file, self.printExpr(in.Name))
	self.parseCommentGroup(g.Doc, &val.Doc)

	for _, method := range it.Methods.List {
		if len(method.Names) == 0 {
			var intf = NewField(file, self.printExpr(method.Type))
			self.parseCommentGroup(method.Doc, &intf.Doc)
			intf.Type = self.printExpr(method.Type)
			intf.Unnamed = true
			val.Interfaces.Put(intf.Type, intf)
		} else {
			var m = NewMethod(file, self.printExpr(method.Names[0]))
			self.parseCommentGroup(method.Doc, &m.Doc)
			var ft = method.Type.(*ast.FuncType)
			self.parseFieldList(file, ft.TypeParams, m.TypeParams)
			self.parseFieldList(file, ft.Params, m.Params)
			self.parseFieldList(file, ft.Results, m.Results)
			val.Methods.Put(m.Name, m)
		}
	}

	out.Put(val.Name, val)

	return
}
