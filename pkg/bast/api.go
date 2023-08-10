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
			Mode: packages.NeedSyntax | packages.NeedCompiledGoFiles,
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
		"Vars":       self.Vars,
		"Consts":     self.Consts,
		"Types":      self.Types,
		"Funcs":      self.Funcs,
		"Methods":    self.Methods,
		"Interfaces": self.Interfaces,
		"Structs":    self.Structs,
	}
}

// Vars returns all variables in self, across all packages.
func (self *Bast) Vars() (out []Declaration) {
	return all[*Var](self.Packages)
}

// Vars returns all variables in self, across all packages.
func (self *Bast) Consts() (out []Declaration) {
	return all[*Const](self.Packages)
}

// Types returns all types in self, across all packages.
func (self *Bast) Types() (out []Declaration) {
	return all[*Type](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Funcs() (out []Declaration) {
	return all[*Func](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Methods() (out []Declaration) {
	return all[*Method](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Interfaces() (out []Declaration) {
	return all[*Interface](self.Packages)
}

// Funcs returns all functions in self, across all packages.
func (self *Bast) Structs() (out []Declaration) {
	return all[*Struct](self.Packages)
}

// -----------------------------------------------------------------------------

func all[T Declaration](p []*Package) (out []Declaration) {
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
