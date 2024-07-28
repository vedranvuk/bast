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
	"fmt"
	"go/printer"
	"go/token"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Bast is a top level struct that contains parsed go packages.
// It also implements all functions usable from a text/template.
type Bast struct {
	fset *token.FileSet
	// Packages is a list of packages parsed into bast using Load().
	//
	// Files outside of a package given to Load will be placed in a package
	// with an empty name.
	Packages map[string]*Package
}

// Load loads and returns bast of Go packages named by the given patterns.
//
// Paths can be absolute or relative paths to a directory containing go source
// files possibly defining a package or a cmd or a go file. It can also be a go
// module path loading of which depends on the current go environment.
//
// If an error in loading a file or a package occurs the error returned will be
// the first error of the first package that contains an error. Nil Bast is
// returned in case of an error.
//
// Patterns which name file paths are parsed into an unnamed package in bast.
func Load(patterns ...string) (bast *Bast, err error) {
	var (
		config = &packages.Config{
			Mode: packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName,
			Logf: func(format string, args ...any) { fmt.Printf(format, args...) },
		}
		pkgs []*packages.Package
	)
	bast = new()
	if len(patterns) == 0 {
		return
	}
	if pkgs, err = packages.Load(config, patterns...); err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package error: %w", pkg.Errors[0])
		}
		bast.parsePackage(pkg, bast.Packages)
	}
	return
}

// LoadPackage loads a single package.
func LoadPackage(dir string) (bast *Bast, err error) {
	var (
		config = &packages.Config{
			Mode: packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName,
			Logf: func(format string, args ...any) { fmt.Printf(format, args...) },
		}
		pkgs []*packages.Package
	)
	bast = new()
	bast.fset = token.NewFileSet()
	if pkgs, err = packages.Load(config, dir); err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package error: %w", pkg.Errors[0])
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

// Var returns a variable whose Name==declName from a package named pkgName.
func (self *Bast) Var(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Var](pkgName, declName, self.Packages)
}

// Var returns a const whose Name==declName from a package named pkgName.
func (self *Bast) Const(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Const](pkgName, declName, self.Packages)
}

// Var returns a type whose Name==declName from a package named pkgName.
func (self *Bast) Type(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Type](pkgName, declName, self.Packages)
}

// Var returns a func whose Name==declName from a package named pkgName.
func (self *Bast) Func(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Func](pkgName, declName, self.Packages)
}

// Var returns a method whose Name==declName from a package named pkgName.
func (self *Bast) Method(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Method](pkgName, declName, self.Packages)
}

// Var returns a interface whose Name==declName from a package named pkgName.
func (self *Bast) Interface(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Interface](pkgName, declName, self.Packages)
}

// Var returns a struct whose Name==declName from a package named pkgName.
func (self *Bast) Struct(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Struct](pkgName, declName, self.Packages)
}

// AllVarsž returns all variables in self, across all packages.
func (self *Bast) AllVars() (out []Declaration) {
	return allDecl[*Var](self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) AllConsts() (out []Declaration) {
	return allDecl[*Const](self.Packages)
}

// AllTypes returns all types in self, across all packages.
func (self *Bast) AllTypes() (out []Declaration) {
	return allDecl[*Type](self.Packages)
}

// AllFuncs returns all functions in self, across all packages.
func (self *Bast) AllFuncs() (out []Declaration) {
	return allDecl[*Func](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllMethods() (out []Declaration) {
	return allDecl[*Method](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllInterfaces() (out []Declaration) {
	return allDecl[*Interface](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) AllStructs() (out []Declaration) {
	return allDecl[*Struct](self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) Vars(pkgName string) (out []Declaration) {
	return pkgDecl[*Var](pkgName, self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) Consts(pkgName string) (out []Declaration) {
	return pkgDecl[*Const](pkgName, self.Packages)
}

// Types returns all types in self, across all packages.
func (self *Bast) Types(pkgName string) (out []Declaration) {
	return pkgDecl[*Type](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Funcs(pkgName string) (out []Declaration) {
	return pkgDecl[*Func](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Methods(pkgName string) (out []Declaration) {
	return pkgDecl[*Method](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Interfaces(pkgName string) (out []Declaration) {
	return pkgDecl[*Interface](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Structs(pkgName string) (out []Declaration) {
	return pkgDecl[*Struct](pkgName, self.Packages)
}

// VarsOfType returns all top level variable declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) VarsOfType(pkgName, typeName string) (out []Declaration) {
	return pkgTypeDecl[*Var](pkgName, typeName, self.Packages)
}

// ConstsOfType returns all top level constant declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) ConstsOfType(pkgName, typeName string) (out []Declaration) {
	return pkgTypeDecl[*Const](pkgName, typeName, self.Packages)
}

// MethodSet returns all methods from a package named pkgName whose receiver
// type matches typeName (star prefixed or not).
func (self *Bast) MethodSet(pkgName, typeName string) (out []Declaration) {
	var (
		pkg *Package
		ok  bool
	)
	if pkg, ok = self.Packages[pkgName]; !ok {
		return
	}
	for _, file := range pkg.Files {
		for _, decl := range file.Declarations {

			if v, ok := decl.(*Method); ok {
				for _, recv := range v.Receivers {
					if strings.TrimLeft(recv.Type, "*") == typeName {
						out = append(out, v)
					}
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
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
				if v, ok := decl.(*Struct); ok {
					for _, field := range v.Fields {
						out = append(out, field.Name)
					}
				}
			}
		}
	}
	return
}

func allDecl[T Declaration](p map[string]*Package) (out []Declaration) {
	for _, pkg := range p {
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
				if v, ok := decl.(T); ok {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// pkgDecl returns all declarations of model T from a package in p named pkgName.
func pkgDecl[T Declaration](pkgName string, p map[string]*Package) (out []Declaration) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
				if v, ok := decl.(T); ok {
					out = append(out, v)
				}
			}
		}
	}
	return
}

// pkgNamedDecl returns a declaration named declName of model T from a package
// in p named pkgName.
func pkgNamedDecl[T Declaration](pkgName, declName string, p map[string]*Package) (out Declaration) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
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

// pkgTypeDecl returns all declarations of model T and type typeName from a
// package in p named pkgName.
func pkgTypeDecl[T Declaration](pkgName, typeName string, p map[string]*Package) (out []Declaration) {
	for _, pkg := range p {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
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

func (self *Bast) printExpr(in any) string {
	if in == nil || reflect.ValueOf(in).IsNil() {
		return ""
	}
	var buf = bytes.Buffer{}
	printer.Fprint(&buf, self.fset, in)
	return buf.String()
}
