// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains bast object model types.

package bast

import (
	"go/printer"
	"go/token"
	"go/types"
	"path"
	"strconv"
	"strings"

	"github.com/vedranvuk/ds/maps"
	"github.com/vedranvuk/strutils"
	"golang.org/x/tools/go/packages"
)

// Bast holds lists of top level declarations found in a set of parsed packages.
//
// It is a reduced model of go source which parses out only top level
// declarations and provides a simple model and interface for their easy
// retrieval, enumeration and inspection.
//
// It is returned the by [Load] function.
type Bast struct {
	// packages maps bast Packages by their import path.
	packages *PackageMap
	// config is the parser configuration used to parse current declarations..
	config *Config
	// fset is the fileset of parsed packages.
	fset *token.FileSet
	// p prints nodes using ast/printer.
	p *printer.Config
}

// new returns a new, empty *Bast.
func new() *Bast {
	return &Bast{
		fset:     token.NewFileSet(),
		p:        &printer.Config{Tabwidth: 8},
		packages: maps.MakeOrderedMap[string, *Package](),
	}
}

// declarations is bast declarations typeset.
type declarations interface {
	*Var | *Const | *Func | *Method | *Type | *Struct | *Interface
}

// Package contains info about a Go package.
type Package struct {
	// Name is the package name, without path, as it appears in source code.
	Name string
	// Path is the package import path as used by go compiler.
	Path string
	// Files maps definitions of parsed go files by their full path.
	Files *FileMap
	// bast is a reference to top level Bast struct.
	bast *Bast
	// pkg is the parsed package.
	pkg *packages.Package
}

// Var returns a Var declaration from File under name or nil if not found.
func (self *Package) Var(name string) (out *Var) {
	return pkgDecl[*Var](self.Path, name, self.bast.packages)
}

// Var returns a Const declaration from File under name or nil if not found.
func (self *Package) Const(name string) (out *Const) {
	return pkgDecl[*Const](self.Path, name, self.bast.packages)
}

// Var returns a Func declaration from File under name or nil if not found.
func (self *Package) Func(name string) (out *Func) {
	return pkgDecl[*Func](self.Path, name, self.bast.packages)
}

// Var returns a Method declaration from File under name or nil if not found.
func (self *Package) Method(name string) (out *Method) {
	return pkgDecl[*Method](self.Path, name, self.bast.packages)
}

// Var returns a Type declaration from File under name or nil if not found.
func (self *Package) Type(name string) (out *Type) {
	return pkgDecl[*Type](self.Path, name, self.bast.packages)
}

// Var returns a Struct declaration from File under name or nil if not found.
func (self *Package) Struct(name string) (out *Struct) {
	return pkgDecl[*Struct](self.Path, name, self.bast.packages)
}

// Var returns a Interface declaration from File under name or nil if not found.
func (self *Package) Interface(name string) (out *Interface) {
	return pkgDecl[*Interface](self.Path, name, self.bast.packages)
}

// DeclFile returns the full filename of a file that contains the declaration
// whose type equals typeName. If not found result is an empty string.
func (self *Package) DeclFile(typeName string) string {
	for _, file := range self.Files.Values() {
		if _, ok := file.Declarations.Get(typeName); ok {
			return file.Name
		}
	}
	return ""
}

// HasDecl returns true if declaration with typeName was found in this package.
func (self *Package) HasDecl(typeName string) bool {
	return self.DeclFile(typeName) != ""
}

// PackageMap maps packages by their import path.
type PackageMap = maps.OrderedMap[string, *Package]

// File contians info about a Go source file.
type File struct {
	// Comments are the file comments, grouped by separation, including docs.
	Comments [][]string
	// Doc is the file doc comment.
	Doc []string
	// Name is the File name, a full file path.
	Name string
	// Imports is a list of file imports.
	Imports *ImportSpecMap
	// Declarations is a list of top level declarations in the file.
	Declarations *DeclarationMap
	// pkg is the parent *Package.
	pkg *Package
}

// Var returns a Var declaration from File under name or nil if not found.
func (self *File) Var(name string) (out *Var) { return fileDecl[*Var](name, self) }

// Var returns a Const declaration from File under name or nil if not found.
func (self *File) Const(name string) (out *Const) { return fileDecl[*Const](name, self) }

// Var returns a Func declaration from File under name or nil if not found.
func (self *File) Func(name string) (out *Func) { return fileDecl[*Func](name, self) }

// Var returns a Method declaration from File under name or nil if not found.
func (self *File) Method(name string) (out *Method) { return fileDecl[*Method](name, self) }

// Var returns a Type declaration from File under name or nil if not found.
func (self *File) Type(name string) (out *Type) { return fileDecl[*Type](name, self) }

// Var returns a Struct declaration from File under name or nil if not found.
func (self *File) Struct(name string) (out *Struct) { return fileDecl[*Struct](name, self) }

// Var returns a Interface declaration from File under name or nil if not found.
func (self *File) Interface(name string) (out *Interface) { return fileDecl[*Interface](name, self) }

// HasDecl returns true if a var, const, func or method declaration with name
// or type, struct or interface with type name was found in this file.
func (self *File) HasDecl(name string) (b bool) {
	_, b = self.Declarations.Get(name)
	return
}

// ImportSpecFromSelector returns an import spec from a selectorExpr.
// If import was not found or selectorExpr is invalid result is nil.
func (self *File) ImportSpecFromSelector(selectorExpr string) *ImportSpec {
	var pkg, _, selector = strings.Cut(selectorExpr, ".")
	if !selector {
		return nil
	}
	for _, imp := range self.Imports.Values() {
		if imp.Name != "" {
			if imp.Name == pkg {
				return imp
			}
		} else {
			if imp.Base() == pkg {
				return imp
			}
		}
	}
	return nil
}

// fileDecl returns a declaration named declName of model T from file.
func fileDecl[T declarations](declName string, file *File) (out T) {
	if decl, ok := file.Declarations.Get(declName); ok {
		out, _ = decl.(T)
	}
	return
}

// FileMap maps files by their FileName in parse order.
type FileMap = maps.OrderedMap[string, *File]

// ImportSpec contains info about an Package or File import.
type ImportSpec struct {
	// Doc is the import doc comment.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// Base returns the base name from the package path.
func (self *ImportSpec) Base() string { return path.Base(self.Path) }

// ImportSpecMap maps imports by their path in parse order.
type ImportSpecMap = maps.OrderedMap[string, *ImportSpec]

// Declaration represents a top level declaration in a Go file.
type Declaration interface {
	// GetFile returns the declarations parent file.
	GetFile() *File
	// GetPackage returns the declarations parent package.
	GetPackage() *Package
}

// DeclarationMap maps declarations by their name in parse order.
type DeclarationMap = maps.OrderedMap[string, Declaration]

// Model is the bast model base with fields shared by all declarations.
//
// Model implements [Declaration] interface].
type Model struct {

	// Doc is the declaration doc comment.
	Doc []string

	// Name is the declaration name.
	//
	// For [Struct], this will be the bare name of the struct type without type
	// parameters. Type parameters are stored separately in a [Struct]
	// definition.
	//
	// If struct field is unnamed Name will be equal to Type. 
	// [Field.Unnamed] will be set to true as well.
	Name string

	// file is the file where the declaration is parsed from.
	file *File
}

// GetFile returns the declarations parent file.
func (self *Model) GetFile() *File { return self.file }

// GetPackage returns the declarations parent package.
func (self *Model) GetPackage() *Package { return self.file.pkg }

// ImportSpecBySelectorExpr returns an ImportSpec whose path is the path of a
// package from which a type qualified by selectorExpr is imported into
// go file being parsed. I.e.: "package.TypeName".
//
// It returns nil if not found or selectorExpr is invalid.
func (self *Model) ImportSpecBySelectorExpr(selectorExpr string) *ImportSpec {

	var pkg, sel, ok = strings.Cut(selectorExpr, ".")
	if !ok || pkg == "" || sel == "" {
		return nil
	}

	for _, imp := range self.file.Imports.Values() {

		// Package is named.
		if imp.Name == pkg {
			return imp
		}

		// Trim major version suffix if present.
		var s, _ = strutils.UnquoteDouble(imp.Path)
		s = path.Base(s)
		if strings.HasPrefix(s, "v") {
			if _, err := strconv.Atoi(s[1:]); err == nil {
				s = path.Base(imp.Path[:len(imp.Path)-len(s)+1])
			}
		}

		// last import path element matches selector package.
		if s == pkg {
			return imp
		}

	}

	return nil
}

// ResolveBasicType returns the basic type name of a derived type typeName by
// searching the type hierarchy of parsed packages.
//
// If typeName is already a name of a basic type it is returned as is.
// If basic type was not found resolved returns an empty string.
//
// Requires [Config.TypeChecking].
func (self *Model) ResolveBasicType(typeName string) string {

	var o types.Object
	if _, name, selector := strings.Cut(typeName, "."); selector {
		if imp := self.ImportSpecBySelectorExpr(typeName); imp != nil {
			if pkg, ok := self.GetPackage().bast.packages.Get(imp.Path); ok {
				o = pkg.pkg.Types.Scope().Lookup(name)
			}
		}
	} else {
		o = self.GetPackage().pkg.Types.Scope().Lookup(typeName)
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

// Var contains info about a variable.
//
// If a variable was declared with implicit type, Type will be empty.
// If a variable was declared without an initial value, Value will be empty.
type Var struct {
	// Model is the declaration base.
	Model
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Const contains info about a constant.
//
// If a variable was declared with implicit type, Type will be empty.
type Const struct {
	// Model is the declaration base.
	Model
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if unspecified.
	Value string
}

// Field contains info about a struct field, method receiver, or method or func
// type params, params or results.
type Field struct {

	// Model is the declaration base.
	Model

	// Type is the field type.
	//
	// For [Struct] receivers, type will be the bare type name without star or
	// type parameters. If it is a pointer receiver [Field.Pointer] will be
	// true. Type parameters are ommited from the name and can be inspected in
	// parent [Struct] definition.
	Type string

	// Tag is the field raw tag string.
	Tag string

	// Unnamed is true if field is unnamed and specifies the type only.
	Unnamed bool

	// Pointer is true if this field is a pointer method receiver.
	Pointer bool
}

// FieldMap maps fields by their name in parse order.
type FieldMap = maps.OrderedMap[string, *Field]

// Func contains info about a function.
type Func struct {
	// Model is the declaration base.
	Model
	// TypeParams are type parameters.
	TypeParams *FieldMap
	//  Params is a list of func arguments.
	Params *FieldMap
	// Results is a list of func returns.
	Results *FieldMap
}

// Method contains info about a method.
type Method struct {
	// Func embeds all Func properties.
	Func
	// Receiver is the method receiver.
	//
	// Receiver is nil if this is an interface method without a receiver.
	Receiver *Field
}

// MethodMap maps methods by their name in parse order.
type MethodMap = maps.OrderedMap[string, *Method]

// Type contains info about a type.
type Type struct {
	// Model is the declaration base.
	Model
	// Type is Type's underlying type.
	//
	// The name can be a selector qualifying the package it originates in.
	Type string
	// IsAlias is true if type is an alias of the type it derives from.
	IsAlias bool
	// TypeParams are type parameters.
	TypeParams *FieldMap
}

// Struct contains info about a struct.
type Struct struct {
	// Model is the declaration base.
	Model
	// Fields is a list of struct fields.
	Fields *FieldMap
	// TypeParams are type parameters.
	TypeParams *FieldMap
}

// Methods returns a slice of methods defined on this struct.
func (self *Struct) Methods() (out []*Method) {
	for _, file := range self.GetPackage().Files.Values() {
		for _, decl := range file.Declarations.Values() {
			if m, ok := decl.(*Method); ok {
				if m.Receiver.Type == self.Name {
					out = append(out, m)
				}
			}
		}
	}
	return
}

// Interface contains info about an interface.
type Interface struct {
	// Model is the declaration base.
	Model
	// Methods is a list of methods defined by the interface.
	Methods *MethodMap
	// Interface is a list of inherited interfaces.
	//
	// Map is keyed by the embeded interface type name.
	Interfaces *FieldMap
}

// NewPackage returns a new *Package.
func NewPackage(name, path string, pkg *packages.Package) *Package {
	return &Package{
		Name:  name,
		Path:  path,
		Files: maps.MakeOrderedMap[string, *File](),
		pkg:   pkg,
	}
}

// NewFile returns a new *File.
func NewFile(pkg *Package, name string) *File {
	return &File{
		Name:         name,
		Imports:      maps.MakeOrderedMap[string, *ImportSpec](),
		Declarations: maps.MakeOrderedMap[string, Declaration](),
		pkg:          pkg,
	}
}

// NewImport returns a new *Import.
func NewImport(name, path string) *ImportSpec {
	return &ImportSpec{
		Name: name,
		Path: path,
	}
}

// NewFunc returns a new *Func.
func NewFunc(file *File, name string) *Func {
	return &Func{
		Model: Model{
			Name: name,
			file: file,
		},
		TypeParams: maps.MakeOrderedMap[string, *Field](),
		Params:     maps.MakeOrderedMap[string, *Field](),
		Results:    maps.MakeOrderedMap[string, *Field](),
	}
}

// NewMethod returns a new *Method.
func NewMethod(file *File, name string) *Method {
	return &Method{
		Func: *NewFunc(file, name),
	}
}

// NewConst returns a new *Const.
func NewConst(file *File, name, typ string) *Const {
	return &Const{
		Model: Model{
			Name: name,
			file: file,
		},
		Type: typ,
	}

}

// NewVar returns a new *Var.
func NewVar(file *File, name, typ string) *Var {
	return &Var{
		Model: Model{
			Name: name,
			file: file,
		},
		Type: typ,
	}
}

// NewType returns a new *Type.
func NewType(file *File, name, typ string) *Type {
	return &Type{
		Model: Model{
			Name: name,
			file: file,
		},
		Type:       typ,
		TypeParams: maps.MakeOrderedMap[string, *Field](),
	}
}

// NewStruct returns a new *Struct.
func NewStruct(file *File, name string) *Struct {
	return &Struct{
		Model: Model{
			Name: name,
			file: file,
		},
		Fields:     maps.MakeOrderedMap[string, *Field](),
		TypeParams: maps.MakeOrderedMap[string, *Field](),
	}

}

// NewField returns a new *Field.
func NewField(file *File, name string) *Field {
	return &Field{
		Model: Model{
			Name: name,
			file: file,
		},
	}
}

// NewInterface returns a new *Interface.
func NewInterface(file *File, name string) *Interface {
	return &Interface{
		Model: Model{
			Name: name,
			file: file,
		},
		Methods:    maps.MakeOrderedMap[string, *Method](),
		Interfaces: maps.MakeOrderedMap[string, *Field](),
	}
}
