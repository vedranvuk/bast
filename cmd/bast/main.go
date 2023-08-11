// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Bast is a reference implementation of bast functionality.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vedranvuk/bast/pkg/bast"
)

const usage = `Bast is a reference command for Bastard AST.

It loads Go inputs into a bast model and passes it to specified file as it is
executed to output directory using 'text/template'. Additional variables can be
defined to be made available to template when executed.

The -v option defines a variable in the format '-v key=value'. It can be 
repeated to define multiple variables. Duplicate keys are not allowed.

The -g option takes a path to a Go file or a directory or go module path and 
loads the ast of given inputs into a bast model. Module paths depend on current
Go build environment. Option can be specified multiple times to load multiple
packages or files.

Placeholder package for go files outside of package will have an empty name.s

Data passed to the template will be in the format:

struct {
	Vars map[string]string
	Bast *Bast
}

For Bast function reference type 'bast -f'.
`

func main() {

	var (
		fset      = flag.NewFlagSet("bast", flag.ExitOnError)
		input     = fset.String("i", "", "Template input glob pattern.")
		output    = fset.String("o", ".", "Output directory or file.")
		ref       = fset.Bool("f", false, "Show BAST reference.")
		overwrite = fset.Bool("w", false, "Overwrite conflicting files in output.")
		stdout    = fset.Bool("s", false, "Write output to stdout.")
		debug     = fset.Bool("d", false, "Print debug info.")
		goinput   []string
		vars      = make(map[string]string)
	)
	fset.Usage = func() {
		fmt.Print(usage)
		fmt.Println()
		fset.PrintDefaults()
	}
	fset.Func("g", "Go file or package input (file, dir, module path).",
		func(value string) error {
			goinput = append(goinput, value)
			return nil
		})
	fset.Func("v", "Define a variable.",
		func(value string) error {
			var key, val, valid = strings.Cut(value, "=")
			if !valid {
				return errors.New("variable must be in key=value format")
			}
			if _, valid = vars[key]; valid {
				return fmt.Errorf("duplicate variable: %s", key)
			}
			vars[key] = val
			return nil
		})
	fset.Parse(os.Args[1:])
	if *ref {
		printRef()
		os.Exit(0)
	}
	if *input == "" {
		fatal(errors.New("no template specified"))
	}
	if len(goinput) == 0 {
		fatal(errors.New("no go input files specified"))
	}
	var (
		err  error
		data = struct {
			Vars map[string]string
			Bast *bast.Bast
		}{vars, nil}
	)
	if data.Bast, err = bast.Load(goinput...); err != nil {
		fatal(err)
	}
	if *debug {
		fmt.Println("Bast:")
		bast.Print(os.Stdout, data.Bast)
		fmt.Println()
	}
	if *input, err = filepath.Abs(*input); err != nil {
		fatal(err)
	}
	var buf []byte
	if buf, err = os.ReadFile(*input); err != nil {
		fatal(err)
	}
	var tmpl *template.Template
	if tmpl, err = template.New(filepath.Base(*input)).
		Funcs(data.Bast.FuncMap()).
		Parse(string(buf)); err != nil {
		fatal(err)
	}
	if *output, err = filepath.Abs(*output); err != nil {
		fatal(err)
	}
	if !*overwrite {
		if _, err = os.Stat(*output); err == nil {
			fatal(fmt.Errorf("target exists: %w", err))
		} else if !errors.Is(err, os.ErrNotExist) {
			fatal(fmt.Errorf("stat target: %w", err))
		}
	}
	if *debug {
		fmt.Printf("Execute '%s' to '%s'\n", *input, *output)
	}
	if *stdout {
		if err = tmpl.Execute(os.Stdout, data); err != nil {
			fatal(fmt.Errorf("execute template: %w", err))
		}
	}
	var outfile *os.File
	if outfile, err = os.OpenFile(*output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
		fatal(fmt.Errorf("open target file: %w", err))
	}
	defer outfile.Close()
	if err = tmpl.Execute(outfile, data); err != nil {
		fatal(fmt.Errorf("execute template: %w", err))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func printRef() {

}
