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
// Patterns which name file paths are parsed into an unnamed package in bast.
func Load(patterns ...string) (bast *Bast, err error) {
	var (
		config = &packages.Config{
			Mode: packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName,
			Logf: func(format string, args ...any) { fmt.Printf(format, args...) },
		}
		pkgs []*packages.Package
	)
	bast = New()
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
	bast = New()
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

// New returns a new, empty *Bast.
func New() *Bast {
	return &Bast{
		fset:     token.NewFileSet(),
		Packages: make(map[string]*Package),
	}
}

func (self *Bast) printExpr(in any) string {
	if in == nil || reflect.ValueOf(in).IsNil()  {
		return ""
	}
	var buf = bytes.Buffer{}
	printer.Fprint(&buf, self.fset, in)
	return buf.String()
}

