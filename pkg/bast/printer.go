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

// Print prints bas to writer w.
func Print(w io.Writer, bast *Bast) {
	DefaultPrintConfig().Print(w, bast)
}

// DefaultPrintConfig returns the default print configuration.
func DefaultPrintConfig() *PrintConfig {
	return &PrintConfig{true, true, true, true, true, true, true, true, true}
}

type PrintConfig struct {
	PrintDoc        bool
	PrintComments   bool
	PrintConsts     bool
	PrintVars       bool
	PrintTypes      bool
	PrintFuncs      bool
	PrintMethods    bool
	PrintStructs    bool
	PrintInterfaces bool
}

func (self *PrintConfig) Print(w io.Writer, bast *Bast) {
	var wr = tabwriter.NewWriter(w, 2, 2, 2, 32, 0)
	var p = func(format string, args ...any) { fmt.Fprintf(wr, format, args...) }
	var pl = func(p string, l []string) {
		for _, s := range l {
			fmt.Fprintf(wr, "%s%s\n", p, s)
		}
	}
	// var pls = func(format string, ls [][]string) {
	// 	for _, l := range ls {
	// 		pl(format, l)
	// 	}
	// }
	for _, pkg := range bast.Packages {
		p("Package\t\"%s\"\n", pkg.Name)
		for _, file := range pkg.Files {
			if self.PrintDoc {
				pl("\t", file.Doc)
			}
			p("\tFile\t\"%s\"\n", file.Name)
			if self.PrintConsts {
				for _, decl := range file.Declarations {
					var c *Const
					c, ok := decl.(*Const)
					if !ok {
						continue
					}
					if self.PrintDoc {
						pl("\t\t", c.Doc)
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
					if self.PrintDoc {
						pl("\t\t", v.Doc)
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
					if self.PrintDoc {
						pl("\t\t", t.Doc)
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
					if self.PrintDoc {
						pl("\t\t", f.Doc)
					}
					p("\t\tFunc\t\"%s\"\n", f.Name)
					for _, tparam := range f.TypeParams {
						if self.PrintDoc {
							pl("\t\t\t", tparam.Doc)
						}
						p("\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
					}
					for _, param := range f.Params {
						if self.PrintDoc {
							pl("\t\t\t", param.Doc)
						}
						p("\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
					}
					for _, result := range f.Results {
						if self.PrintDoc {
							pl("\t\t\t", result.Doc)
						}
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
					if self.PrintDoc {
						pl("\t\t", m.Doc)
					}
					p("\t\tMethod\t\"%s\"\n", m.Name)
					if m.Receiver != nil {
						if self.PrintDoc {
							pl("\t\t\t", m.Receiver.Doc)
							p("\t\t\tReceiver\t\"%s\"\t(%s)\n", m.Receiver.Name, m.Receiver.Type)
						}
					}
					for _, tparam := range m.TypeParams {
						if self.PrintDoc {
							pl("\t\t\t", tparam.Doc)
						}
						p("\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
					}
					for _, param := range m.Params {
						if self.PrintDoc {
							pl("\t\t\t", param.Doc)
						}
						p("\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
					}
					for _, result := range m.Results {
						if self.PrintDoc {
							pl("\t\t\t", result.Doc)
						}
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
					if self.PrintDoc {
						pl("\t\t", s.Doc)
					}
					p("\t\tStruct\t\"%s\"\n", s.Name)
					for _, field := range s.Fields {
						if self.PrintDoc {
							pl("\t\t\t", field.Doc)
						}
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
					if self.PrintDoc {
						pl("\t\t", i.Doc)
					}
					p("\t\tInterface\t\"%s\"\n", i.Name)
					for _, method := range i.Methods {
						if self.PrintDoc {
							pl("\t\t\t", method.Doc)
						}
						p("\t\t\tMethod\t\"%s\"\n", method.Name)
						if method.Receiver != nil {
							if self.PrintDoc {
								pl("\t\t\t\t", method.Receiver.Doc)
							}
							p("\t\t\t\tReceiver\t\"%s\"\t(%s)\n", method.Receiver.Name, method.Receiver.Type)
						}
						for _, tparam := range method.TypeParams {
							if self.PrintDoc {
								pl("\t\t\t\t", tparam.Doc)
							}
							p("\t\t\t\tType Param\t\"%s\"\t(%s)\n", tparam.Name, tparam.Type)
						}
						for _, param := range method.Params {
							if self.PrintDoc {
								pl("\t\t\t\t", param.Doc)
							}
							p("\t\t\t\tParam\t\"%s\"\t(%s)\n", param.Name, param.Type)
						}
						for _, result := range method.Results {
							if self.PrintDoc {
								pl("\t\t\t\t", result.Doc)
							}
							p("\t\t\t\tResult\t\"%s\"\t(%s)\n", result.Name, result.Type)
						}
					}
					for _, intf := range i.Interfaces {
						if self.PrintDoc {
							pl("\t\t\t", intf.Doc)
						}
						p("\t\t\tInterface\t\"%s\"\n", intf.Name)

					}
				}
			}
		}

	}
	wr.Flush()
}
