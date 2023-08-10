// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

func parsePackage(name string, in *packages.Package, out *[]*Package) {
	var val = new(Package)
	val.Name = in.Name
	for _, file := range in.Syntax {
		parseFile(name, file, &val.Files)
	}
	*out = append(*out, val)
	return
}

func parseFile(name string, in *ast.File, out *[]*File) {
	var val = new(File)
	val.Name = name
	var cg []string
	for _, comment := range in.Comments {
		parseCommentGroup(comment, &cg)
		val.Comments = append(val.Comments, cg)
	}
	parseCommentGroup(in.Doc, &val.Doc)
	for _, imprt := range in.Imports {
		parseImportSpec(imprt, &val.Imports)
	}
	for _, d := range in.Decls {
		parseDeclarations(d.(ast.Node), &val.Declarations)
	}
	*out = append(*out, val)
	return
}

func parseDeclarations(in ast.Node, out *[]Declaration) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			parseConsts(n, out)
		case token.VAR:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					parseVar(s, out)
				default:
					panic("!")
				}
			}
		case token.TYPE:
			for _, spec := range n.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Assign != token.NoPos {
						continue
					}
					switch x := s.Type.(type) {
					case *ast.InterfaceType:
						parseInterface(s, out)
					case *ast.StructType:
						parseStruct(s, out)
					case *ast.ArrayType:
						parseType(s, out)
					case *ast.FuncType:
						parseFuncType(s, out)
					case *ast.Ident:
						parseType(s, out)
					case *ast.ChanType:
						parseType(s, out)
					case *ast.MapType:
						parseType(s, out)
					default:
						_ = x
						panic("parseDeclarations: unsuported gendecl spec type")
					}
				default:
					panic("parseDeclarations: unsupported gendecl spec")
				}
			}
		case token.IMPORT:
		default:
			panic("parseDeclarations: unsupported gendecl token")
		}
	case *ast.FuncDecl:
		if n.Recv != nil {
			parseMethod(n, out)
		} else {
			parseFunc(n, out)
		}
	default:
		panic("parseDeclarations: unsupported node")
	}
	return
}

func parseCommentGroup(in *ast.CommentGroup, out *[]string) {
	if in == nil {
		return
	}
	for _, entry := range in.List {
		*out = append(*out, entry.Text)
	}
	return
}

func parseImportSpec(in *ast.ImportSpec, out *[]*Import) {
	var val = new(Import)
	if in.Name != nil {
		val.Name = printExpr(in.Name)
	}
	val.Path = printExpr(in.Path)
	parseCommentGroup(in.Doc, &val.Doc)
	parseCommentGroup(in.Comment, &val.Comment)
	return
}

func parseConsts(in *ast.GenDecl, out *[]Declaration) {
	var lastType string
	for _, spec := range in.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			for i := 0; i < len(s.Names); i++ {
				var val = new(Const)
				val.Name = printExpr(s.Names[i])
				parseCommentGroup(s.Comment, &val.Comment)
				parseCommentGroup(s.Doc, &val.Doc)
				if s.Type != nil {
					val.Type = printExpr(s.Type)
					lastType = val.Type
				} else if lastType != "" {
					val.Type = lastType
				}
				if s.Values != nil {
					val.Value = printExpr(s.Values[i])
				}
				*out = append(*out, val)
			}
		default:
			panic("parseConsts: unsupported spec")
		}
	}
}

func parseVar(in *ast.ValueSpec, out *[]Declaration) {
	for i := 0; i < len(in.Names); i++ {
		var val = new(Var)
		val.Name = printExpr(in.Names[i])
		parseCommentGroup(in.Comment, &val.Comment)
		parseCommentGroup(in.Doc, &val.Doc)
		val.Type = printExpr(in.Type)
		if in.Values != nil {
			val.Value = printExpr(in.Values[i])
		}
		*out = append(*out, val)
	}
}

func parseFunc(in *ast.FuncDecl, out *[]Declaration) {
	var val = new(Func)
	val.Name = printExpr(in.Name)
	parseCommentGroup(in.Doc, &val.Doc)
	parseFieldList(in.Type.TypeParams, &val.TypeParams)
	parseFieldList(in.Type.Params, &val.Params)
	parseFieldList(in.Type.Results, &val.Results)
	*out = append(*out, val)
}

func parseMethod(in *ast.FuncDecl, out *[]Declaration) {
	var val = new(Method)
	val.Name = printExpr(in.Name)
	parseCommentGroup(in.Doc, &val.Doc)
	parseFieldList(in.Recv, &val.Receivers)
	parseFieldList(in.Type.TypeParams, &val.TypeParams)
	parseFieldList(in.Type.Params, &val.Params)
	parseFieldList(in.Type.Results, &val.Results)
	*out = append(*out, val)
	return
}

func parseFuncType(in *ast.TypeSpec, out *[]Declaration) {
	var val = new(Func)
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	var ft = in.Type.(*ast.FuncType)
	parseFieldList(ft.TypeParams, &val.TypeParams)
	parseFieldList(ft.Params, &val.Params)
	parseFieldList(ft.Results, &val.Results)
	*out = append(*out, val)
	return
}

func parseType(in *ast.TypeSpec, out *[]Declaration) {
	var val = new(Type)
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Comment)
	val.Name = printExpr(in.Name)
	val.Type = printExpr(in.Type)
	val.IsAlias = in.Assign.IsValid()
	*out = append(*out, val)
	return
}

func parseFieldList(in *ast.FieldList, out *[]*Field) {
	if in == nil {
		return
	}
	for _, field := range in.List {
		for _, name := range field.Names {
			var val = new(Field)
			parseCommentGroup(field.Doc, &val.Doc)
			parseCommentGroup(field.Comment, &val.Comment)
			val.Name = printExpr(name)
			val.Type = printExpr(field.Type)
			val.Tag = printExpr(field.Tag)
			*out = append(*out, val)
		}
	}
}

func parseInterface(in *ast.TypeSpec, out *[]Declaration) {
	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}
	var val = new(Interface)
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	parseInterfaceMethods(it.Methods, &val.Methods)
	*out = append(*out, val)
	return
}

func parseInterfaceMethods(in *ast.FieldList, out *[]*Method) {
	for _, field := range in.List {
		for _, name := range field.Names {
			var val = new(Method)
			val.Name = printExpr(name)
			parseCommentGroup(field.Comment, &val.Comment)
			parseCommentGroup(field.Doc, &val.Doc)
			var ft = field.Type.(*ast.FuncType)
			parseFieldList(ft.TypeParams, &val.TypeParams)
			parseFieldList(ft.Params, &val.Params)
			parseFieldList(ft.Results, &val.Results)
			*out = append(*out, val)
		}
	}
}

func parseStruct(in *ast.TypeSpec, out *[]Declaration) {
	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}
	var val = new(Struct)
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	for _, field := range st.Fields.List {
		parseStructField(field, &val.Fields)
	}
	*out = append(*out, val)
	return
}

func parseStructField(in *ast.Field, out *[]*Field) {
	for _, name := range in.Names {
		var val = new(Field)
		parseCommentGroup(in.Comment, &val.Comment)
		parseCommentGroup(in.Doc, &val.Doc)
		val.Name = printExpr(name)
		val.Type = printExpr(in.Type)
		val.Tag = printExpr(in.Tag)
		*out = append(*out, val)
	}
	return
}
