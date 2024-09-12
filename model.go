// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains bast object model types.

package bast

import (
	"go/printer"
	"go/token"
	"path"
	"strings"

	"github.com/vedranvuk/ds/maps"
	"golang.org/x/tools/go/packages"
)

// Bast is a top level struct that contains parsed go packages.
type Bast struct {
	// config is the parser configuration.
	config *Config
	// fset is the fileset of the parsed package.
	fset *token.FileSet
	// p is ast/printer for printing nodes.
	p *printer.Config
	// Packages is a list of packages parsed into bast using Load().
	//
	// Files outside of a package given to Load will be placed in a package
	// with an empty name.
	Packages *PackageMap
}

// new returns a new, empty *Bast.
func new() *Bast {
	return &Bast{
		fset:     token.NewFileSet(),
		p:        &printer.Config{Tabwidth: 8},
		Packages: maps.MakeOrderedMap[string, *Package](),
	}
}

// Declaration represents a top level declaration in a Go file.
type Declaration interface {
	// GetName returns the Declaration name.
	GetName() string
}

// DeclarationMap maps declarations by their name in parse order.
type DeclarationMap = maps.OrderedMap[string, Declaration]

// Package contains info about a Go package.
type Package struct {
	// Name is the package name, without path, as it appears in source code..
	Name string
	// Path is the package path as it appears in go import path.
	Path string
	// Files is a list of files in the package.
	Files *FileMap
	// pkg is the parsed package.
	pkg *packages.Package
}

// PackageMap maps packages by their import path.
type PackageMap = maps.OrderedMap[string, *Package]

// File contians info about a Go source file.
type File struct {
	// Comments are the file comments, grouped by separation, including docs.
	Comments [][]string
	// Doc is the file doc comment.
	Doc []string
	// Name is the File name, without path.
	Name string
	// Imports is a list of file imports.
	Imports *ImportSpecMap
	// Declarations is a list of top level declarations in the file.
	Declarations *DeclarationMap
}

// FileMap maps files by their name in parse order.
type FileMap = maps.OrderedMap[string, *File]

// ImportSpecForTypeSelector returns an ImportSpec for package that contains
// the type specified by typeName. Returns nil if not found.
//
// Requires [Config.TypeChecking].
func (self File) ImportSpecForTypeSelector(typeName string) *ImportSpec {
	var pkg, _, ok = strings.Cut(typeName, ".")
	if !ok {
		return nil
	}
	for _, v := range self.Imports.Values() {
		if path.Base(v.Path) == pkg {
			return v
		}
	}
	return nil
}

// ImportSpec contians info about an import.
type ImportSpec struct {
	// Doc is the import doc.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// ImportSpecMap maps imports by their name in parse order.
type ImportSpecMap = maps.OrderedMap[string, *ImportSpec]

// Func contains info about a function.
type Func struct {
	// Doc is the func doc comment.
	Doc []string
	// Name is the func name.
	Name string
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

// Const contains info about a constant.
type Const struct {
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Var contains info about a variable.
type Var struct {
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Type contains info about a type.
type Type struct {
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Type is Type's underlying type.
	// The name can be a selector qualifying the package it originates in.
	Type string
	// IsAlias is true if type is an alias of the type it derives from.
	IsAlias bool
}

// Interface contains info about an interface.
type Interface struct {
	// Doc is the interface doc comment.
	Doc []string
	// Name is the interface name.
	Name string
	// Methods is a list of methods defined by the interface.
	Methods *MethodMap
	// Interface is a list of inherited interfaces.
	//
	// Map is keyed by the embeded interface type name.
	Interfaces *FieldMap
}

// Struct contains info about a struct.
type Struct struct {
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Fields is a list of struct fields.
	Fields *FieldMap
}

// Field contains info about a struct field, method receiver, or method or func
// type params, params or results.
type Field struct {
	// Doc is the field doc comment.
	Doc []string
	// Name is the field name.
	Name string
	// Type is the field type.
	Type string
	// Tag is the field raw tag string.
	Tag string
	// Unnamed is true if field is unnamed and specifies the type only.
	Unnamed bool
}

// FieldMap maps fields by their name in parse order.
type FieldMap = maps.OrderedMap[string, *Field]

// NewPackage returns a new *Package.
func NewPackage() *Package {
	return &Package{
		Files: maps.MakeOrderedMap[string, *File](),
	}
}

// NewFile returns a new *File.
func NewFile() *File {
	return &File{
		Imports:      maps.MakeOrderedMap[string, *ImportSpec](),
		Declarations: maps.MakeOrderedMap[string, Declaration](),
	}
}

// NewImport returns a new *Import.
func NewImport() *ImportSpec { return &ImportSpec{} }

// NewFunc returns a new *Func.
func NewFunc() *Func {
	return &Func{
		TypeParams: maps.MakeOrderedMap[string, *Field](),
		Params:     maps.MakeOrderedMap[string, *Field](),
		Results:    maps.MakeOrderedMap[string, *Field](),
	}
}

// NewMethod returns a new *Method.
func NewMethod() *Method {
	return &Method{
		Func: *NewFunc(),
	}
}

// NewConst returns a new *Const.
func NewConst() *Const { return &Const{} }

// NewVar returns a new *Var.
func NewVar() *Var { return &Var{} }

// NewType returns a new *Type.
func NewType() *Type { return &Type{} }

// NewInterface returns a new *Interface.
func NewInterface() *Interface {
	return &Interface{
		Methods:    maps.MakeOrderedMap[string, *Method](),
		Interfaces: maps.MakeOrderedMap[string, *Field](),
	}
}

// NewStruct returns a new *Struct.
func NewStruct() *Struct { return &Struct{Fields: maps.MakeOrderedMap[string, *Field]()} }

// NewField returns a new *Field.
func NewField() *Field { return &Field{} }

func (self *Package) GetName() string    { return self.Name }
func (self *File) GetName() string       { return self.Name }
func (self *ImportSpec) GetName() string { return self.Name }
func (self *Func) GetName() string       { return self.Name }
func (self *Method) GetName() string     { return self.Name }
func (self *Var) GetName() string        { return self.Name }
func (self *Const) GetName() string      { return self.Name }
func (self *Type) GetName() string       { return self.Name }
func (self *Interface) GetName() string  { return self.Name }
func (self *Struct) GetName() string     { return self.Name }
func (self *Field) GetName() string      { return self.Name }
