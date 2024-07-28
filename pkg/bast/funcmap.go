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
		"varsoftype":   self.VarsOfType,
		"constsoftype": self.ConstsOfType,
		"methodset":    self.MethodSet,
		"fieldnames":   self.FieldNames,
		// Get one by package and declaration name.
		"var":       self.Var,
		"const":     self.Const,
		"type":      self.Type,
		"func":      self.Func,
		"method":    self.Method,
		"interface": self.Interface,
		"struct":    self.Struct,
		// Get all by package.
		"vars":       self.Vars,
		"consts":     self.Consts,
		"types":      self.Types,
		"funcs":      self.Funcs,
		"methods":    self.Methods,
		"interfaces": self.Interfaces,
		"structs":    self.Structs,
		// Get all by kind.
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

func (self *Bast) _datefmt(layout string) string { return time.Now().Format(layout) }

func (self *Bast) _dateutcfmt(layout string) string { return time.Now().UTC().Format(layout) }

