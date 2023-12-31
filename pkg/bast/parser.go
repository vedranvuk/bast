// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/packages"
)

func parsePackage(name string, in *packages.Package, out map[string]*Package) {
	var val = NewPackage()
	val.Name = in.Name
	for _, file := range in.Syntax {
		parseFile(name, file, val.Files)
	}
	out[val.Name] = val
	return
}

func parseFile(name string, in *ast.File, out map[string]*File) {
	var val = NewFile()
	val.Name = name
	var cg []string
	for _, comment := range in.Comments {
		parseCommentGroup(comment, &cg)
		val.Comments = append(val.Comments, cg)
	}
	parseCommentGroup(in.Doc, &val.Doc)
	for _, imprt := range in.Imports {
		parseImportSpec(imprt, val.Imports)
	}
	for _, d := range in.Decls {
		parseDeclarations(d.(ast.Node), val.Declarations)
	}
	out[val.Name] = val
	return
}

func parseDeclarations(in ast.Node, out map[string]Declaration) {
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

func parseImportSpec(in *ast.ImportSpec, out map[string]*Import) {
	var val = NewImport()
	if in.Name != nil {
		val.Name = printExpr(in.Name)
	}
	val.Path = printExpr(in.Path)
	parseCommentGroup(in.Doc, &val.Doc)
	parseCommentGroup(in.Comment, &val.Comment)
	return
}

func parseConsts(in *ast.GenDecl, out map[string]Declaration) {
	var lastType string
	for _, spec := range in.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			for i := 0; i < len(s.Names); i++ {
				var val = NewConst()
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
				out[val.Name] = val
			}
		default:
			panic("parseConsts: unsupported spec")
		}
	}
}

func parseVar(in *ast.ValueSpec, out map[string]Declaration) {
	for i := 0; i < len(in.Names); i++ {
		var val = NewVar()
		val.Name = printExpr(in.Names[i])
		parseCommentGroup(in.Comment, &val.Comment)
		parseCommentGroup(in.Doc, &val.Doc)
		val.Type = printExpr(in.Type)
		if in.Values != nil {
			val.Value = printExpr(in.Values[i])
		}
		out[val.Name] = val
	}
}

func parseFunc(in *ast.FuncDecl, out map[string]Declaration) {
	var val = NewFunc()
	val.Name = printExpr(in.Name)
	parseCommentGroup(in.Doc, &val.Doc)
	parseFieldList(in.Type.TypeParams, val.TypeParams)
	parseFieldList(in.Type.Params, val.Params)
	parseFieldList(in.Type.Results, val.Results)
	out[val.Name] = val
}

func parseMethod(in *ast.FuncDecl, out map[string]Declaration) {
	var val = NewMethod()
	val.Name = printExpr(in.Name)
	parseCommentGroup(in.Doc, &val.Doc)
	parseFieldList(in.Recv, val.Receivers)
	parseFieldList(in.Type.TypeParams, val.TypeParams)
	parseFieldList(in.Type.Params, val.Params)
	parseFieldList(in.Type.Results, val.Results)
	out[val.Name] = val
	return
}

func parseFuncType(in *ast.TypeSpec, out map[string]Declaration) {
	var val = NewFunc()
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	var ft = in.Type.(*ast.FuncType)
	parseFieldList(ft.TypeParams, val.TypeParams)
	parseFieldList(ft.Params, val.Params)
	parseFieldList(ft.Results, val.Results)
	out[val.Name] = val
	return
}

func parseType(in *ast.TypeSpec, out map[string]Declaration) {
	var val = NewType()
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Comment)
	val.Name = printExpr(in.Name)
	val.Type = printExpr(in.Type)
	val.IsAlias = in.Assign.IsValid()
	out[val.Name] = val
	return
}

func parseFieldList(in *ast.FieldList, out map[string]*Field) {
	if in == nil {
		return
	}
	for _, field := range in.List {
		if field.Names == nil {
			out[""] = &Field{
				Type: printExpr(field.Type),
			}
			continue
		}
		for _, name := range field.Names {
			var val = NewField()
			parseCommentGroup(field.Doc, &val.Doc)
			parseCommentGroup(field.Comment, &val.Comment)
			val.Name = printExpr(name)
			val.Type = printExpr(field.Type)
			val.Tag = printExpr(field.Tag)
			out[val.Name] = val
		}
	}
}

func parseInterface(in *ast.TypeSpec, out map[string]Declaration) {
	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}
	var val = NewInterface()
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	parseInterfaceMethods(it.Methods, val.Methods)
	out[val.Name] = val
	return
}

func parseInterfaceMethods(in *ast.FieldList, out map[string]*Method) {
	for _, field := range in.List {
		for _, name := range field.Names {
			var val = NewMethod()
			val.Name = printExpr(name)
			parseCommentGroup(field.Comment, &val.Comment)
			parseCommentGroup(field.Doc, &val.Doc)
			var ft = field.Type.(*ast.FuncType)
			parseFieldList(ft.TypeParams, val.TypeParams)
			parseFieldList(ft.Params, val.Params)
			parseFieldList(ft.Results, val.Results)
			out[val.Name] = val
		}
	}
}

func parseStruct(in *ast.TypeSpec, out map[string]Declaration) {
	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}
	var val = NewStruct()
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Doc)
	val.Name = printExpr(in.Name)
	for _, field := range st.Fields.List {
		parseStructField(field, val.Fields)
	}
	out[val.Name] = val
	return
}

func parseStructField(in *ast.Field, out map[string]*Field) {
	for _, name := range in.Names {
		var val = NewField()
		parseCommentGroup(in.Comment, &val.Comment)
		parseCommentGroup(in.Doc, &val.Doc)
		val.Name = printExpr(name)
		val.Type = printExpr(in.Type)
		val.Tag = printExpr(in.Tag)
		out[val.Name] = val
	}
	return
}
