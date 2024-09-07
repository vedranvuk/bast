// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func (self *Bast) parsePackage(in *packages.Package, out map[string]*Package) {

	var val = NewPackage()
	val.Name = in.Name

	for idx, file := range in.Syntax {
		self.parseFile(filepath.Base(in.CompiledGoFiles[idx]), file, val.Files)
	}

	out[val.Name] = val

	return
}

func (self *Bast) parseFile(name string, in *ast.File, out map[string]*File) {

	var val = NewFile()
	val.Name = name

	for _, comment := range in.Comments {
		var cg []string
		self.parseCommentGroup(comment, &cg)
		val.Comments = append(val.Comments, cg)
	}

	self.parseCommentGroup(in.Doc, &val.Doc)

	for _, imprt := range in.Imports {
		self.parseImportSpec(imprt, val.Imports)
	}

	for _, d := range in.Decls {
		self.parseDeclarations(d.(ast.Node), val.Declarations)
	}

	out[val.Name] = val

	return
}

func (self *Bast) parseDeclarations(in ast.Node, out map[string]Declaration) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			self.parseConsts(n, out)
		case token.VAR:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					self.parseVar(s, out)
				}
			}
		case token.TYPE:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Assign != token.NoPos {
						continue
					}
					switch s.Type.(type) {
					case *ast.InterfaceType:
						self.parseInterface(n, s, out)
					case *ast.StructType:
						self.parseStruct(n, s, out)
					case *ast.ArrayType:
						self.parseType(n, s, out)
					case *ast.FuncType:
						self.parseFuncType(n, s, out)
					case *ast.Ident:
						self.parseType(n, s, out)
					case *ast.ChanType:
						self.parseType(n, s, out)
					case *ast.MapType:
						self.parseType(n, s, out)
					}
				}
			}
		}
	case *ast.FuncDecl:
		if n.Recv != nil {
			self.parseMethod(n, out)
		} else {
			self.parseFunc(n, out)
		}
	}
	return
}

func (self *Bast) parseCommentGroup(in *ast.CommentGroup, out *[]string) {
	if in == nil {
		return
	}
	for _, entry := range in.List {
		*out = append(*out, entry.Text)
	}
	return
}

func (self *Bast) parseImportSpec(in *ast.ImportSpec, out map[string]*Import) {
	var val = NewImport()
	if in.Name != nil {
		val.Name = self.printExpr(in.Name)
	}
	val.Path = self.printExpr(in.Path)
	self.parseCommentGroup(in.Doc, &val.Doc)
	if in.Name != nil {
		out[in.Name.Name] = val
	} else {
		out[path.Base(val.Path)] = val
	}
	return
}

func (self *Bast) parseConsts(in *ast.GenDecl, out map[string]Declaration) {
	var lastType string
	for _, spec := range in.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			for i := 0; i < len(s.Names); i++ {
				var val = NewConst()
				val.Name = self.printExpr(s.Names[i])
				self.parseCommentGroup(s.Doc, &val.Doc)
				if s.Type != nil {
					val.Type = self.printExpr(s.Type)
					lastType = val.Type
				} else if lastType != "" {
					val.Type = lastType
				}
				if s.Values != nil {
					val.Value = self.printExpr(s.Values[i])
				}
				out[val.Name] = val
			}
		default:
			panic("parseConsts: unsupported spec")
		}
	}
}

func (self *Bast) parseVar(in *ast.ValueSpec, out map[string]Declaration) {
	for i := 0; i < len(in.Names); i++ {
		var val = NewVar()
		val.Name = self.printExpr(in.Names[i])
		self.parseCommentGroup(in.Doc, &val.Doc)
		val.Type = self.printExpr(in.Type)
		if in.Values != nil {
			val.Value = self.printExpr(in.Values[i])
		}
		out[val.Name] = val
	}
}

func (self *Bast) parseFunc(in *ast.FuncDecl, out map[string]Declaration) {
	var val = NewFunc()
	val.Name = self.printExpr(in.Name)
	self.parseCommentGroup(in.Doc, &val.Doc)
	self.parseFieldList(in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(in.Type.Params, val.Params)
	self.parseFieldList(in.Type.Results, val.Results)
	out[val.Name] = val
}

func (self *Bast) parseMethod(in *ast.FuncDecl, out map[string]Declaration) {
	var val = NewMethod()
	self.parseCommentGroup(in.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)

	if in.Recv != nil {
		val.Receiver = NewField()
		val.Receiver.Name = self.printExpr(in.Recv.List[0].Names[0])
		val.Receiver.Type = self.printExpr(in.Recv.List[0].Type)
	}

	self.parseFieldList(in.Type.TypeParams, val.TypeParams)
	self.parseFieldList(in.Type.Params, val.Params)
	self.parseFieldList(in.Type.Results, val.Results)
	out[val.Name] = val
	return
}

func (self *Bast) parseFuncType(g *ast.GenDecl, in *ast.TypeSpec, out map[string]Declaration) {
	var val = NewFunc()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	var ft = in.Type.(*ast.FuncType)
	self.parseFieldList(ft.TypeParams, val.TypeParams)
	self.parseFieldList(ft.Params, val.Params)
	self.parseFieldList(ft.Results, val.Results)
	out[val.Name] = val
	return
}

func (self *Bast) parseType(g *ast.GenDecl, in *ast.TypeSpec, out map[string]Declaration) {
	var val = NewType()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	val.Type = self.printExpr(in.Type)
	val.IsAlias = in.Assign.IsValid()
	out[val.Name] = val
	return
}

func (self *Bast) parseFieldList(in *ast.FieldList, out map[string]*Field) {
	if in == nil {
		return
	}
	for idx, field := range in.List {
		if field.Names == nil {
			out[fmt.Sprintf("%s [%d]", self.printExpr(field.Type), idx)] = &Field{
				Type: self.printExpr(field.Type),
			}
			continue
		}
		for _, name := range field.Names {
			var val = NewField()
			self.parseCommentGroup(field.Doc, &val.Doc)
			val.Name = self.printExpr(name)
			val.Type = self.printExpr(field.Type)
			val.Tag = self.printExpr(field.Tag)
			out[val.Name] = val
		}
	}
}

func (self *Bast) parseInterface(g *ast.GenDecl, in *ast.TypeSpec, out map[string]Declaration) {

	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	var val = NewInterface()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)

	for _, method := range it.Methods.List {
		if len(method.Names) == 0 {
			var intf = NewField()
			self.parseCommentGroup(method.Doc, &intf.Doc)
			intf.Type = self.printExpr(method.Type)
			intf.Name = intf.Type
			intf.Unnamed = true
			val.Interfaces[intf.Type] = intf
		} else {
			var m = NewMethod()
			self.parseCommentGroup(method.Doc, &m.Doc)
			m.Name = self.printExpr(method.Names[0])
			var ft = method.Type.(*ast.FuncType)
			self.parseFieldList(ft.TypeParams, m.TypeParams)
			self.parseFieldList(ft.Params, m.Params)
			self.parseFieldList(ft.Results, m.Results)
			val.Methods[m.Name] = m
		}
	}

	out[val.Name] = val

	return
}

func (self *Bast) parseStruct(g *ast.GenDecl, in *ast.TypeSpec, out map[string]Declaration) {
	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}
	var val = NewStruct()
	self.parseCommentGroup(g.Doc, &val.Doc)
	val.Name = self.printExpr(in.Name)
	for _, field := range st.Fields.List {
		self.parseStructField(field, val.Fields)
	}
	out[val.Name] = val
	return
}

func (self *Bast) parseStructField(in *ast.Field, out map[string]*Field) {

	var val = NewField()

	if len(in.Names) == 0 {
		val.Unnamed = true
		val.Type = self.printExpr(in.Type)
		val.Tag = self.printExpr(in.Tag)
		out[val.Name] = val
		return
	}

	for _, name := range in.Names {
		self.parseCommentGroup(in.Doc, &val.Doc)
		val.Name = self.printExpr(name)
		val.Type = self.printExpr(in.Type)
		val.Tag = self.printExpr(in.Tag)
		out[val.Name] = val
	}

	return
}
