// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

// Load loads bast of inputs which can be module paths, absolute or relative
// paths to go files or packages. If no inputs are given Load returns an empty
// bast.
//
// Inputs that point to files, i.e. are outside of a package are put into a
// placeholder package named "command-line-package" which mirrors how
// "golang.org/x/tools/go/packages" names it.
//
// If an error occurs it is returned.
func Load(inputs ...string) (bast *Bast, err error) {

	bast = new(Bast)

	var (
		fp *Package
		fi os.FileInfo
		ff = token.NewFileSet()
	)

	for _, input := range inputs {
		if fi, err = os.Stat(input); err != nil {
			err = fmt.Errorf("stat input: %w", err)
			return
		}
		// Load complete package...
		if fi.IsDir() {
			var (
				fs   = token.NewFileSet()
				pkgs map[string]*ast.Package
			)
			if pkgs, err = parser.ParseDir(fs, input, nil, parseMode); err != nil {
				return
			}
			for _, pkg := range pkgs {
				parsePackage(fs, pkg, &bast.Packages)
			}
			continue
		}
		// ... or load file into placeholder root package.
		if fp == nil {
			fp = new(Package)
			fp.Name = "command-line-package"
		}
		var f *ast.File
		if f, err = parser.ParseFile(ff, input, nil, parseMode); err != nil {
			return
		}
		parseFile(ff, f, &fp.Files)
	}

	// Add placeholder package to parsed packages.
	if fp != nil {
		bast.Packages = append(bast.Packages, fp)
	}

	return
}

// ParseSrc returns a Bast of input source src or an error if one occurs.
func ParseSrc(src string) (bast *Bast, err error) {
	bast = new(Bast)
	var (
		pkg  = new(Package)
		fset = token.NewFileSet()
		file *ast.File
	)
	if file, err = parser.ParseFile(fset, "", src, parseMode); err != nil {
		return
	}
	pkg.Name = "command-line-package"
	parseFile(fset, file, &pkg.Files)
	bast.Packages = append(bast.Packages, pkg)
	return
}

// parseMode is the mode Bast uses for parsing go files.
const parseMode = parser.ParseComments | parser.DeclarationErrors | parser.AllErrors

func parsePackage(fs *token.FileSet, in *ast.Package, out *[]*Package) {
	var val = new(Package)
	val.Name = in.Name
	for _, file := range in.Files {
		parseFile(fs, file, &val.Files)
	}
	*out = append(*out, val)
	return
}

func parseFile(fs *token.FileSet, in *ast.File, out *[]*File) {
	var val = new(File)
	val.Name = in.Name.Name
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
		parseDeclarations(fs, d.(ast.Node), &val.Declarations)
	}
	*out = append(*out, val)
	return
}

func parseDeclarations(fs *token.FileSet, in ast.Node, out *[]Declaration) {
	switch n := in.(type) {
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			parseConst(n, out)
		case token.VAR:
			parseVar(n, out)
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
						parseArrayType(s, out)
					case *ast.FuncType:
						parseFuncType(s, out)
					case *ast.Ident:
						parseType(s, out)
					default:
						_ = x
						panic("parseDeclarations: unsuported type declaration")
					}
				default:
					panic("!")
				}
			}
		case token.IMPORT:
		default:
			panic("parseDeclarations: unsupported token type")
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

func parseFunc(in *ast.FuncDecl, out *[]Declaration) {
	var val = new(Func)
	val.Name = printExpr(in.Name)
	parseCommentGroup(in.Doc, &val.Doc)
	parseFieldList(in.Type.TypeParams, &val.TypeParams)
	parseFieldList(in.Type.Params, &val.Params)
	parseFieldList(in.Type.Results, &val.Results)
	*out = append(*out, val)
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

func parseArrayType(in *ast.TypeSpec, out *[]Declaration) {
	var val = new(Array)
	parseCommentGroup(in.Comment, &val.Comment)
	parseCommentGroup(in.Doc, &val.Comment)
	val.Name = printExpr(in.Name)
	val.Type = printExpr(in.Type)
	val.Length = printExpr(in.Type.(*ast.ArrayType).Len)
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

func parseConst(in *ast.GenDecl, out *[]Declaration) {
	var lastType string
	for _, spec := range in.Specs {
		var (
			vs, ok = spec.(*ast.ValueSpec)
			id     *ast.Ident
			val    = new(Const)
		)
		if !ok {
			continue
		}
		for i := 0; i < len(vs.Names); i++ {
			val.Name = vs.Names[i].Name
			parseCommentGroup(vs.Comment, &val.Comment)
			parseCommentGroup(vs.Doc, &val.Doc)
			if vs.Type != nil {
				if id, ok = vs.Type.(*ast.Ident); !ok {
					continue
				}
				val.Type = id.Name
				lastType = id.Name
			} else if lastType != "" {
				val.Type = lastType
			}
			if vs.Values != nil {
				switch v := vs.Values[i].(type) {
				case *ast.Ident:
					val.Value = v.Name
				case *ast.BasicLit:
					val.Value, _ = strconv.Unquote(v.Value)
				case *ast.BinaryExpr:
					var (
						lit *ast.BasicLit
					)
					if id, ok = v.X.(*ast.Ident); !ok || id.Name != "iota" {
						continue
					}
					if lit, ok = v.Y.(*ast.BasicLit); !ok {
						continue
					}
					val.Value = fmt.Sprintf("%s %s %s", id.Name, v.Op.String(), lit.Value)
				default:
					continue
				}
			}
			*out = append(*out, val)
		}
	}
}

func parseVar(in *ast.GenDecl, out *[]Declaration) {
	var lastType string
	for _, spec := range in.Specs {
		var (
			vs, ok = spec.(*ast.ValueSpec)
			id     *ast.Ident
			val    = new(Var)
		)
		if !ok {
			continue
		}
		for i := 0; i < len(vs.Names); i++ {
			val.Name = vs.Names[i].Name
			parseCommentGroup(vs.Comment, &val.Comment)
			parseCommentGroup(vs.Doc, &val.Doc)
			if vs.Type != nil {
				if id, ok = vs.Type.(*ast.Ident); !ok {
					continue
				}
				val.Type = id.Name
				lastType = id.Name
			} else if lastType != "" {
				val.Type = lastType
			}
			if vs.Values != nil {
				val.Value = printExpr(vs.Values[i])
			}
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
	for _, method := range it.Methods.List {
		_ = method
		// parseMethod(method, &val.Methods)
	}
	*out = append(*out, val)
	return
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
