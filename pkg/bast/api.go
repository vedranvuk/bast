// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"text/template"
)

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
		"Vars":          self.Vars,
		"VarsForPkg":    self.VarsForPkg,
		"StructMethods": self.StructMethods,
	}
}

// Vars returns all variables in self, across all packages.
func (self *Bast) Vars() []*Var {
	return nil
}

func (self *Bast) VarsForPkg(pkgName string) []*Var {
	return nil
}

func (self *Bast) StructMethods(pkgName, structName string) []*Method {
	return nil
}
