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

It loads Go inputs into a bast model and passes it to specified templates as it
executes them to output directory. Additional variables can be defined to be
made available to templates when executed.

The -v option defines a variable additionally available to templates in the
format '-v key=value'. It can be repeated to define multiple variables. 
Duplicate keys are not allowed.

The -i option takes a path to a Go file or a directory or go module path and 
loads the ast of given inputs into a bast model. Module paths depend on current
Go build environment. Option can be specified multiple times to load multiple
packages or files.

Data passed to the template will be in the format:

struct {
	Vars map[string]string
	Bast *Bast
}

For Bast function reference type 'bast -f'.
`

func main() {

	var (
		fset       = flag.NewFlagSet("bast", flag.ExitOnError)
		ref        *bool
		outputDir  *string
		overwrite  *bool
		print      *bool
		inGoFiles  []string
		inTemplate *string
		variables  = make(map[string]string)
	)
	{
		fset.Usage = func() {
			fmt.Print(usage)
			fmt.Println()
			fset.PrintDefaults()
		}
		fset.Func("v", "Define a variable.",
			func(value string) error {
				var key, val, valid = strings.Cut(value, "=")
				if !valid {
					return errors.New("variable must be in key=value format")
				}
				if _, valid = variables[key]; valid {
					return fmt.Errorf("duplicate variable: %s", key)
				}
				variables[key] = val
				return nil
			})
		fset.Func("i", "Go file or package input (file, dir, module path).",
			func(value string) error {
				inGoFiles = append(inGoFiles, value)
				return nil
			})
		ref = fset.Bool("f", false, "Show BAST reference.")
		print = fset.Bool("p", false, "Print commands instead of executing them.")
		outputDir = fset.String("o", ".", "Output directory.")
		overwrite = fset.Bool("w", false, "Overwrite conflicting files in output.")
		inTemplate = fset.String("t", "", "Template input glob pattern.")
	}
	fset.Parse(os.Args[1:])

	if *ref {
		printRef()
		os.Exit(0)
	}
	if len(inGoFiles) == 0 {
		fatal(errors.New("no go input files specified"))
	}
	if *inTemplate == "" {
		fatal(errors.New("no template specified"))
	}

	var (
		inputs []string
		output string
		file   *os.File
		tmpl   *template.Template
		buf    []byte
		err    error
		data   = struct {
			Vars map[string]string
			Bast *bast.Bast
		}{variables, nil}
	)

	if data.Bast, err = bast.Load(inGoFiles...); err != nil {
		fatal(err)
	}
	if *inTemplate, err = filepath.Abs(*inTemplate); err != nil {
		fatal(err)
	}
	if inputs, err = filepath.Glob(*inTemplate); err != nil {
		fatal(err)
	}
	for _, fn := range inputs {
		if buf, err = os.ReadFile(fn); err != nil {
			fatal(err)
		}
		if tmpl, err = template.New(filepath.Base(fn)).
			Funcs(data.Bast.FuncMap()).
			Parse(string(buf)); err != nil {
			fatal(err)
		}
		if output, err = filepath.Rel(filepath.Dir(*inTemplate), fn); err != nil {
			fatal(err)
		}
		if output, err = filepath.Abs(filepath.Join(*outputDir, output)); err != nil {
			fatal(err)
		}
		if !*overwrite {
			if _, err = os.Stat(output); err == nil {
				fatal(fmt.Errorf("target exists: %w", err))
			} else if !errors.Is(err, os.ErrNotExist) {
				fatal(fmt.Errorf("stat target: %w", err))
			}
		}
		if *print {
			fmt.Printf("Execute '%s' to '%s'\n", fn, output)
			continue
		}
		if file, err = os.OpenFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
			fatal(fmt.Errorf("open target file: %w", err))
		}
		defer file.Close()
		if err = tmpl.Execute(file, data); err != nil {
			fatal(fmt.Errorf("execute template: %w", err))
		}
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func printRef() {

}
