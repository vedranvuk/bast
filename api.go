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
	out = make([]string, 0, self.packages.Len())
	for _, v := range self.packages.Values() {
		out = append(out, v.Name)
	}
	return
}

// Packages returns parsed bast packages.
func (self *Bast) Packages() []*Package { return self.packages.Values() }

// PackageImportPaths returns package paths of all loaded packages.
func (self *Bast) PackageImportPaths() []string { return self.packages.Keys() }

// PkgByImportPath returns a package by its import path or nil if not found.
func (self *Bast) PkgByImportPath(pkgPath string) (p *Package) {
	var exists bool
	if p, exists = self.packages.Get(pkgPath); !exists {
		return nil
	}
	return
}

// ResolveBasicType returns the basic type name of a derived type typeName by
// searching the type hierarchy of parsed packages.
//
// If typeName is already a name of a basic type it is returned as is.
// If basic type was not found resolved returns an empty string.
//
// Requires [Config.TypeChecking].
func (self *Bast) ResolveBasicType(typeName string) string {

	var o types.Object
	for _, p := range self.packages.Values() {
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

// VarsOfType returns all top level variable declarations from a package named
// pkgPath whose type name equals typeName.
func (self *Bast) VarsOfType(pkgPath, typeName string) (out []*Var) {
	return pkgTypeDecl[*Var](pkgPath, typeName, self.packages)
}

// ConstsOfType returns all top level constant declarations from a package named
// pkgPath whose type name equals typeName.
func (self *Bast) ConstsOfType(pkgPath, typeName string) (out []*Const) {
	return pkgTypeDecl[*Const](pkgPath, typeName, self.packages)
}

// TypesOfType returns all top level type declarations from a package named
// pkgPath whose type name equals typeName.
func (self *Bast) TypesOfType(pkgPath, typeName string) (out []*Const) {
	return pkgTypeDecl[*Const](pkgPath, typeName, self.packages)
}

// MethodSet returns all methods from a package named pkgPath whose receiver
// type matches typeName (star prefixed or not).
func (self *Bast) MethodSet(pkgPath, typeName string) (out []*Method) {
	var (
		pkg *Package
		ok  bool
	)
	if pkg, ok = self.packages.Get(pkgPath); !ok {
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
// structName residing in some file in package named pkgPath.
func (self *Bast) FieldNames(pkgPath, structName string) (out []string) {

	for _, pkg := range self.packages.Values() {

		if pkg.Name != pkgPath {
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
	return anyDecl[*Var](declName, self.packages)
}

// Const returns a const whose Name==declName from from any package.
func (self *Bast) AnyConst(declName string) (out *Const) {
	return anyDecl[*Const](declName, self.packages)
}

// Func returns a func whose Name==declName from from any package.
func (self *Bast) AnyFunc(declName string) (out *Func) {
	return anyDecl[*Func](declName, self.packages)
}

// Method returns a method whose Name==declName from from any package.
func (self *Bast) AnyMethod(declName string) (out *Method) {
	return anyDecl[*Method](declName, self.packages)
}

// Type returns a type whose Name==declName from from any package.
func (self *Bast) AnyType(declName string) (out *Type) {
	return anyDecl[*Type](declName, self.packages)
}

// Struct returns a struct whose Name==declName from from any package.
func (self *Bast) AnyStruct(declName string) (out *Struct) {
	return anyDecl[*Struct](declName, self.packages)
}

// Interface returns a interface whose Name==declName from from any package.
func (self *Bast) AnyInterface(declName string) (out *Interface) {
	return anyDecl[*Interface](declName, self.packages)
}

// PkgVar returns a variable whose Name==declName from a package named pkgPath.
func (self *Bast) PkgVar(pkgPath, declName string) (out *Var) {
	return pkgDecl[*Var](pkgPath, declName, self.packages)
}

// PkgConst returns a const whose Name==declName from a package named pkgPath.
func (self *Bast) PkgConst(pkgPath, declName string) (out *Const) {
	return pkgDecl[*Const](pkgPath, declName, self.packages)
}

// PkgFunc returns a func whose Name==declName from a package named pkgPath.
func (self *Bast) PkgFunc(pkgPath, declName string) (out *Func) {
	return pkgDecl[*Func](pkgPath, declName, self.packages)
}

// PkgMethod returns a method whose Name==declName from a package named pkgPath.
func (self *Bast) PkgMethod(pkgPath, declName string) (out *Method) {
	return pkgDecl[*Method](pkgPath, declName, self.packages)
}

// PkgType returns a type whose Name==declName from a package named pkgPath.
func (self *Bast) PkgType(pkgPath, declName string) (out *Type) {
	return pkgDecl[*Type](pkgPath, declName, self.packages)
}

// PkgStruct returns a struct whose Name==declName from a package named pkgPath.
func (self *Bast) PkgStruct(pkgPath, declName string) (out *Struct) {
	return pkgDecl[*Struct](pkgPath, declName, self.packages)
}

// PkgInterface returns an interface whose Name==declName from a package named pkgPath.
func (self *Bast) PkgInterface(pkgPath, declName string) (out *Interface) {
	return pkgDecl[*Interface](pkgPath, declName, self.packages)
}

// PkgVars returns all variables in pkgPath.
func (self *Bast) PkgVars(pkgPath string) (out []*Var) {
	return pkgDecls[*Var](pkgPath, self.packages)
}

// PgkConsts returns all consts in pkgPath.
func (self *Bast) PkgConsts(pkgPath string) (out []*Const) {
	return pkgDecls[*Const](pkgPath, self.packages)
}

// PkgFuncs returns all functions in pkgPath.
func (self *Bast) PkgFuncs(pkgPath string) (out []*Func) {
	return pkgDecls[*Func](pkgPath, self.packages)
}

// PkgMethods returns all methods in pkgPath.
func (self *Bast) PkgMethods(pkgPath string) (out []*Method) {
	return pkgDecls[*Method](pkgPath, self.packages)
}

// PkgTypes returns all types in pkgPath.
func (self *Bast) PkgTypes(pkgPath string) (out []*Type) {
	return pkgDecls[*Type](pkgPath, self.packages)
}

// PkgStructs returns all structs in pkgPath.
func (self *Bast) PkgStructs(pkgPath string) (out []*Struct) {
	return pkgDecls[*Struct](pkgPath, self.packages)
}

// PkgInterfaces returns all interfaces in pkgPath.
func (self *Bast) PkgInterfaces(pkgPath string) (out []*Interface) {
	return pkgDecls[*Interface](pkgPath, self.packages)
}

// AllPackages returns all parsed packages.
func (self *Bast) AllPackages() (out []*Package) {
	out = make([]*Package, 0, self.packages.Len())
	self.packages.EnumValues(func(p *Package) bool {
		out = append(out, p)
		return true
	})
	return
}

// AllVars returns all variables in self, across all packages.
func (self *Bast) AllVars() (out []*Var) {
	return allDecls[*Var](self.packages)
}

// AllConsts returns all consts in self, across all packages.
func (self *Bast) AllConsts() (out []*Const) {
	return allDecls[*Const](self.packages)
}

// AllFuncs returns all functions in self, across all packages.
func (self *Bast) AllFuncs() (out []*Func) {
	return allDecls[*Func](self.packages)
}

// Funcs returns all methods in self, across all packages.
func (self *Bast) AllMethods() (out []*Method) {
	return allDecls[*Method](self.packages)
}

// AllTypes returns all types in self, across all packages.
func (self *Bast) AllTypes() (out []*Type) {
	return allDecls[*Type](self.packages)
}

// Funcs returns all structs in self, across all packages.
func (self *Bast) AllStructs() (out []*Struct) {
	return allDecls[*Struct](self.packages)
}

// Funcs returns all interfaces in self, across all packages.
func (self *Bast) AllInterfaces() (out []*Interface) {
	return allDecls[*Interface](self.packages)
}

// pkgTypeDecl returns all declarations of model T and type typeName from a
// package in p witj pkgPath.
func pkgTypeDecl[T declarations](pkgPath, typeName string, p *PackageMap) (out []T) {

	var pkg, ok = p.Get(pkgPath)
	if !ok {
		return
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
	return
}

// pkgDecl returns a declaration named declName of model T from a package
// in p named pkgPath.
func pkgDecl[T declarations](pkgPath, declName string, p *PackageMap) (out T) {

	var pkg, ok = p.Get(pkgPath)
	if !ok {
		return
	}

	for _, file := range pkg.Files.Values() {
		if decl, ok := file.Declarations.Get(declName); ok {
			out, _ = decl.(T)
			return
		}
	}

	return
}

// anyDecl returns a declaration named declName of model T from any package.
func anyDecl[T declarations](declName string, p *PackageMap) (out T) {
	for _, pkg := range p.Values() {
		for _, file := range pkg.Files.Values() {
			if decl, ok := file.Declarations.Get(declName); ok {
				out, _ = decl.(T)
				return
			}
		}
	}
	return
}

// pkgDecls returns all declarations of model T from a package in p named pkgPath.
func pkgDecls[T declarations](pkgPath string, p *PackageMap) (out []T) {

	var pkg, ok = p.Get(pkgPath)
	if !ok {
		return
	}

	for _, file := range pkg.Files.Values() {
		for _, decl := range file.Declarations.Values() {
			if v, ok := decl.(T); ok {
				out = append(out, v)
			}
		}
	}

	return
}

// allDecls returns all declarations of type T from all packages p.
func allDecls[T declarations](p *PackageMap) (out []T) {
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
