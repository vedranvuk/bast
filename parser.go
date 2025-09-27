// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"reflect"

	"github.com/vedranvuk/strutils"
	"golang.org/x/tools/go/packages"
)

// Config configures the loading and parsing behavior of Load.
type Config struct {

	// Dir is the base directory for the build system's query tool.
	// If empty, the current directory is used.
	//
	// Package patterns in Load are relative to this directory.
	// Default is ".".
	Dir string `json:"dir,omitempty"`

	// BuildFlags are additional command-line flags passed to the build system's query tool.
	BuildFlags []string `json:"buildFlags,omitempty"`

	// Env is the environment variables used when invoking the build system's query tool.
	// If nil, the current environment is used.
	// Only the last value for each key is used.
	//
	// To set specific variables, append to os.Environ():
	//     opt.Env = append(os.Environ(), "GOOS=plan9", "GOARCH=386")
	Env []string `json:"env,omitempty"`

	// Tests, if true, includes related test packages when loading.
	// This includes test variants of the package and the test executable.
	Tests bool `json:"tests,omitempty"`

	// TypeChecking enables type checking during loading, required for type resolution utilities
	// like Bast.ResolveBasicType.
	// Default is true.
	TypeChecking bool `json:"typeChecking,omitempty"`

	// TypeCheckingErrors, if true, causes Load to return an error if type checking fails
	// or if any loaded package has errors.
	// Default is true.
	TypeCheckingErrors bool `json:"typeCheckingErrors,omitempty"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Dir:                ".",
		TypeChecking:       true,
		TypeCheckingErrors: true,
	}
}

// Default is an alias for DefaultConfig.
func Default() *Config { return DefaultConfig() }

// Load loads the Go packages matching the given patterns and returns a Bast representation.
//
// cfg configures the loading process; if nil, DefaultConfig is used.
//
// Patterns follow the same syntax as go list or go build (e.g., "./...", "github.com/user/repo").
func Load(config *Config, patterns ...string) (bast *Bast, err error) {

	if config == nil {
		config = DefaultConfig()
	}

	var mode = packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName
	if config.TypeChecking {
		mode |= packages.NeedTypes | packages.NeedDeps | packages.NeedImports
	}

	var (
		cfg = &packages.Config{
			Mode:       mode,
			Dir:        config.Dir,
			BuildFlags: config.BuildFlags,
			Env:        config.Env,
			Tests:      config.Tests,
		}
		pkgs []*packages.Package
	)

	if pkgs, err = packages.Load(cfg, patterns...); err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	return NewParser(config).Parse(pkgs)
}

// Parser transforms loaded go/packages into the bast model.
type Parser struct {
	config *Config
	fset   *token.FileSet
	p      *printer.Config
}

// NewParser creates a new Parser with the given config.
func NewParser(config *Config) *Parser {
	var p = &Parser{
		config: config,
		fset:   token.NewFileSet(),
	}
	p.p = &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	return p
}

// Parse parses the given loaded packages into a Bast.
//
// It returns an error if parsing fails or if TypeCheckingErrors is true and any package has errors.
func (self *Parser) Parse(pkgs []*packages.Package) (*Bast, error) {
	var bast = new()
	bast.config = self.config
	bast.fset = self.fset

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 && self.config.TypeCheckingErrors {
			var errs = make([]error, 0, len(pkg.Errors))
			for _, err := range pkg.Errors {
				errs = append(errs, err)
			}
			return nil, errors.Join(errs...)
		}
		bastPkg, err := self.parsePackage(pkg)
		if err != nil {
			return nil, err
		}
		bastPkg.bast = bast
		bast.packages.Put(bastPkg.Path, bastPkg)
	}
	return bast, nil
}

// parsePackage parses a package into a bast package, adds it to PackageMap
// keying it by its package path.
func (self *Parser) parsePackage(in *packages.Package) (*Package, error) {
	var pkg = NewPackage(in.Name, in.PkgPath, in)
	for idx, file := range in.Syntax {
		if err := self.parseFile(pkg, in.CompiledGoFiles[idx], file, pkg.Files); err != nil {
			return nil, err
		}
	}
	return pkg, nil
}

// parseFile parses an ast file parsed from fileName into a bast [File] and
// adds it to [FileMap], keyed by filename.
func (self *Parser) parseFile(pkg *Package, fileName string, in *ast.File, out *FileMap) error {

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

	return nil
}

// parseDeclaration parses in node into a DeclarationMap out.
func (self *Parser) parseDeclaration(file *File, in ast.Node, out *DeclarationMap) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.VAR:
			self.parseVars(file, n, out)
		case token.CONST:
			self.parseConsts(file, n, out)
		case token.TYPE:
			for _, spec := range n.Specs {
				var tspec, ok = spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				switch tspec.Type.(type) {
				case *ast.StructType:
					self.parseStruct(file, n, tspec, out)
				case *ast.InterfaceType:
					self.parseInterface(file, n, tspec, out)
				case *ast.FuncType:
					self.parseFuncType(file, n, tspec, out)
				default:
					self.parseType(file, n, tspec, out)
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
}

// parseCommentGroup a comment group into a string slice, line per entry.
func (self *Parser) parseCommentGroup(in *ast.CommentGroup, out *[]string) {
	if in == nil {
		return
	}
	for _, entry := range in.List {
		*out = append(*out, entry.Text)
	}
}

// parseImportSpec parses import spec into a map keyed by path.
func (self *Parser) parseImportSpec(in *ast.ImportSpec, out *ImportSpecMap) {
	var val = NewImport(
		self.printExpr(in.Name),
		"",
	)
	val.Path, _ = strutils.UnquoteDouble(self.printExpr(in.Path))
	self.parseCommentGroup(in.Doc, &val.Doc)
	out.Put(val.Path, val)
}

// parseVars parses a GenDecl in of vars into a DeclarationMap out.
func (self *Parser) parseVars(file *File, in *ast.GenDecl, out *DeclarationMap) {

	for _, spec := range in.Specs {

		var vspec, ok = spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i := 0; i < len(vspec.Names); i++ {
			var val = NewVar(file, self.printExpr(vspec.Names[i]), "")
			self.parseCommentGroup(vspec.Doc, &val.Doc)
			if vspec.Type != nil {
				val.Type = self.printExpr(vspec.Type)
			}
			if len(vspec.Values) > 0 && i < len(vspec.Values) {
				val.Value = self.printExpr(vspec.Values[i])
			}
			out.Put(val.Name, val)
		}
	}
}

// parseVars parses a GenDecl in of consts into a DeclarationMap out.
func (self *Parser) parseConsts(file *File, in *ast.GenDecl, out *DeclarationMap) {
	for _, spec := range in.Specs {

		var vspec, ok = spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i := 0; i < len(vspec.Names); i++ {
			var val = NewConst(file, self.printExpr(vspec.Names[i]), "")
			self.parseCommentGroup(vspec.Doc, &val.Doc)
			if vspec.Type != nil {
				val.Type = self.printExpr(vspec.Type)
			}
			if len(vspec.Values) > 0 && i < len(vspec.Values) {
				val.Value = self.printExpr(vspec.Values[i])
			}
			out.Put(val.Name, val)
		}
	}
}

// parseFunc parses in func decl into DeclarationMap out.
func (self *Parser) parseFunc(file *File, in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewFunc(file, self.printExpr(in.Name))
	self.parseCommentGroup(in.Doc, &val.Doc)
	self.parseFieldList(file, in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(file, in.Type.Params, val.Params)
	self.parseFieldList(file, in.Type.Results, val.Results)
	out.Put(val.Name, val)
}

// parseMethod parses in method decl into DeclarationMap out.
func (self *Parser) parseMethod(file *File, in *ast.FuncDecl, out *DeclarationMap) {
	var val = NewMethod(file, self.printExpr(in.Name))
	self.parseCommentGroup(in.Doc, &val.Doc)

	if in.Recv != nil {
		val.Receiver = NewField(file, "")
		if len(in.Recv.List[0].Names) > 0 {
			val.Receiver.Name = self.printExpr(in.Recv.List[0].Names[0])
		}
		// Parse out the bare receiver type name. Exclude star and type params.
		var expr = in.Recv.List[0].Type
		if star, ok := expr.(*ast.StarExpr); ok {
			val.Receiver.Pointer = true
			expr = star.X
		}
		if index, ok := expr.(*ast.IndexExpr); ok {
			expr = index.X
		}
		if index, ok := expr.(*ast.IndexListExpr); ok {
			expr = index.X
		}
		val.Receiver.Type = self.printExpr(expr)
		// val.Receiver.Type = self.printExpr(in.Recv.List[0].Type)
	}

	self.parseFieldList(file, in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(file, in.Type.Params, val.Params)
	self.parseFieldList(file, in.Type.Results, val.Results)
	out.Put(val.Name, val)
}

// parseFuncType parses func type spec in into a DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Parser) parseFuncType(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
	var val = NewFunc(file, self.printExpr(in.Name))
	self.parseCommentGroup(g.Doc, &val.Doc)
	var ft = in.Type.(*ast.FuncType)
	self.parseFieldList(file, in.TypeParams, val.TypeParams)
	self.parseFieldList(file, ft.Params, val.Params)
	self.parseFieldList(file, ft.Results, val.Results)
	out.Put(val.Name, val)
}

// parseType parses type spec in into a DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Parser) parseType(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {
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
}

// parseFieldList parses in field list into FieldMap out.
func (self *Parser) parseFieldList(file *File, in *ast.FieldList, out *FieldMap) {

	if in == nil {
		return
	}

	for idx, field := range in.List {
		var val = NewField(file, "")
		val.Type = self.printExpr(field.Type)
		if len(field.Names) > 0 {
			val.Name = self.printExpr(field.Names[0])
		} else {
			val.Name = fmt.Sprintf("unnamed%d", idx)
		}
		self.parseCommentGroup(field.Doc, &val.Doc)
		out.Put(val.Name, val)
	}
}

// parseStruct parses a struct declaration in into DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Parser) parseStruct(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {

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
}

// parseStructField parses a struct field in into a FieldMap out.
func (self *Parser) parseStructField(file *File, in *ast.Field, out *FieldMap) {

	var val = NewField(file, "")
	self.parseCommentGroup(in.Doc, &val.Doc)
	val.Type = self.printExpr(in.Type)
	if in.Tag != nil {
		val.Tag, _ = strutils.UnquoteDouble(in.Tag.Value)
	}

	// Unnamed/Embedded field.
	if len(in.Names) == 0 {
		val.Unnamed = true
		val.Name = val.Type
		out.Put(val.Name, val)
		return
	}

	// Named fields.
	for _, name := range in.Names {
		var f = val.Clone()
		f.Name = self.printExpr(name)
		out.Put(f.Name, f)
	}
}

// parseStruct parses an interface declaration in into DeclarationMap out.
// Uses parent GenDecl g docs as doc source.
func (self *Parser) parseInterface(file *File, g *ast.GenDecl, in *ast.TypeSpec, out *DeclarationMap) {

	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	var val = NewInterface(file, self.printExpr(in.Name))
	self.parseCommentGroup(g.Doc, &val.Doc)
	self.parseCommentGroup(in.Doc, &val.Doc)

	for _, method := range it.Methods.List {
		switch m := method.Type.(type) {
		case *ast.FuncType:
			var meth = NewMethod(file, self.printExpr(method.Names[0]))
			self.parseCommentGroup(method.Doc, &meth.Doc)
			self.parseFieldList(file, m.Params, meth.Params)
			self.parseFieldList(file, m.Results, meth.Results)
			val.Methods.Put(meth.Name, meth)
		default:
			// Embedded interface.
			var intf = NewInterface(file, self.printExpr(method.Type))
			self.parseCommentGroup(method.Doc, &intf.Doc)
			val.Interfaces.Put(intf.Name, intf)
		}
	}

	self.parseFieldList(file, in.TypeParams, val.TypeParams)

	out.Put(val.Name, val)
}

// printExpr prints an ast.Node.
func (self *Parser) printExpr(in any) (s string) {
	if in == nil || reflect.ValueOf(in).IsNil() {
		return ""
	}
	var buf = bytes.Buffer{}
	self.p.Fprint(&buf, self.fset, in)
	return buf.String()
}
