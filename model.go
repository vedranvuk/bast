// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

// Bast is the top-level type that holds parsed packages and their declarations.
//
// It provides methods for querying and retrieving declarations across all packages.
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
		packages: maps.NewOrderedMap[string, *Package](),
	}
}

// declarations is the interface for all bast declaration types.
type declarations interface {
	*Var | *Const | *Func | *Method | *Type | *Struct | *Interface
}

// Package represents a parsed Go package.
//
// It contains the package name, import path, files, and top-level declarations.
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

// Var returns the variable named name from this package, or nil if not found.
func (self *Package) Var(name string) (out *Var) {
	return pkgDecl[*Var](self.Path, name, self.bast.packages)
}

// Const returns the constant named name from this package, or nil if not found.
func (self *Package) Const(name string) (out *Const) {
	return pkgDecl[*Const](self.Path, name, self.bast.packages)
}

// Func returns the function named name from this package, or nil if not found.
func (self *Package) Func(name string) (out *Func) {
	return pkgDecl[*Func](self.Path, name, self.bast.packages)
}

// Method returns the method named name from this package, or nil if not found.
func (self *Package) Method(name string) (out *Method) {
	return pkgDecl[*Method](self.Path, name, self.bast.packages)
}

// Type returns the type named name from this package, or nil if not found.
func (self *Package) Type(name string) (out *Type) {
	return pkgDecl[*Type](self.Path, name, self.bast.packages)
}

// Struct returns the struct named name from this package, or nil if not found.
func (self *Package) Struct(name string) (out *Struct) {
	return pkgDecl[*Struct](self.Path, name, self.bast.packages)
}

// Interface returns the interface named name from this package, or nil if not found.
func (self *Package) Interface(name string) (out *Interface) {
	return pkgDecl[*Interface](self.Path, name, self.bast.packages)
}

// DeclFile returns the full filename of the file containing the declaration named typeName in this package.
// It returns an empty string if not found.
func (self *Package) DeclFile(typeName string) string {
	for _, file := range self.Files.Values() {
		if _, ok := file.Declarations.Get(typeName); ok {
			return file.Name
		}
	}
	return ""
}

// HasDecl returns true if a declaration named typeName exists in this package.
func (self *Package) HasDecl(typeName string) bool {
	return self.DeclFile(typeName) != ""
}

// PackageMap is an ordered map of packages keyed by their import path.
type PackageMap = maps.OrderedMap[string, *Package]

// File represents a parsed Go source file.
//
// It contains comments, imports, and top-level declarations.
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

// Var returns the variable named name from this file, or nil if not found.
func (self *File) Var(name string) (out *Var) { return fileDecl[*Var](name, self) }

// Const returns the constant named name from this file, or nil if not found.
func (self *File) Const(name string) (out *Const) { return fileDecl[*Const](name, self) }

// Func returns the function named name from this file, or nil if not found.
func (self *File) Func(name string) (out *Func) { return fileDecl[*Func](name, self) }

// Method returns the method named name from this file, or nil if not found.
func (self *File) Method(name string) (out *Method) { return fileDecl[*Method](name, self) }

// Type returns the type named name from this file, or nil if not found.
func (self *File) Type(name string) (out *Type) { return fileDecl[*Type](name, self) }

// Struct returns the struct named name from this file, or nil if not found.
func (self *File) Struct(name string) (out *Struct) { return fileDecl[*Struct](name, self) }

// Interface returns the interface named name from this file, or nil if not found.
func (self *File) Interface(name string) (out *Interface) { return fileDecl[*Interface](name, self) }

// HasDecl returns true if a declaration named name exists in this file.
func (self *File) HasDecl(name string) (b bool) {
	_, b = self.Declarations.Get(name)
	return
}

// ImportSpecFromSelector returns the ImportSpec for the given selector expression (e.g., "pkg.Type").
// It returns nil if the import is not found or the selector is invalid.
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

// fileDecl is an internal helper to retrieve a declaration of type T from the file.
func fileDecl[T declarations](declName string, file *File) (out T) {
	if decl, ok := file.Declarations.Get(declName); ok {
		out, _ = decl.(T)
	}
	return
}

// FileMap is an ordered map of files keyed by their filename in parse order.
type FileMap = maps.OrderedMap[string, *File]

// ImportSpec represents an import specification for a package.
type ImportSpec struct {
	// Doc is the import doc comment.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// Base returns the base name of the imported package path.
func (self *ImportSpec) Base() string { return path.Base(self.Path) }

// ImportSpecMap is an ordered map of import specs keyed by their path in parse order.
type ImportSpecMap = maps.OrderedMap[string, *ImportSpec]

// Declaration is the interface implemented by all top-level declarations.
type Declaration interface {
	// GetFile returns the declarations parent file.
	GetFile() *File
	// GetPackage returns the declarations parent package.
	GetPackage() *Package
}

// DeclarationMap is an ordered map of declarations keyed by their name in parse order.
type DeclarationMap = maps.OrderedMap[string, Declaration]

// Model is the base struct embedded by all declarations.
//
// It provides common fields like documentation and name, and implements the Declaration interface.
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

// GetFile returns the parent file of the declaration.
func (self *Model) GetFile() *File { return self.file }

// GetPackage returns the parent package of the declaration.
func (self *Model) GetPackage() *Package { return self.file.pkg }

// ImportSpecBySelectorExpr returns the ImportSpec for the package from which the type
// qualified by selectorExpr (e.g., "pkg.TypeName") is imported.
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

// ResolveBasicType resolves the underlying basic type name for the given typeName
// by searching the type hierarchy of the parsed packages.
//
// If typeName is already a basic type name, it returns typeName as is.
// If no basic type is found, it returns an empty string.
//
// This method requires Config.TypeChecking to be enabled.
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

// Var represents a top-level variable declaration.
type Var struct {
	Model
	// Type is the variable's type, empty if inferred.
	Type string
	// Value is the variable's initial value, empty if not specified.
	Value string
}

// Const represents a top-level constant declaration.
type Const struct {
	Model
	// Type is the constant's type, empty if inferred.
	Type string
	// Value is the constant's value.
	Value string
}

// Field represents a field in a struct, a parameter or result in a function,
// or a receiver in a method.
type Field struct {
	Model

	// Type is the field's type.
	//
	// For method receivers, this is the bare type name without "*" or type parameters.
	// Use Pointer to check for pointer receivers, and inspect the parent for type parameters.
	Type string

	// Tag is the field's raw struct tag string.
	Tag string

	// Unnamed is true if the field is unnamed (embedded field).
	Unnamed bool

	// Pointer is true if this is a pointer receiver for a method.
	Pointer bool
}

// Clone returns a copy of the field.
func (self *Field) Clone() *Field {
	return &Field{
		Model: Model{
			Doc:  self.Doc,
			Name: self.Name,
			file: self.file,
		},
		Type:    self.Type,
		Tag:     self.Tag,
		Unnamed: self.Unnamed,
		Pointer: self.Pointer,
	}
}

// FieldMap is an ordered map of fields keyed by name in parse order.
type FieldMap = maps.OrderedMap[string, *Field]

// Func represents a top-level function declaration.
type Func struct {
	Model
	// TypeParams are the function's type parameters.
	TypeParams *FieldMap
	// Params are the function's parameters.
	Params *FieldMap
	// Results are the function's return values.
	Results *FieldMap
}

// Method represents a top-level method declaration.
type Method struct {
	Func
	// Receiver is the method's receiver, or nil for interface methods.
	Receiver *Field
}

// MethodMap is an ordered map of methods keyed by name in parse order.
type MethodMap = maps.OrderedMap[string, *Method]

// Type represents a top-level type declaration (not struct or interface).
type Type struct {
	Model
	// Type is the underlying type of this type declaration.
	//
	// This may be a qualified selector like "pkg.Type".
	Type string
	// IsAlias is true if this is a type alias (using := instead of =).
	IsAlias bool
	// TypeParams are the type's type parameters.
	TypeParams *FieldMap
}

// Struct represents a top-level struct type declaration.
type Struct struct {
	Model
	// Fields are the struct's fields.
	Fields *FieldMap
	// TypeParams are the struct's type parameters.
	TypeParams *FieldMap
}

// Methods returns the methods defined on this struct.
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

// Interface represents a top-level interface type declaration.
type Interface struct {
	Model
	// Methods are the methods declared by this interface.
	Methods *MethodMap
	// Interfaces are the embedded interfaces.
	//
	// Keyed by the embedded interface type name.
	Interfaces *InterfaceMap
	// TypeParams are the interface's type parameters.
	TypeParams *FieldMap
}

// NewPackage creates a new Package with the given name, path, and underlying packages.Package.
func NewPackage(name, path string, pkg *packages.Package) *Package {
	return &Package{
		Name:  name,
		Path:  path,
		Files: maps.NewOrderedMap[string, *File](),
		pkg:   pkg,
	}
}

// NewFile creates a new File for the given package and filename.
func NewFile(pkg *Package, name string) *File {
	return &File{
		Name:         name,
		Imports:      maps.NewOrderedMap[string, *ImportSpec](),
		Declarations: maps.NewOrderedMap[string, Declaration](),
		pkg:          pkg,
	}
}

// NewImport creates a new ImportSpec with the given name and path.
func NewImport(name, path string) *ImportSpec {
	return &ImportSpec{
		Name: name,
		Path: path,
	}
}

// NewFunc creates a new Func for the given file and name.
func NewFunc(file *File, name string) *Func {
	return &Func{
		Model: Model{
			Name: name,
			file: file,
		},
		TypeParams: maps.NewOrderedMap[string, *Field](),
		Params:     maps.NewOrderedMap[string, *Field](),
		Results:    maps.NewOrderedMap[string, *Field](),
	}
}

// NewMethod creates a new Method for the given file and name.
func NewMethod(file *File, name string) *Method {
	return &Method{
		Func: *NewFunc(file, name),
	}
}

// NewConst creates a new Const for the given file, name, and type.
func NewConst(file *File, name, typ string) *Const {
	return &Const{
		Model: Model{
			Name: name,
			file: file,
		},
		Type: typ,
	}

}

// NewVar creates a new Var for the given file, name, and type.
func NewVar(file *File, name, typ string) *Var {
	return &Var{
		Model: Model{
			Name: name,
			file: file,
		},
		Type: typ,
	}
}

// NewType creates a new Type for the given file, name, and underlying type.
func NewType(file *File, name, typ string) *Type {
	return &Type{
		Model: Model{
			Name: name,
			file: file,
		},
		Type:       typ,
		TypeParams: maps.NewOrderedMap[string, *Field](),
	}
}

// NewStruct creates a new Struct for the given file and name.
func NewStruct(file *File, name string) *Struct {
	return &Struct{
		Model: Model{
			Name: name,
			file: file,
		},
		Fields:     maps.NewOrderedMap[string, *Field](),
		TypeParams: maps.NewOrderedMap[string, *Field](),
	}

}

// NewField creates a new Field for the given file and name.
func NewField(file *File, name string) *Field {
	return &Field{
		Model: Model{
			Name: name,
			file: file,
		},
	}
}

// NewInterface creates a new Interface for the given file and name.
func NewInterface(file *File, name string) *Interface {
	return &Interface{
		Model: Model{
			Name: name,
			file: file,
		},
		Methods:    maps.NewOrderedMap[string, *Method](),
		Interfaces: maps.NewOrderedMap[string, *Interface](),
		TypeParams: maps.NewOrderedMap[string, *Field](),
	}
}

// InterfaceMap is an ordered map of interfaces keyed by name in parse order.
type InterfaceMap = maps.OrderedMap[string, *Interface]
