// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/vedranvuk/bast/pkg/bast"
)

func main() {
	if err := Test(); err != nil {
		fmt.Println(err)
	}
}

func Test() (err error) {
	var val *bast.Bast
	if val, err = bast.Load(
		// "github.com/vedranvuk/stringsex",
		// "../../pkg/bast",
		"../../pkg/bast/bast.go",
		// "/usr/lib/go/src/net/http",
		// "/usr/lib/go/src/archive/tar",
		// "/usr/lib/go/src/archive/zip",
		// "/usr/lib/go/src/fmt",
		// "/usr/lib/go/src/fmt",
		// "/usr/lib/go/src/maps",
	); err != nil {
		return
	}

	var buf []byte
	if buf, err = os.ReadFile("tester.tmpl"); err != nil {
		return err
	}

	var tmpl = template.New("tester").Funcs(val.FuncMap())

	if tmpl, err = tmpl.Parse(string(buf)); err != nil {
		return err
	}

	if err = tmpl.Execute(os.Stdout, val); err != nil {
		return err
	}

	var config = &bast.Config{
		PrintMethods: true,
	}

	config.Print(os.Stdout, val)
	return
}
