package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/vedranvuk/bast"
)

func main() {
	if err := Test(); err != nil {
		fmt.Println(err)
	}
}

func Test() (err error) {
	var val *bast.Bast
	if val, err = bast.Load(
		"/usr/lib/go/src/fmt",
		"/usr/lib/go/src/go/ast",
		"/usr/lib/go/src/go/build",
		"/usr/lib/go/src/go/constant",
		"/usr/lib/go/src/go/format",
		"/usr/lib/go/src/go/importer",
		"/usr/lib/go/src/go/parser",
		"/usr/lib/go/src/go/printer",
		"/usr/lib/go/src/go/scanner",
		"/usr/lib/go/src/go/token",
		"/usr/lib/go/src/go/types",
		"/usr/lib/go/src/maps",
	); err != nil {
		return
	}

	var buf []byte
	if buf, err = os.ReadFile("basttest.tmpl"); err != nil {
		return err
	}

	var tmpl = template.New("basttest").Funcs(val.FuncMap())

	if tmpl, err = tmpl.Parse(string(buf)); err != nil {
		return err
	}

	if err = tmpl.Execute(os.Stdout, val); err != nil {
		return err
	}

	bast.Print(os.Stdout, val)
	return
}
