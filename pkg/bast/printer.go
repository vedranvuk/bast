// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains functions for printing bast.

package bast

// TODO Limit types that get printed.
// TODO Config and formatting.

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
	"text/tabwriter"
)

func Print(w io.Writer, bast *Bast) {
	DefaultConfig().Print(w, bast)
}

func DefaultConfig() *Config { return &Config{true, true, true, true, true, true, true} }

type Config struct {
	PrintConsts     bool
	PrintVars       bool
	PrintTypes      bool
	PrintFuncs      bool
	PrintMethods    bool
	PrintStructs    bool
	PrintInterfaces bool
}

func (self *Config) Print(w io.Writer, bast *Bast) {
	var wr = tabwriter.NewWriter(w, 2, 2, 2, 32, 0)
	var p = func(format string, args ...any) { fmt.Fprintf(wr, format, args...) }
	for _, pkg := range bast.Packages {
		p("Package\t\"%s\"\n", pkg.Name)
		for _, file := range pkg.Files {
			p("\tFile\t\"%s\"\n", file.Name)
			if self.PrintConsts {
				for _, decl := range file.Declarations {
					var c *Const
					c, ok := decl.(*Const)
					if !ok {
						continue
					}
					p("\t\tConst\t\"%s\"\t(%s)\t'%s'\n", c.Name, c.Type, c.Value)
				}
			}
			if self.PrintVars {
				for _, decl := range file.Declarations {
					var v *Var
					v, ok := decl.(*Var)
					if !ok {
						continue
					}
					p("\t\tVar\t\"%s\"\t(%s)\t'%s'\n", v.Name, v.Type, v.Value)
				}
			}
			if self.PrintTypes {
				for _, decl := range file.Declarations {
					var t *Type
					t, ok := decl.(*Type)
					if !ok {
						continue
					}
					p("\t\tType\t\"%s\"\t(%s)\n", t.Name, t.Type)
				}
			}
			if self.PrintFuncs {
				for _, decl := range file.Declarations {
					var f *Func
					f, ok := decl.(*Func)
					if !ok {
						continue
					}
					p("\t\tFunc\t\"%s\"\n", f.Name)
					for _, tparam := range f.TypeParams {
						p("\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
					}
					for _, param := range f.Params {
						p("\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
					}
					for _, result := range f.Results {
						p("\t\t\tResult\t\"%s\"\t(%s)\n", result.Name, result.Type)
					}
				}
			}
			if self.PrintMethods {
				for _, decl := range file.Declarations {
					var m *Method
					m, ok := decl.(*Method)
					if !ok {
						continue
					}
					p("\t\tMethod\t\"%s\"\n", m.Name)
					for _, receiver := range m.Receivers {
						p("\t\t\tReceiver\t\"%s\"\t(%s)\n", receiver.Name, receiver.Type)
					}
					for _, tparam := range m.TypeParams {
						p("\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
					}
					for _, param := range m.Params {
						p("\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
					}
					for _, result := range m.Results {
						p("\t\t\tResult\t\"%s\"\t(%s)\n", result.Name, result.Type)
					}
				}
			}
			if self.PrintStructs {
				for _, decl := range file.Declarations {
					var s *Struct
					s, ok := decl.(*Struct)
					if !ok {
						continue
					}
					p("\t\tStruct\t\"%s\"\n", s.Name)
					for _, field := range s.Fields {
						p("\t\t\tField\t\"%s\"\t(%s)\t%s\n", field.Name, field.Type, field.Tag)
					}
				}
			}
			if self.PrintInterfaces {
				for _, decl := range file.Declarations {
					var i *Interface
					i, ok := decl.(*Interface)
					if !ok {
						continue
					}
					p("\t\tInterface\t\"%s\"\n", i.Name)
					for _, method := range i.Methods {
						p("\t\t\tMethod\t\"%s\"\n", method.Name)
						for _, receiver := range method.Receivers {
							p("\t\t\t\tReceiver\t\"%s\"\t(%s)\n", receiver.Name, receiver.Type)
						}
						for _, tparam := range method.TypeParams {
							p("\t\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
						}
						for _, param := range method.Params {
							p("\t\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
						}
						for _, result := range method.Results {
							p("\t\t\t\tResult\t\"%s\"\t(%s)\n", result.Name, result.Type)
						}
					}
				}
			}
		}

	}
	wr.Flush()
}

func printExpr(in ast.Expr) (out string) {
	if in == nil {
		return
	}
	switch e := in.(type) {
	case *ast.KeyValueExpr:
		return printKeyValueExpr(e)
	case *ast.CompositeLit:
		return printCompositeLit(e)
	case *ast.BasicLit:
		return printBasicLit(e)
	case *ast.Ident:
		return printIdent(e)
	case *ast.CallExpr:
		return printCallExpr(e)
	case *ast.UnaryExpr:
		return printUnaryExpr(e)
	case *ast.SelectorExpr:
		return printSelectorExpr(e)
	case *ast.ArrayType:
		return printArrayType(e)
	case *ast.StarExpr:
		return printStarExpr(e)
	case *ast.IndexExpr:
		return printIndexExpr(e)
	case *ast.FuncType:
		return printFuncType(e)
	case *ast.ChanType:
		return printChanType(e)
	case *ast.InterfaceType:
		return printInterfaceType(e)
	case *ast.StructType:
		return printStructType(e)
	case *ast.Ellipsis:
		return printEllipsis(e)
	case *ast.BinaryExpr:
		return printBinaryExpr(e)
	case *ast.MapType:
		return printMapType(e)
	case *ast.ParenExpr:
		return printParenExpr(e)
	case *ast.FuncLit:
		return printFuncLit(e)
	case *ast.SliceExpr:
		return printSliceExpr(e)
	default:
		panic("unsupported node")
	}
}

func printIdents(in []*ast.Ident) (out string) {
	var names []string
	for _, name := range in {
		names = append(names, printExpr(name))
	}
	return strings.Join(names, ", ")
}

func printIdent(in *ast.Ident) (out string) {
	return in.Name
}

func printBasicLit(in *ast.BasicLit) (out string) {
	if in == nil {
		return
	}
	return in.Value
}

func printFuncLit(in *ast.FuncLit) (out string) {
	out = "func"
	if in.Type.TypeParams != nil {
		out += " "
		out += "[" + printFieldList(in.Type.TypeParams) + "]"
	}
	if in.Type.Params != nil {
		if out == "func" {
			out += " "
		}
		out += "(" + printFieldList(in.Type.Params) + ")"
	}
	if in.Type.Results != nil {
		if len(in.Type.Results.List) > 1 {
			out += "("
		}
		out += " " + printFieldList(in.Type.Results)
		if len(in.Type.Results.List) > 1 {
			out += ")"
		}
	}
	// TODO printFuncLit: print body.
	return
}

func printCompositeLit(in *ast.CompositeLit) (out string) {
	// TODO Indent is fxd
	out = printExpr(in.Type) + "{"
	for _, elt := range in.Elts {
		out += "\n\t" + printExpr(elt)
	}
	out += "}"
	return
}

func printParenExpr(in *ast.ParenExpr) (out string) {
	return fmt.Sprintf("(%s)", printExpr(in.X))
}

func printUnaryExpr(in *ast.UnaryExpr) (out string) {
	return fmt.Sprintf("%s %s", in.Op.String(), printExpr(in.X))
}

func printBinaryExpr(in *ast.BinaryExpr) (out string) {
	return fmt.Sprintf("%s %s %s", printExpr(in.X), in.Op.String(), printExpr(in.Y))
}

func printSliceExpr(in *ast.SliceExpr) (out string) {
	if in == nil {
		return
	}
	if in.Slice3 {
		return fmt.Sprintf(
			"%s:%s:%s", printExpr(in.Low), printExpr(in.High), printExpr(in.Max),
		)
	}
	return fmt.Sprintf(
		"%s:%s", printExpr(in.Low), printExpr(in.High),
	)
}

func printSelectorExpr(in *ast.SelectorExpr) (out string) {
	return fmt.Sprintf("%s.%s", printExpr(in.X), printExpr(in.Sel))
}

func printKeyValueExpr(in *ast.KeyValueExpr) (out string) {
	return fmt.Sprintf("%s: %s,", printExpr(in.Key), printExpr(in.Value))
}

func printIndexExpr(in *ast.IndexExpr) (out string) {
	return fmt.Sprintf("%s[%s]", printExpr(in.X), printExpr(in.Index))
}

func printStarExpr(in *ast.StarExpr) (out string) {
	return "*" + printExpr(in.X)
}

func printCallExpr(in *ast.CallExpr) (out string) {
	var args []string
	out = printExpr(in.Fun)
	for _, arg := range in.Args {
		args = append(args, printExpr(arg))
	}
	out += "(" + strings.Join(args, ", ") + ")"
	return
}

func printEllipsis(in *ast.Ellipsis) (out string) {
	return "..." + printExpr(in.Elt)
}

func printField(in *ast.Field) (out string) {
	var names []string
	for _, name := range in.Names {
		names = append(names, printExpr(name))
	}
	out = strings.Join(names, ", ")
	if out != "" {
		out += " "
	}
	if in.Tag != nil {
		return out + printExpr(in.Type) + printExpr(in.Tag)
	}
	return out + printExpr(in.Type)
}

func printFieldList(in *ast.FieldList) (out string) {
	var names []string
	for _, field := range in.List {
		names = append(names, printField(field))
	}
	return strings.Join(names, ", ")
}

func printArrayType(in *ast.ArrayType) (out string) {
	return fmt.Sprintf("[%s]%s", printExpr(in.Len), printExpr(in.Elt))
}

func printChanType(in *ast.ChanType) (out string) {
	switch in.Dir {
	case ast.SEND:
		out = "<-chan"
	case ast.RECV:
		out = "chan<-"
	case ast.SEND | ast.RECV:
		out = "chan"
	}
	out += " " + printExpr(in.Value)
	return
}

func printFuncType(in *ast.FuncType) (out string) {
	out = "func"
	if in.TypeParams != nil {
		out += " "
		out += "[" + printFieldList(in.TypeParams) + "]"
	}
	if in.Params != nil {
		if out == "func" {
			out += " "
		}
		out += "(" + printFieldList(in.Params) + ")"
	}
	if in.Results != nil {
		out += " "
		if len(in.Results.List) > 1 {
			out += "("
		}
		out += printFieldList(in.Results)
		if len(in.Results.List) > 1 {
			out += ")"
		}
	}
	return
}

func printMapType(in *ast.MapType) (out string) {
	return fmt.Sprintf("map[%s]%s", printExpr(in.Key), printExpr(in.Value))
}

func printInterfaceType(in *ast.InterfaceType) (out string) {
	if len(in.Methods.List) > 0 {
		panic("interface has methods")
	}
	return "interface{}"
}

func printStructType(in *ast.StructType) (out string) {
	// TODO Newlines in struct printing
	out = "struct{"
	var (
		idx   int
		field *ast.Field
	)
	for idx, field = range in.Fields.List {
		out += "\n"
		out += fmt.Sprintf("\t%s\n", printField(field))
	}
	if idx > 0 {
		out += "\n"
	}
	out += "}"
	return
}
