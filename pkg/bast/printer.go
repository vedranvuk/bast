// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains functions for printing bast.

package bast

// TODO Limit types that get printed.
// TODO Config and formatting.

import (
	"fmt"
	"io"
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