// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast implements a simple intermediate object model of top level Go
// declarations from AST of go source files and is designed for use in text
// based code-generation with Go's text/template templating engine.
//
// Bast's structure can be easily traversed from a template and provides a
// number of functions to help with data retrieval and other utils.
package bast

import (
	"bytes"
	"errors"
	"fmt"
	"go/printer"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Bast is a top level struct that contains parsed go packages.
// It also implements all functions usable from a text/template.
type Bast struct {
	// config is the parser configuration.
	config *ParsePackagesConfig
	// fset is the fileset of the parsed package.
	fset *token.FileSet
	// Packages is a list of packages parsed into bast using Load().
	//
	// Files outside of a package given to Load will be placed in a package
	// with an empty name.
	Packages map[string]*Package
}

// ParsePackagesConfig configures ParsePackages.
type ParsePackagesConfig struct {

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

// DefaultParseConfig returns the default configuration.
func DefaultParseConfig() *ParsePackagesConfig {
	return &ParsePackagesConfig{
		Dir:          ".",
		TypeChecking: true,
	}
}

// ParsePackage loads packages specified by pattern and returns a *Bast of it
// or an error.
//
// An optional config configures what is parsed, paths, etc.
// See [ParsePackagesConfig].
func ParsePackages(config *ParsePackagesConfig, patterns ...string) (bast *Bast, err error) {

	bast = new()
	if bast.config = config; bast.config == nil {
		bast.config = DefaultParseConfig()
	}

	var mode = packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName
	if config.TypeChecking {
		mode |= packages.NeedTypes | packages.NeedDeps | packages.NeedImports
	}

	var (
		cfg = &packages.Config{
			Mode:       mode,
			Logf:       func(format string, args ...any) { fmt.Printf(format, args...) },
			Dir:        bast.config.Dir,
			BuildFlags: bast.config.BuildFlags,
			Env:        bast.config.Env,
			Tests:      bast.config.Tests,
		}
		pkgs []*packages.Package
	)

	if pkgs, err = packages.Load(cfg, patterns...); err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var errs []error
			for _, e := range pkg.Errors {
				errs = append(errs, e)
			}
			return nil, fmt.Errorf("parse error: %w", errors.Join(errs...))
		}
		bast.parsePackage(pkg, bast.Packages)
	}

	return
}

// new returns a new, empty *Bast.
func new() *Bast {
	return &Bast{
		fset:     token.NewFileSet(),
		Packages: make(map[string]*Package),
	}
}

// PackageNames returns names of all parsed packages.
func (self *Bast) PackageNames() (out []string) {
	out = make([]string, 0, len(self.Packages))
	for _, v := range self.Packages {
		out = append(out, v.Name)
	}
	return
}

// ResolveBasicType returns the basic type of a derived type under the
// specified name. It returns an empty string if the base type was not found.
func (self *Bast) ResolveBasicType(typeName string) string {

	var o types.Object
	for _, p := range self.Packages {
		if o = p.pkg.Types.Scope().Lookup(typeName); o != nil {
			break
		}
	}

	var t types.Type = o.Type()
	for {
		if t.Underlying() == t {
			break
		}
		t = t.Underlying()
	}

	return t.String()
}

// VarsOfType returns all top level variable declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) VarsOfType(pkgName, typeName string) (out []*Var) {
	return pkgTypeDecl[*Var](pkgName, typeName, self.Packages)
}

// ConstsOfType returns all top level constant declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) ConstsOfType(pkgName, typeName string) (out []*Const) {
	return pkgTypeDecl[*Const](pkgName, typeName, self.Packages)
}

// MethodSet returns all methods from a package named pkgName whose receiver
// type matches typeName (star prefixed or not).
func (self *Bast) MethodSet(pkgName, typeName string) (out []*Method) {
	var (
		pkg *Package
		ok  bool
	)
	if pkg, ok = self.Packages[pkgName]; !ok {
		return
	}
	for _, file := range pkg.Files.Values() {
		for _, decl := range file.Declarations.Values() {

			if v, ok := decl.(*Method); ok {
				if strings.TrimLeft(v.Receiver.Type, "*") == typeName {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// FieldNames returns a slice of names of fields of a struct named by
// structName residing in some file in package named pkgName.
func (self *Bast) FieldNames(pkgName, structName string) (out []string) {

	for _, pkg := range self.Packages {

		if pkg.Name != pkgName {
			continue
		}

		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				if v, ok := decl.(*Struct); ok {
					for _, field := range v.Fields.Values() {
						out = append(out, field.Name)
					}
				}
			}
		}
	}

	return
}

// Var returns a variable whose Name==declName from a package named pkgName.
func (self *Bast) Var(pkgName, declName string) (out *Var) {
	return decl[*Var](pkgName, declName, self.Packages)
}

// Const returns a const whose Name==declName from a package named pkgName.
func (self *Bast) Const(pkgName, declName string) (out *Const) {
	return decl[*Const](pkgName, declName, self.Packages)
}

// Type returns a type whose Name==declName from a package named pkgName.
func (self *Bast) Type(pkgName, declName string) (out *Type) {
	return decl[*Type](pkgName, declName, self.Packages)
}

// Func returns a func whose Name==declName from a package named pkgName.
func (self *Bast) Func(pkgName, declName string) (out *Func) {
	return decl[*Func](pkgName, declName, self.Packages)
}

// Method returns a method whose Name==declName from a package named pkgName.
func (self *Bast) Method(pkgName, declName string) (out *Method) {
	return decl[*Method](pkgName, declName, self.Packages)
}

// Interface returns a interface whose Name==declName from a package named pkgName.
func (self *Bast) Interface(pkgName, declName string) (out *Interface) {
	return decl[*Interface](pkgName, declName, self.Packages)
}

// Struct returns a struct whose Name==declName from a package named pkgName.
func (self *Bast) Struct(pkgName, declName string) (out *Struct) {
	return decl[*Struct](pkgName, declName, self.Packages)
}

// Var returns a variable whose Name==declName from any package.
func (self *Bast) AnyVar(declName string) (out *Var) {
	return anyDecl[*Var](declName, self.Packages)
}

// Const returns a const whose Name==declName from from any package.
func (self *Bast) AnyConst(declName string) (out *Const) {
	return anyDecl[*Const](declName, self.Packages)
}

// Type returns a type whose Name==declName from from any package.
func (self *Bast) AnyType(declName string) (out *Type) {
	return anyDecl[*Type](declName, self.Packages)
}

// Func returns a func whose Name==declName from from any package.
func (self *Bast) AnyFunc(declName string) (out *Func) {
	return anyDecl[*Func](declName, self.Packages)
}

// Method returns a method whose Name==declName from from any package.
func (self *Bast) AnyMethod(declName string) (out *Method) {
	return anyDecl[*Method](declName, self.Packages)
}

// Interface returns a interface whose Name==declName from from any package.
func (self *Bast) AnyInterface(declName string) (out *Interface) {
	return anyDecl[*Interface](declName, self.Packages)
}

// Struct returns a struct whose Name==declName from from any package.
func (self *Bast) AnyStruct(declName string) (out *Struct) {
	return anyDecl[*Struct](declName, self.Packages)
}

// PkgVars returns all variables in self, across all packages.
func (self *Bast) PkgVars(pkgName string) (out []*Var) {
	return pkgDecl[*Var](pkgName, self.Packages)
}

// PgkConsts returns all variables in self, across all packages.
func (self *Bast) PkgConsts(pkgName string) (out []*Const) {
	return pkgDecl[*Const](pkgName, self.Packages)
}

// PkgTypes returns all types in self, across all packages.
func (self *Bast) PkgTypes(pkgName string) (out []*Type) {
	return pkgDecl[*Type](pkgName, self.Packages)
}

// PkgFuncs returns all functions in self, across all packages.
func (self *Bast) PkgFuncs(pkgName string) (out []*Func) {
	return pkgDecl[*Func](pkgName, self.Packages)
}

// PkgMethods returns all functions in self, across all packages.
func (self *Bast) PkgMethods(pkgName string) (out []*Method) {
	return pkgDecl[*Method](pkgName, self.Packages)
}

// PkgInterfaces returns all functions in self, across all packages.
func (self *Bast) PkgInterfaces(pkgName string) (out []*Interface) {
	return pkgDecl[*Interface](pkgName, self.Packages)
}

// PkgStructs returns all functions in self, across all packages.
func (self *Bast) PkgStructs(pkgName string) (out []*Struct) {
	return pkgDecl[*Struct](pkgName, self.Packages)
}

// AllPackages returns all parsed packages.
func (self *Bast) AllPackages() (out []*Package) {
	out = make([]*Package, 0, len(self.Packages))
	for _, p := range self.Packages {
		out = append(out, p)
	}
	return
}

// AllVars returns all variables in self, across all packages.
func (self *Bast) AllVars() (out []*Var) {
	return allDecl[*Var](self.Packages)
}

// AllConsts returns all variables in self, across all packages.
func (self *Bast) AllConsts() (out []*Const) {
	return allDecl[*Const](self.Packages)
}

// AllTypes returns all types in self, across all packages.
func (self *Bast) AllTypes() (out []*Type) {
	return allDecl[*Type](self.Packages)
}

// AllFuncs returns all functions in self, across all packages.
func (self *Bast) AllFuncs() (out []*Func) {
	return allDecl[*Func](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllMethods() (out []*Method) {
	return allDecl[*Method](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllInterfaces() (out []*Interface) {
	return allDecl[*Interface](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllStructs() (out []*Struct) {
	return allDecl[*Struct](self.Packages)
}

// pkgTypeDecl returns all declarations of model T and type typeName from a
// package in p named pkgName.
func pkgTypeDecl[T Declaration](pkgName, typeName string, p map[string]*Package) (out []T) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				switch d := decl.(type) {
				case *Var:
					if d.Type != typeName {
						continue
					}
				case *Const:
					if d.Type != typeName {
						continue
					}
				case *Type:
					if d.Type != typeName {
						continue
					}
				case *Interface:
					if d.Name != typeName {
						continue
					}
				case *Struct:
					if d.Name != typeName {
						continue
					}
				}
				if v, ok := decl.(T); ok {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// decl returns a declaration named declName of model T from a package
// in p named pkgName.
func decl[T Declaration](pkgName, declName string, p map[string]*Package) (out T) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				if v, ok := decl.(T); ok {
					if v.GetName() == declName {
						return v
					}
				}
			}
		}
	}
	return
}

// anyDecl returns a declaration named declName of model T from any package.
func anyDecl[T Declaration](declName string, p map[string]*Package) (out T) {
	for _, pkg := range p {
		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				if v, ok := decl.(T); ok {
					if v.GetName() == declName {
						return v
					}
				}
			}
		}
	}
	return
}

// pkgDecl returns all declarations of model T from a package in p named pkgName.
func pkgDecl[T Declaration](pkgName string, p map[string]*Package) (out []T) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				if v, ok := decl.(T); ok {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// allDecl returns all declarations of type T from all packages p.
func allDecl[T Declaration](p map[string]*Package) (out []T) {
	for _, pkg := range p {
		for _, file := range pkg.Files.Values() {
			for _, decl := range file.Declarations.Values() {
				if v, ok := decl.(T); ok {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// printExpr prints an ast.Node.
func (self *Bast) printExpr(in any) string {
	if in == nil || reflect.ValueOf(in).IsNil() {
		return ""
	}
	var buf = bytes.Buffer{}
	printer.Fprint(&buf, self.fset, in)
	return buf.String()
}
