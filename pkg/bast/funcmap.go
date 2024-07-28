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
		"varsoftype":   self._varsOfType,
		"constsoftype": self._constsOfType,
		"methodset":    self._methodset,
		"fieldnames":   self._fieldNames,
		// Get one by package and declaration name.
		"var":       self._var,
		"const":     self._const,
		"type":      self._type,
		"func":      self._func,
		"method":    self._method,
		"interface": self._interface,
		"struct":    self._struct,
		// Get all by package.
		"vars":       self._vars,
		"consts":     self._consts,
		"types":      self._types,
		"funcs":      self._funcs,
		"methods":    self._methods,
		"interfaces": self._interfaces,
		"structs":    self._structs,
		// Get all by kind.
		"allvars":       self._allvars,
		"allconsts":     self._allconsts,
		"alltypes":      self._alltypes,
		"allfuncs":      self._allfuncs,
		"allmethods":    self._allmethods,
		"allinterfaces": self._allinterfaces,
		"allstructs":    self._allstructs,
	}
}

func (self *Bast) _datefmt(layout string) string { return time.Now().Format(layout) }

func (self *Bast) _dateutcfmt(layout string) string { return time.Now().UTC().Format(layout) }

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

// _methodset returns all methods from a package named pkgName whose receiver
// type matches typeName (star prefixed or not).
func (self *Bast) _methodset(pkgName, typeName string) (out []Declaration) {
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

// _fieldNames returns a slice of names of fields of a struct named by
// structName residing in some file in package named pkgName.
func (self *Bast) _fieldNames(pkgName, structName string) (out []string) {
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
