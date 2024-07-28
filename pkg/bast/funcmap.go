package bast

import (
	"strings"
	"text/template"
	"time"
)

// FuncMap returns a funcmap for use with text/template templates.
func (self *Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		// String utils
		"trimpfx":   strings.TrimPrefix,
		"trimsfx":   strings.TrimSuffix,
		"lowercase": strings.ToLower,
		"uppercase": strings.ToUpper,
		"split":     strings.Split,
		"join":      self._join,
		"repeat":    self._repeat,
		// Other utils
		"datefmt":    self._datefmt,
		"dateutcfmt": self._dateutcfmt,
		// Retrieval utils
		"packagenames": self.PackageNames,
		"varsoftype":   self.VarsOfType,
		"constsoftype": self.ConstsOfType,
		"methodset":    self.MethodSet,
		"fieldnames":   self.FieldNames,
		// Get one by name from specific package.
		"var":       self.Var,
		"const":     self.Const,
		"type":      self.Type,
		"func":      self.Func,
		"method":    self.Method,
		"interface": self.Interface,
		"struct":    self.Struct,
		// Get all by kind from specific package.
		"pkgvars":       self.PkgVars,
		"pkgconsts":     self.PkgConsts,
		"pkgtypes":      self.PkgTypes,
		"pkgfuncs":      self.PkgFuncs,
		"pkgmethods":    self.PkgMethods,
		"pkginterfaces": self.PkgInterfaces,
		"pkgstructs":    self.PkgStructs,
		// Get all by kind from all packages.
		"allvars":       self.AllVars,
		"allconsts":     self.AllConsts,
		"alltypes":      self.AllTypes,
		"allfuncs":      self.AllFuncs,
		"allmethods":    self.AllMethods,
		"allinterfaces": self.AllInterfaces,
		"allstructs":    self.AllStructs,
	}
}

// _join joins s with sep.
func (self *Bast) _join(sep string, s ...string) string { return strings.Join(s, sep) }

// _repeat repeats string s n times, separates it with sep and returns it.
func (self *Bast) _repeat(s, delim string, n int) string {
	var a []string
	for i := 0; i < n; i++ {
		a = append(a, s)
	}
	return strings.Join(a, delim)
}

// _datefmt formats current time according to layout.
func (self *Bast) _datefmt(layout string) string { return time.Now().Format(layout) }

// _datefmt formats current time in UTC according to layout.
func (self *Bast) _dateutcfmt(layout string) string { return time.Now().UTC().Format(layout) }
