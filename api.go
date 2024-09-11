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
	"go/types"
	"reflect"
	"strings"
)

// PackageNames returns names of all parsed packages.
func (self *Bast) PackageNames() (out []string) {
	out = make([]string, 0, self.Packages.Len())
	for _, v := range self.Packages.Values() {
		out = append(out, v.Name)
	}
	return
}

// ResolveBasicType returns the basic type name of a derived type typeName by
// searching the type hierarchy of parsed packages.
//
// If typeName is already a name of a basic type it is returned as is.
// If basic type was not found resolved returns an empty string.
//
// [Config.TypeChecking] must be enabled for this method to work.
func (self *Bast) ResolveBasicType(typeName string) string {

	var o types.Object
	for _, p := range self.Packages.Values() {
		if o = p.pkg.Types.Scope().Lookup(typeName); o != nil {
			break
		}
	}

	if o == nil {
		switch tn := typeName; tn {
		case "bool", "byte",
			"int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"complex64", "complex128", "string":
			return tn
		default:
			return ""
		}
	}
	var t types.Type = o.Type()
	for {
		if t.Underlying() == nil {
			return t.String()
		}
		if t.Underlying() == t {
			break
		}
		t = t.Underlying()
	}

	return t.String()
}

// TypeImports returns imports needed for declaration.
func (self *Bast) TypeImports(typeName string) (imports []string) {

	var o types.Object
	for _, p := range self.Packages.Values() {
		if o = p.pkg.Types.Scope().Lookup(typeName); o != nil {
			break
		}
	}

	if o == nil {
		return nil
	}

	var t types.Type = o.Type()
	for {
		var p = o.Pkg()
		imports = append(imports, p.Path())
		if t.Underlying() == nil {
			return
		}
		if t.Underlying() == t {
			break
		}
		t = t.Underlying()
	}

	return
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
	if pkg, ok = self.Packages.Get(pkgName); !ok {
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

	for _, pkg := range self.Packages.Values() {

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

// PkgVar returns a variable whose Name==declName from a package named pkgName.
func (self *Bast) PkgVar(pkgName, declName string) (out *Var) {
	return pkgDecl[*Var](pkgName, declName, self.Packages)
}

// PkgConst returns a const whose Name==declName from a package named pkgName.
func (self *Bast) PkgConst(pkgName, declName string) (out *Const) {
	return pkgDecl[*Const](pkgName, declName, self.Packages)
}

// PkgType returns a type whose Name==declName from a package named pkgName.
func (self *Bast) PkgType(pkgName, declName string) (out *Type) {
	return pkgDecl[*Type](pkgName, declName, self.Packages)
}

// PkgFunc returns a func whose Name==declName from a package named pkgName.
func (self *Bast) PkgFunc(pkgName, declName string) (out *Func) {
	return pkgDecl[*Func](pkgName, declName, self.Packages)
}

// PkgMethod returns a method whose Name==declName from a package named pkgName.
func (self *Bast) PkgMethod(pkgName, declName string) (out *Method) {
	return pkgDecl[*Method](pkgName, declName, self.Packages)
}

// PkgInterface returns an interface whose Name==declName from a package named pkgName.
func (self *Bast) PkgInterface(pkgName, declName string) (out *Interface) {
	return pkgDecl[*Interface](pkgName, declName, self.Packages)
}

// PkgStruct returns a struct whose Name==declName from a package named pkgName.
func (self *Bast) PkgStruct(pkgName, declName string) (out *Struct) {
	return pkgDecl[*Struct](pkgName, declName, self.Packages)
}

// PkgVars returns all variables in pkgName.
func (self *Bast) PkgVars(pkgName string) (out []*Var) {
	return pkgDecls[*Var](pkgName, self.Packages)
}

// PgkConsts returns all consts in pkgName.
func (self *Bast) PkgConsts(pkgName string) (out []*Const) {
	return pkgDecls[*Const](pkgName, self.Packages)
}

// PkgTypes returns all types in pkgName.
func (self *Bast) PkgTypes(pkgName string) (out []*Type) {
	return pkgDecls[*Type](pkgName, self.Packages)
}

// PkgFuncs returns all functions in pkgName.
func (self *Bast) PkgFuncs(pkgName string) (out []*Func) {
	return pkgDecls[*Func](pkgName, self.Packages)
}

// PkgMethods returns all methods in pkgName.
func (self *Bast) PkgMethods(pkgName string) (out []*Method) {
	return pkgDecls[*Method](pkgName, self.Packages)
}

// PkgInterfaces returns all interfaces in pkgName.
func (self *Bast) PkgInterfaces(pkgName string) (out []*Interface) {
	return pkgDecls[*Interface](pkgName, self.Packages)
}

// PkgStructs returns all structs in pkgName.
func (self *Bast) PkgStructs(pkgName string) (out []*Struct) {
	return pkgDecls[*Struct](pkgName, self.Packages)
}

// AllPackages returns all parsed packages.
func (self *Bast) AllPackages() (out []*Package) {
	out = make([]*Package, 0, self.Packages.Len())
	for _, p := range self.Packages.Values() {
		out = append(out, p)
	}
	return
}

// AllVars returns all variables in self, across all packages.
func (self *Bast) AllVars() (out []*Var) {
	return allDecls[*Var](self.Packages)
}

// AllConsts returns all consts in self, across all packages.
func (self *Bast) AllConsts() (out []*Const) {
	return allDecls[*Const](self.Packages)
}

// AllTypes returns all types in self, across all packages.
func (self *Bast) AllTypes() (out []*Type) {
	return allDecls[*Type](self.Packages)
}

// AllFuncs returns all functions in self, across all packages.
func (self *Bast) AllFuncs() (out []*Func) {
	return allDecls[*Func](self.Packages)
}

// Funcs returns all methods in self, across all packages.
func (self *Bast) AllMethods() (out []*Method) {
	return allDecls[*Method](self.Packages)
}

// Funcs returns all interfaces in self, across all packages.
func (self *Bast) AllInterfaces() (out []*Interface) {
	return allDecls[*Interface](self.Packages)
}

// Funcs returns all structs in self, across all packages.
func (self *Bast) AllStructs() (out []*Struct) {
	return allDecls[*Struct](self.Packages)
}

// pkgTypeDecl returns all declarations of model T and type typeName from a
// package in p named pkgName.
func pkgTypeDecl[T Declaration](pkgName, typeName string, p *PackageMap) (out []T) {
	for _, pkg := range p.Values() {
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

// pkgDecl returns a declaration named declName of model T from a package
// in p named pkgName.
func pkgDecl[T Declaration](pkgName, declName string, p *PackageMap) (out T) {
	for _, pkg := range p.Values() {
		if pkg.Name != pkgName {
			continue
		}
		for _, file := range pkg.Files.Values() {
			if decl, ok := file.Declarations.Get(declName); ok {
				out, _ = decl.(T)
				return
			}
		}
	}
	return
}

// anyDecl returns a declaration named declName of model T from any package.
func anyDecl[T Declaration](declName string, p *PackageMap) (out T) {
	for _, pkg := range p.Values() {
		for _, file := range pkg.Files.Values() {
			if decl, ok := file.Declarations.Get(declName); ok {
				out, _ = decl.(T)
			}
			return
		}
	}
	return
}

// pkgDecls returns all declarations of model T from a package in p named pkgName.
func pkgDecls[T Declaration](pkgName string, p *PackageMap) (out []T) {
	for _, pkg := range p.Values() {
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

// allDecls returns all declarations of type T from all packages p.
func allDecls[T Declaration](p *PackageMap) (out []T) {
	for _, pkg := range p.Values() {
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
func (self *Bast) printExpr(in any) (s string) {
	if in == nil || reflect.ValueOf(in).IsNil() {
		return ""
	}
	var buf = bytes.Buffer{}
	self.p.Fprint(&buf, self.fset, in)
	return buf.String()
}
