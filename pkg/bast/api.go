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
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

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
// Patterns which name file paths ar epared into an unnamed package in bast.
func Load(patterns ...string) (bast *Bast, err error) {
	var (
		config = &packages.Config{
			Mode: packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName,
		}
		pkgs []*packages.Package
	)
	if pkgs, err = packages.Load(config, patterns...); err != nil {
		return nil, fmt.Errorf("load error: %w", err)
	}
	bast = new(Bast)
	for idx, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package error: %w", pkg.Errors[0])
		}
		parsePackage(pkg.CompiledGoFiles[idx], pkg, &bast.Packages)
	}
	return
}

// Bast is a top level struct that contains parsed go packages.
// It also implements all functions usable from a text/template.
type Bast struct {
	// Packages is a list of packages parsed into bast using Load().
	//
	// Files outside of a package given to Load will be placed in a package
	// with an empty name.
	Packages []*Package
}

// FuncMap returns a funcmap for use with text/template templates.
func (self *Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		// String utils
		"trimpfx": strings.TrimPrefix,
		"trimsfx": strings.TrimSuffix,
		"split":   strings.Split,
		"join":    self._join,
		"repeat":  self._repeat,
		// Retrieval by package and name.
		"var":       self._var,
		"const":     self._const,
		"type":      self._type,
		"func":      self._func,
		"method":    self._method,
		"interface": self._interface,
		"struct":    self._struct,
		// Retrieval by package.
		"vars":       self._vars,
		"consts":     self._consts,
		"types":      self._types,
		"funcs":      self._funcs,
		"methods":    self._methods,
		"interfaces": self._interfaces,
		"structs":    self._structs,
		// Global retrieval by kind.
		"allvars":       self._allvars,
		"allconsts":     self._allconsts,
		"alltypes":      self._alltypes,
		"allfuncs":      self._allfuncs,
		"allmethods":    self._allmethods,
		"allinterfaces": self._allinterfaces,
		"allstructs":    self._allstructs,
		// Type retrieval
		"varsoftype":   self._varsOfType,
		"constsoftype": self._constsOfType,
		// Struct helpers.
		"structmethods":    self._structMethods,
		"structfieldnames": self._structFieldNames,
	}
}

// -----------------------------------------------------------------------------
// string utils
// -----------------------------------------------------------------------------

// _repeat repeats string s n times, separates it with sep and returns it.
func (self *Bast) _repeat(s, delim string, n int) string {
	var a []string
	for i := 0; i < n; i++ {
		a = append(a, s)
	}
	return strings.Join(a, delim)
}

// _join joins s with sep.
func (self *Bast) _join(sep string, s ...string) string { return strings.Join(s, sep) }

// -----------------------------------------------------------------------------
// Single declaration retrieval by kind and name.
// -----------------------------------------------------------------------------

// _var returns a variable whose Name==declName from a package named pkgName.
func (self *Bast) _var(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Var](pkgName, declName, self.Packages)
}

// Var returns a const whose Name==declName from a package named pkgName.
func (self *Bast) _const(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Const](pkgName, declName, self.Packages)
}

// Var returns a type whose Name==declName from a package named pkgName.
func (self *Bast) _type(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Type](pkgName, declName, self.Packages)
}

// Var returns a func whose Name==declName from a package named pkgName.
func (self *Bast) _func(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Func](pkgName, declName, self.Packages)
}

// Var returns a method whose Name==declName from a package named pkgName.
func (self *Bast) _method(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Method](pkgName, declName, self.Packages)
}

// Var returns a interface whose Name==declName from a package named pkgName.
func (self *Bast) _interface(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Interface](pkgName, declName, self.Packages)
}

// Var returns a struct whose Name==declName from a package named pkgName.
func (self *Bast) _struct(pkgName, declName string) (out Declaration) {
	return pkgNamedDecl[*Struct](pkgName, declName, self.Packages)
}

// -----------------------------------------------------------------------------
// All declaration retrieval by kind.
// -----------------------------------------------------------------------------

// _allvars returns all variables in self, across all packages.
func (self *Bast) _allvars() (out []Declaration) {
	return allDecl[*Var](self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) _allconsts() (out []Declaration) {
	return allDecl[*Const](self.Packages)
}

// _alltypes returns all types in self, across all packages.
func (self *Bast) _alltypes() (out []Declaration) {
	return allDecl[*Type](self.Packages)
}

// _allfuncs returns all functions in self, across all packages.
func (self *Bast) _allfuncs() (out []Declaration) {
	return allDecl[*Func](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _allmethods() (out []Declaration) {
	return allDecl[*Method](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _allinterfaces() (out []Declaration) {
	return allDecl[*Interface](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _allstructs() (out []Declaration) {
	return allDecl[*Struct](self.Packages)
}

// -----------------------------------------------------------------------------
// Package declaration retrieval by kind.
// -----------------------------------------------------------------------------

// Vars returns all variables in self, across all packages.
func (self *Bast) _vars(pkgName string) (out []Declaration) {
	return pkgDecl[*Var](pkgName, self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) _consts(pkgName string) (out []Declaration) {
	return pkgDecl[*Const](pkgName, self.Packages)
}

// Types returns all types in self, across all packages.
func (self *Bast) _types(pkgName string) (out []Declaration) {
	return pkgDecl[*Type](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _funcs(pkgName string) (out []Declaration) {
	return pkgDecl[*Func](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _methods(pkgName string) (out []Declaration) {
	return pkgDecl[*Method](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _interfaces(pkgName string) (out []Declaration) {
	return pkgDecl[*Interface](pkgName, self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) _structs(pkgName string) (out []Declaration) {
	return pkgDecl[*Struct](pkgName, self.Packages)
}

// -----------------------------------------------------------------------------
// Package declaration retrieval by type name.
// -----------------------------------------------------------------------------

// _varsOfType returns all top level variable declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) _varsOfType(pkgName, typeName string) (out []Declaration) {
	return pkgTypeDecl[*Var](pkgName, typeName, self.Packages)
}

// _constsOfType returns all top level constant declarations from a package named
// pkgName whose type name equals typeName.
func (self *Bast) _constsOfType(pkgName, typeName string) (out []Declaration) {
	return pkgTypeDecl[*Const](pkgName, typeName, self.Packages)
}

// -----------------------------------------------------------------------------
// Struct utils.
// -----------------------------------------------------------------------------

// _structMethods returns all methods from a package named pkgName whose receiver
// type matches structName (star prefixed or not).
func (self *Bast) _structMethods(pkgName, structName string) (out []Declaration) {
	for _, pkg := range self.Packages {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files {
			for _, decl := range file.Declarations {
				if v, ok := decl.(*Method); ok {
					for _, recv := range v.Receivers {
						if strings.TrimLeft(recv.Type, "*") == structName {
							out = append(out, v)
						}
					}
				}
			}
		}
	}
	return
}

// _structFieldNames returns a slice of names of fields of a struct named by
// structName residing in some file in package named pkgName. Optional prefix
// is prepended to each name.
func (self *Bast) _structFieldNames(pkgName, structName string) (out []string) {
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

// -----------------------------------------------------------------------------
// internals
// -----------------------------------------------------------------------------

func allDecl[T Declaration](p []*Package) (out []Declaration) {
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

func pkgDecl[T Declaration](pkgName string, p []*Package) (out []Declaration) {
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

func pkgNamedDecl[T Declaration](pkgName, declName string, p []*Package) (out Declaration) {
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

func pkgTypeDecl[T Declaration](pkgName, typeName string, p []*Package) (out []Declaration) {
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
