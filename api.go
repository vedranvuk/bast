// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast provides a simple intermediate representation of top-level Go
// declarations parsed from source files using the go/ast package. It is designed
// for use in text-based code generation with Go's text/template package.
//
// The Bast type holds the parsed packages, files, and declarations. Its structure
// can be easily traversed in templates, and it provides utility methods for
// retrieving and inspecting elements.
package bast

import (
	"go/types"
	"strings"
)

// PackageNames returns the names of all parsed packages.
func (self *Bast) PackageNames() (out []string) {
	out = make([]string, 0, self.packages.Len())
	for _, v := range self.packages.Values() {
		out = append(out, v.Name)
	}
	return
}

// Packages returns the parsed packages.
func (self *Bast) Packages() []*Package { return self.packages.Values() }

// PackageImportPaths returns the import paths of all loaded packages.
func (self *Bast) PackageImportPaths() []string { return self.packages.Keys() }

// PackageByPath returns the package with the given import path, or nil if not found.
func (self *Bast) PackageByPath(pkgPath string) (p *Package) {
	var exists bool
	if p, exists = self.packages.Get(pkgPath); !exists {
		return nil
	}
	return
}

// ResolveBasicType resolves the underlying basic type name for the given typeName
// by searching the type hierarchy of the parsed packages.
//
// If typeName is already a basic type name, it returns typeName as is.
// If no basic type is found, it returns an empty string.
//
// This method requires Config.TypeChecking to be enabled.
func (self *Bast) ResolveBasicType(typeName string) string {

	// Handle qualified names (e.g., "pkg.Type")
	if pkg, name, hasSelector := strings.Cut(typeName, "."); hasSelector {
		// Try to resolve the qualified type by finding a package that imports it
		for _, p := range self.packages.Values() {
			for _, file := range p.Files.Values() {
				for _, imp := range file.Imports.Values() {
					// Check if this import matches the package selector
					var importMatches bool
					if imp.Name != "" {
						// Aliased import - check alias name
						importMatches = imp.Name == pkg
					} else {
						// Direct import - check base package name
						importMatches = imp.Base() == pkg
					}
					
					if importMatches {
						// Found matching import, now look for the type in that package
						if targetPkg, ok := self.packages.Get(imp.Path); ok {
							if o := targetPkg.pkg.Types.Scope().Lookup(name); o != nil {
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
						}
					}
				}
			}
		}
		return ""
	}

	// Handle unqualified names
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
			"float32", "float64",
			"complex64", "complex128", "string":
			return tn
		case "[]string":
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

// VarsOfType returns all top-level variable declarations from the package with path pkgPath
// whose type matches typeName.
func (self *Bast) VarsOfType(pkgPath, typeName string) (out []*Var) {
	return pkgTypeDecl[*Var](pkgPath, typeName, self.packages)
}

// ConstsOfType returns all top-level constant declarations from the package with path pkgPath
// whose type matches typeName.
func (self *Bast) ConstsOfType(pkgPath, typeName string) (out []*Const) {
	return pkgTypeDecl[*Const](pkgPath, typeName, self.packages)
}

// TypesOfType returns all top-level type declarations from the package with path pkgPath
// whose underlying type matches typeName.
func (self *Bast) TypesOfType(pkgPath, typeName string) (out []*Type) {
	return pkgTypeDecl[*Type](pkgPath, typeName, self.packages)
}

// MethodSet returns all methods from the package with path pkgPath whose receiver
// type matches typeName (with or without a pointer prefix).
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

// FieldNames returns the names of the fields of the struct named structName
// in the package with path pkgPath.
func (self *Bast) FieldNames(pkgPath, structName string) (out []string) {

	var pkg, ok = self.packages.Get(pkgPath)
	if !ok {
		return
	}

	for _, file := range pkg.Files.Values() {
		for _, decl := range file.Declarations.Values() {
			if v, ok := decl.(*Struct); ok && v.Name == structName {
				for _, field := range v.Fields.Values() {
					out = append(out, field.Name)
				}
			}
		}
	}

	return
}

// AnyVar returns the variable named declName from any parsed package, or nil if not found.
func (self *Bast) AnyVar(declName string) (out *Var) {
	return anyDecl[*Var](declName, self.packages)
}

// AnyConst returns the constant named declName from any parsed package, or nil if not found.
func (self *Bast) AnyConst(declName string) (out *Const) {
	return anyDecl[*Const](declName, self.packages)
}

// AnyFunc returns the function named declName from any parsed package, or nil if not found.
func (self *Bast) AnyFunc(declName string) (out *Func) {
	return anyDecl[*Func](declName, self.packages)
}

// AnyMethod returns the method named declName from any parsed package, or nil if not found.
func (self *Bast) AnyMethod(declName string) (out *Method) {
	return anyDecl[*Method](declName, self.packages)
}

// AnyType returns the type named declName from any parsed package, or nil if not found.
func (self *Bast) AnyType(declName string) (out *Type) {
	return anyDecl[*Type](declName, self.packages)
}

// AnyStruct returns the struct named declName from any parsed package, or nil if not found.
func (self *Bast) AnyStruct(declName string) (out *Struct) {
	return anyDecl[*Struct](declName, self.packages)
}

// AnyInterface returns the interface named declName from any parsed package, or nil if not found.
func (self *Bast) AnyInterface(declName string) (out *Interface) {
	return anyDecl[*Interface](declName, self.packages)
}

// PkgVar returns the variable named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgVar(pkgPath, declName string) (out *Var) {
	return pkgDecl[*Var](pkgPath, declName, self.packages)
}

// PkgConst returns the constant named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgConst(pkgPath, declName string) (out *Const) {
	return pkgDecl[*Const](pkgPath, declName, self.packages)
}

// PkgFunc returns the function named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgFunc(pkgPath, declName string) (out *Func) {
	return pkgDecl[*Func](pkgPath, declName, self.packages)
}

// PkgMethod returns the method named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgMethod(pkgPath, declName string) (out *Method) {
	return pkgDecl[*Method](pkgPath, declName, self.packages)
}

// PkgType returns the type named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgType(pkgPath, declName string) (out *Type) {
	return pkgDecl[*Type](pkgPath, declName, self.packages)
}

// PkgStruct returns the struct named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgStruct(pkgPath, declName string) (out *Struct) {
	return pkgDecl[*Struct](pkgPath, declName, self.packages)
}

// PkgInterface returns the interface named declName from the package with path pkgPath, or nil if not found.
func (self *Bast) PkgInterface(pkgPath, declName string) (out *Interface) {
	return pkgDecl[*Interface](pkgPath, declName, self.packages)
}

// PkgVars returns all top-level variables in the package with path pkgPath.
func (self *Bast) PkgVars(pkgPath string) (out []*Var) {
	return pkgDecls[*Var](pkgPath, self.packages)
}

// PkgConsts returns all top-level constants in the package with path pkgPath.
func (self *Bast) PkgConsts(pkgPath string) (out []*Const) {
	return pkgDecls[*Const](pkgPath, self.packages)
}

// PkgFuncs returns all top-level functions in the package with path pkgPath.
func (self *Bast) PkgFuncs(pkgPath string) (out []*Func) {
	return pkgDecls[*Func](pkgPath, self.packages)
}

// PkgMethods returns all top-level methods in the package with path pkgPath.
func (self *Bast) PkgMethods(pkgPath string) (out []*Method) {
	return pkgDecls[*Method](pkgPath, self.packages)
}

// PkgTypes returns all top-level types in the package with path pkgPath.
func (self *Bast) PkgTypes(pkgPath string) (out []*Type) {
	return pkgDecls[*Type](pkgPath, self.packages)
}

// PkgStructs returns all top-level structs in the package with path pkgPath.
func (self *Bast) PkgStructs(pkgPath string) (out []*Struct) {
	return pkgDecls[*Struct](pkgPath, self.packages)
}

// PkgInterfaces returns all top-level interfaces in the package with path pkgPath.
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

// AllVars returns all top-level variables across all parsed packages.
func (self *Bast) AllVars() (out []*Var) {
	return allDecls[*Var](self.packages)
}

// AllConsts returns all top-level constants across all parsed packages.
func (self *Bast) AllConsts() (out []*Const) {
	return allDecls[*Const](self.packages)
}

// AllFuncs returns all top-level functions across all parsed packages.
func (self *Bast) AllFuncs() (out []*Func) {
	return allDecls[*Func](self.packages)
}

// AllMethods returns all top-level methods across all parsed packages.
func (self *Bast) AllMethods() (out []*Method) {
	return allDecls[*Method](self.packages)
}

// AllTypes returns all top-level types across all parsed packages.
func (self *Bast) AllTypes() (out []*Type) {
	return allDecls[*Type](self.packages)
}

// AllStructs returns all top-level structs across all parsed packages.
func (self *Bast) AllStructs() (out []*Struct) {
	return allDecls[*Struct](self.packages)
}

// AllInterfaces returns all top-level interfaces across all parsed packages.
func (self *Bast) AllInterfaces() (out []*Interface) {
	return allDecls[*Interface](self.packages)
}

// pkgTypeDecl returns all declarations of type T with the specified typeName
// from the specified package.
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

// pkgDecl returns a declaration of type T with the specified name from the
// specified package, or nil if not found.
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

// anyDecl returns the first declaration of type T with the specified name
// found in any package, or nil if not found.
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

// pkgDecls returns all declarations of type T from the specified package.
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

// allDecls returns all declarations of type T from all packages.
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
