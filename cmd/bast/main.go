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

const usage = `

For Bast function reference type 'bast -f'.
`

func main() {

	var (
		fset      = flag.NewFlagSet("bast", flag.ExitOnError)
		input     = fset.String("i", "", "Input template file name.")
		output    = fset.String("o", ".", "Output file name.")
		ref       = fset.Bool("f", false, "Show BAST reference.")
		overwrite = fset.Bool("w", false, "Overwrite conflicting files in output.")
		stdout    = fset.Bool("s", false, "Write output to stdout.")
		debug     = fset.Bool("d", false, "Print debug info.")
		goinput   []string
		vars      = make(map[string]string)
	)
	fset.Usage = func() {
		fmt.Printf("Usage: bast -i <input-filename> ...-g <go-input> [options].\n")
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
	if len(os.Args[1:]) == 0 {
		fset.Usage()
		os.Exit(0)
	}
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
			fatal(fmt.Errorf("target exists: %s", *output))
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
	if err = os.MkdirAll(filepath.Dir(*output), os.ModePerm); err != nil {
		fatal(fmt.Errorf("make output dir: %w", err))
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
	const text = `Bast is a reference command for Bastard AST.

It loads Go source into a bast model and passes it to an input file being
executed to output file using 'text/template'. Additional simple variables can 
be defined to be made available to a template when executed.

Top level data structure passed to a template will be in the format:

struct {
	// Map of variables defined on command line.
	// Example: {{.Vars "MyVariable"}}
	Vars map[string]string

	// Bast model of loaded go inputs.
	// Example: {{constsoftype "main" "MyEnum"}}
	// Example: {{structfields "models" "MyModel"}}
	// Example: {{.Bast "main" "main.go" "MyVar"}}
	Bast *Bast
}

-i is required and specifies the input template file. 

-g takes a path to a Go source file, directory containing multiple Go source 
files defining a package or command or a go module path and loads the ast of 
given inputs into a bast model. 

A directory or file is specified either by an absolute or a relative path
(starting with "/", "." or ".."). If a path has no prefix it is considered 
a module path. Parsing go source from a module path depends on current 
environment variables loaded by this cmd.

Input files (i.e. not directories or modules) are loaded in bast into an unnamed 
package so when addressing declarations from them use an empty string for the
name of the package.

-o specifies the output file name. It is optional and if omitted saves the 
output file to the current directory under the same base file name as the
input template file.

-v defines a variable in the format '-v key=value'. It can be 
repeated to define multiple variables. Duplicate keys are not allowed.

-w option disables checking if the output file exists and overwrites it without 
asking for confirmation.

-s option enables printing the output to stdout in addition to the output file.

-d option enables printing of additional debug info.


Functions:

Standard library functions
  Utilities
    call       Calls first parameter passing all the rest as arguments.
	           This function is not used in bast cmd as data passed to templates
			   contains no exported functions to call except GetName on
			   a declaration.
    len        Returns length of first param or 0 and an error if undefined.
    index      Indexes first param with dim indices defined by rest of params.
    slice      Slices first param with second and optionally third.
    print      Prints any params.
    printf     Printf's using first param as format, rest as args.
    println    Println's any params.
  Escaping   
    html        HTML escapes first param.
    js          JS escapes first param.
    urlquery    URLQuery escapes first param.
  Logic
    and    and of param 1 and 2
    or     or of param 1 and 2
    not    not of param 1 and 2
  Comparisons
    eq    true if param 1 == param 2
    ge    true if param 1 >= param 2
    gt    true if param 1  > param 2
    le    true if param 1 <= param 2
    lt    true if param 1  < param 2
    ne    true if param 1 != param 2
Bast functions
  Utilities
    varsoftype      Return variables from a package by type name.
    constsoftype    Return consts from a package by type name.
    fieldset        Return methods 
    fieldnames      Return names of struct fields from package
  Additional string utilities     
    trimpfx      remove 2nd param defining suffix from 1st param.
    trimsfx      remove 2nd param defining prefix from 1st param.
	lowercase    lowercases input.
	uppercase    uppercases input.
    join         join 2nd+ params using 1st param defining separator.
    repeat       repeat 1st param, delimiting with 2nd, 3rd param times.
  Additional utilities
    datefmt       Return formated now using go time formatting layout.
    dateutcfmt    Return formated nowutc using go time formatting layout.
Retrieve a declaration from package by name.
  Param 1 = package name, param 2 = declaration name.
    var          
    const        
    type         
    func         
    method       
    interface    
    struct       
  Retrieve all declarations from a package.
  Param 1 = package name.
    vars          
    consts        
    types         
    funcs         
    methods       
    interfaces    
    structs       
  Retrieve all declarations by kind.
  These functions take no params.
    allvars          
    allconsts        
    alltypes         
    allfuncs         
    allmethods       
    allinterfaces    
    allstructs


Structure:

	Bast {
		Packages: map[string]*Package {
			Files: map[string]*File {
				Declarations: map[string]Declaration {
					Var
					Const
					Type
					Func
					Method
					Interface
					Struct
				}
			}
		}
	}


Models:

// Declaration defines a top level declaration in a go source file.
type Declaration interface {
	GetName() string // Returns declaration name, the identifier.
}

// Package contians info about a Go package.
type Package struct {
	Name  string           // Package name.
	Files map[string]*File // Map of files by their name.
}

// File contians info about a Go source file.
type File struct {
	Comments     [][]string             // Comment groups.
	Doc          []string               // Documentation.
	Name         string                 // Name with ext without path.
	Imports      map[string]*Import     // Map of imports by import path.
	Declarations map[string]Declaration // Map of declarations by name.
}

// Import contians info about an import.
type Import struct {
	Comment []string // Comment.
	Doc     []string // Documentation.
	Name    string   // Optional custom import name including reserved ".".
	Path    string   // Import path.
}

// Func contains info about a function.
type Func struct {
	Comment    []string          // Comment.
	Doc        []string          // Documentation.
	Name       string            // Name.
	TypeParams map[string]*Field // Map of type parameters by name.
	Params     map[string]*Field // Map of params by param name.
	Results    map[string]*Field // Map of results by result name.
}

// Method contains info about a method.
type Method struct {
	Func                        // Embeds all Func fields.
	Receivers map[string]*Field // Map of receivers by receiver name.
}

// Const contains info about a constant.
type Const struct {
	Comment []string // Comments.
	Doc     []string // Documentation.
	Name    string   // Name.
	Type    string   // Optional type name.
	Value   string   // Optional value.
}

// Var contains info about a variable.
type Var struct {
	Comment []string // Comments.
	Doc     []string // Documentation.
	Name    string   // Name.
	Type    string   // Optional type.
	Value   string   // Optional initial value.
}

// Type contains info about a type.
type Type struct {
	Comment []string // Comments.
	Doc     []string // Documentation.
	Name    string   // Name.
	Type    string   // Underlying type, type derived from.
	IsAlias bool     // True if a type alias.
}

// Interface contains info about an interface.
type Interface struct {
	Comment []string           // Comments.
	Doc     []string           // Documentation.
	Name    string             // Name.
	Methods map[string]*Method // Map of methods by method name.
}

// Struct contains info about a struct.
type Struct struct {
	Comment []string          // Comments.
	Doc     []string          // Documentation.
	Name    string            // Name.
	Fields  map[string]*Field // Map of Fields by field name.
}

// Field contains info about a struct field, method receiver, or method or func
// type params, params or results.
type Field struct {
	Comment []string // Comments.
	Doc     []string // Documentation.
	Name    string   // Name.
	Type    string   // Type name.
	Tag     string   // Raw tag.
}
`
	fmt.Print(text)
}
