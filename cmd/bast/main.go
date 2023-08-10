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
	"strings"
)

var (
	ref         *bool
	output      *string
	inGoFiles   []string
	inTemplates []string
	variables   map[string]string
)

const usage = `Bast is a reference implementation for bast package.

It loads go inputs into a bast model and passes it to specified templates as it
executes them to output directory. Additional variables can be defined to be
made available to templates when executed.

The -v option defines a variable in the format '-v key=value'. it can be 
repeated to define multiple variables. Duplicate keys are not allowed.

The -i option takes a path to a go file or a directory or go module path and 
loads the ast of given inputs into a bast model. Option can be specified 
multiple times.
`

func main() {
	var fset = flag.NewFlagSet("bast", flag.ExitOnError)
	{
		fset.Usage = func() {
			fmt.Print(usage)
			fmt.Println()
			fset.PrintDefaults()
		}
		fset.Func("v", "Define a variable.", parseVariable)
		fset.Func("i", "Go file or package input (file, dir, module path).", parseGoInput)
		fset.Func("t", "Template input (file or directory).", parseTemplateInput)
		output = fset.String("o", ".", "Output directory.")
		ref = fset.Bool("r", false, "Show BAST reference.")
	}
	fset.Parse(os.Args[1:])

	if *ref {
		printRef()
		os.Exit(0)
	}
	if len(inGoFiles) == 0 {
		fatal(errors.New("no go input files specified"))
	}
	if len(inTemplates) == 0 {
		fatal(errors.New("no template files specified"))
	}

}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func parseVariable(value string) error {
	var key, val, valid = strings.Cut(value, "=")
	if !valid {
		return errors.New("variable must be in key=value format")
	}
	if _, valid = variables[key]; valid {
		return fmt.Errorf("duplicate variable: %s", key)
	}
	variables[key] = val
	return nil
}

func parseGoInput(value string) error {
	inGoFiles = append(inGoFiles, value)
	return nil
}

func parseTemplateInput(value string) error {
	inTemplates = append(inTemplates, value)
	return nil
}

func printRef() {

}
