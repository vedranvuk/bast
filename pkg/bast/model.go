// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains bast object model types.

package bast

import "golang.org/x/tools/go/packages"

// Declaration represents a top level declaration in a Go file.
type Declaration interface {
	// GetName returns the Declaration name.
	GetName() string
}

// Package contians info about a Go package.
type Package struct {
	// Name is the package name, without path.
	Name string
	// Files is a list of files in the package.
	Files map[string]*File
	// pkg is the parsed package.
	pkg *packages.Package
}

// File contians info about a Go source file.
type File struct {
	// Comments are the file comments, grouped by separation, including docs.
	Comments [][]string
	// Doc is the file doc comment.
	Doc []string
	// Name is the File name, without path.
	Name string
	// Imports is a list of file imports.
	Imports map[string]*Import
	// Declarations is a list of top level declarations in the file.
	Declarations map[string]Declaration
}

// Import contians info about an import.
type Import struct {
	// Doc is the import doc.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// Func contains info about a function.
type Func struct {
	// Doc is the func doc comment.
	Doc []string
	// Name is the func name.
	Name string
	// TypeParams are type parameters.
	TypeParams map[string]*Field
	//  Params is a list of func arguments.
	Params map[string]*Field
	// Results is a list of func returns.
	Results map[string]*Field
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
	Methods map[string]*Method
	// Interface is a list of inherited interfaces.
	//
	// Map is keyed by the embeded interface type name.
	Interfaces map[string]*Field
}

// Struct contains info about a struct.
type Struct struct {
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Fields is a list of struct fields.
	Fields map[string]*Field
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

// NewPackage returns a new *Package.
func NewPackage() *Package {
	return &Package{
		Files: make(map[string]*File),
	}
}

// NewFile returns a new *File.
func NewFile() *File {
	return &File{
		Imports:      make(map[string]*Import),
		Declarations: make(map[string]Declaration),
	}
}

// NewImport returns a new *Import.
func NewImport() *Import { return &Import{} }

// NewFunc returns a new *Func.
func NewFunc() *Func {
	return &Func{
		TypeParams: make(map[string]*Field),
		Params:     make(map[string]*Field),
		Results:    make(map[string]*Field),
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
		Methods:    make(map[string]*Method),
		Interfaces: make(map[string]*Field),
	}
}

// NewStruct returns a new *Struct.
func NewStruct() *Struct { return &Struct{Fields: make(map[string]*Field)} }

// NewField returns a new *Field.
func NewField() *Field { return &Field{} }

func (self *Package) GetName() string   { return self.Name }
func (self *File) GetName() string      { return self.Name }
func (self *Import) GetName() string    { return self.Name }
func (self *Func) GetName() string      { return self.Name }
func (self *Method) GetName() string    { return self.Name }
func (self *Var) GetName() string       { return self.Name }
func (self *Const) GetName() string     { return self.Name }
func (self *Type) GetName() string      { return self.Name }
func (self *Interface) GetName() string { return self.Name }
func (self *Struct) GetName() string    { return self.Name }
func (self *Field) GetName() string     { return self.Name }
