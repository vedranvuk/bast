package bast

import (
	"go/ast"
	"text/template"
)

// Bast is a top level struct containign parsed go packages and/or files.
type Bast struct {
	pkgs     map[string]*ast.Package
	Packages []*Package
}

func (self *Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		"Vars": self.Vars,
	}
}

func (self *Bast) Vars() []*Var {
	return nil
}