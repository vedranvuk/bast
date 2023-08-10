// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains bast object model types.

package bast

// Package
type Package struct {
	// Name is the package name, without path.
	Name string
	// Files is a list of files in the package.
	Files []*File
}

// File describes a go source file.
type File struct {
	// Comments are the file comments, grouped by separation, without positions,
	// including docs.
	Comments [][]string
	// Doc is the file doc comment.
	Doc []string
	// Name is the File name, without path.
	Name string
	// Imports is a list of file imports.
	Imports []*Import
	// Declarations is a list of top level declarations in the file.
	Declarations []Declaration
}

// Import represents a package import entry in a File.
type Import struct {
	// Comment is the import comment.
	Comment []string
	// Doc is the import doc.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// Declaration represents a top level declaration in a Go file.
type Declaration interface {
	// GetName returns the Declaration name.
	GetName() string
}

// Func represents a func.
type Func struct {
	// Comment is the func comment.
	Comment []string
	// Doc is the func doc comment.
	Doc []string
	// Name is the func name.
	Name string
	// TypeParams are type parameters.
	TypeParams []*Field
	//  Params is a list of func arguments.
	Params []*Field
	// Results is a list of func returns.
	Results []*Field
}

// Method represents a method.
type Method struct {
	// Func embeds all Func properties.
	Func
	// Receiver is the method receiver.
	Receivers []*Field
}

// Const represents a constant
type Const struct {
	// Comment is the const comment.
	Comment []string
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Const represents a constant
type Var struct {
	// Comment is the const comment.
	Comment []string
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Struct represents a struct type.
type Type struct {
	// Comment is the struct comment.
	Comment []string
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Type is Type's underlying type.
	Type string
	// IsAlias is true if type is an alias of the type it derives from.
	IsAlias bool
}

// Array is an array type, same as Type but with a Length.
type Array struct {
	// Comment is the field comment.
	Comment []string
	// Doc is the field doc comment.
	Doc []string
	// Name is the field name.
	Name string
	// Length is the array length, if any.
	Length string
	// Type is the array'd type.
	Type string
}

// Interface represents an interface.
type Interface struct {
	// Comment is the interface comment.
	Comment []string
	// Doc is the interface doc comment.
	Doc []string
	// Name is the interface name.
	Name string
	// Methods is a list of methods defined by the interface.
	Methods []*Method
}

// Struct represents a struct type.
type Struct struct {
	// Comment is the struct comment.
	Comment []string
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Fields is a list of struct fields.
	Fields []*Field
}

// Field represents a struct field.
type Field struct {
	// Comment is the field comment.
	Comment []string
	// Doc is the field doc comment.
	Doc []string
	// Name is the field name.
	Name string
	// Type is the field type.
	Type string
	// Tag is the field raw tag string.
	Tag string
}

func (self *Package) GetName() string   { return self.Name }
func (self *File) GetName() string      { return self.Name }
func (self *Import) GetName() string    { return self.Name }
func (self *Func) GetName() string      { return self.Name }
func (self *Method) GetName() string    { return self.Name }
func (self *Var) GetName() string       { return self.Name }
func (self *Const) GetName() string     { return self.Name }
func (self *Type) GetName() string      { return self.Name }
func (self *Array) GetName() string     { return self.Name }
func (self *Interface) GetName() string { return self.Name }
func (self *Struct) GetName() string    { return self.Name }
func (self *Field) GetName() string     { return self.Name }
