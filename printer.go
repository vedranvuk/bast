package bast

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
	"text/tabwriter"
)

type Config struct {
	PrintValues bool
}

func Print(w io.Writer, bast *Bast) {
	var wr = tabwriter.NewWriter(w, 2, 2, 2, 32, 0)
	var p = func(format string, args ...any) { fmt.Fprintf(wr, format, args...) }
	for _, pkg := range bast.Packages {
		p("Package\t\"%s\"\n", pkg.Name)
		for _, file := range pkg.Files {
			p("\tFile\t\"%s\"\n", file.Name)
			for _, decl := range file.Declarations {
				var c *Const
				c, ok := decl.(*Const)
				if !ok {
					continue
				}
				p("\t\tConst\t\"%s\"\t(%s)\t'%s'\n", c.Name, c.Type, c.Value)
			}
			for _, decl := range file.Declarations {
				var v *Var
				v, ok := decl.(*Var)
				if !ok {
					continue
				}
				p("\t\tVar\t\"%s\"\t(%s)\t'%s'\n", v.Name, v.Type, v.Value)
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

func printCompositeLit(in *ast.CompositeLit) (out string) {
	out = printExpr(in.Type) + "{"
	for _, elt := range in.Elts {
		out += "\n\t" + printExpr(elt)
	}
	out = "}"
	return
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

func printCallExpr(in *ast.CallExpr) (out string) {
	var args []string
	out = printExpr(in.Fun)
	for _, arg := range in.Args {
		args = append(args, printExpr(arg))
	}
	out += "(" + strings.Join(args, ", ") + ")"
	return
}

func printUnaryExpr(in *ast.UnaryExpr) (out string) {
	return fmt.Sprintf("%s %s", in.Op.String(), printExpr(in.X))
}

func printIdent(in *ast.Ident) (out string) {
	return in.Name
}

func printBasicLit(in *ast.BasicLit) (out string) {
	return in.Value
}

func printSelectorExpr(in *ast.SelectorExpr) (out string) {
	return fmt.Sprintf("%s.%s", printExpr(in.Sel), printExpr(in.X))
}

func printKeyValueExpr(in *ast.KeyValueExpr) (out string) {
	return fmt.Sprintf("%s: %s,", printExpr(in.Key), printExpr(in.Value))
}

func printEllipsis(in *ast.Ellipsis) (out string) {
	return "..." + printExpr(in.Elt)
}

func printInterfaceType(in *ast.InterfaceType) (out string) {
	if len(in.Methods.List) > 0 {
		panic("interface has methods")
	}
	return "interface{}"
}

func printParenExpr(in *ast.ParenExpr) (out string) {
	return fmt.Sprintf("(%s)", printExpr(in.X))
}

func printBinaryExpr(in *ast.BinaryExpr) (out string) {
	return fmt.Sprintf("%s %s %s", printExpr(in.X), in.Op.String(), printExpr(in.Y))
}

func printIndexExpr(in *ast.IndexExpr) (out string) {
	return fmt.Sprintf("%s[%s]", printExpr(in.X), printExpr(in.Index))
}

func printStarExpr(in *ast.StarExpr) (out string) {
	return "*" + printExpr(in.X)
}

func printArrayType(in *ast.ArrayType) (out string) {
	return fmt.Sprintf("[]%s", printExpr(in.Elt))
}

func printMapType(in *ast.MapType) (out string) {
	return fmt.Sprintf("map[%s]%s", printExpr(in.Key), printExpr(in.Value))
}

func printStructType(in *ast.StructType) (out string) {
	out = "{"
	for _, field := range in.Fields.List {
		out += "\n"
		out += fmt.Sprintf("\t%s\n", printField(field))
	}
	if out != "{" {
		out += "\n"
	}
	out += "}"
	return
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

func printIdents(in []*ast.Ident) (out string) {
	var names []string
	for _, name := range in {
		names = append(names, name.Name)
	}
	return strings.Join(names, ", ")
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
		if len(in.Results.List) > 1 {
			out += "("
		}
		out += " " + printFieldList(in.Results)
		if len(in.Results.List) > 1 {
			out += ")"
		}
	}
	return
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

func printFieldList(in *ast.FieldList) (out string) {
	for _, field := range in.List {
		out += printField(field)
	}
	return
}
