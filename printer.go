// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// This file contains functions for printing bast.

package bast

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// Print prints bast to writer w using the default printer.
func Print(w io.Writer, bast *Bast) {
	DefaultPrinter().Print(w, bast)
}

// DefaultPrinter returns the default print configuration.
func DefaultPrinter() *Printer {
	return &Printer{
		PrintDoc:        true,
		PrintComments:   true,
		PrintConsts:     true,
		PrintVars:       true,
		PrintTypes:      true,
		PrintFuncs:      true,
		PrintMethods:    true,
		PrintStructs:    true,
		PrintInterfaces: true,
		Indentation:     "\t",
	}
}

// Printer holds the configuration for printing a Bast model.
type Printer struct {
	PrintDoc        bool
	PrintComments   bool
	PrintConsts     bool
	PrintVars       bool
	PrintTypes      bool
	PrintFuncs      bool
	PrintMethods    bool
	PrintStructs    bool
	PrintInterfaces bool
	Indentation     string
}

// Print prints the Bast model to the writer w.
func (self *Printer) Print(w io.Writer, bast *Bast) {
	var wr = tabwriter.NewWriter(w, 2, 2, 2, ' ', 0)
	for _, pkg := range bast.packages.Values() {
		self.printPackage(wr, pkg)
	}
	wr.Flush()
}

func (self *Printer) printPackage(w *tabwriter.Writer, pkg *Package) {
	fmt.Fprintf(w, "Package\t\"%s\"\t(%s)\n", pkg.Name, pkg.Path)
	for _, file := range pkg.Files.Values() {
		self.printFile(w, file, self.Indentation)
	}
}

func (self *Printer) printFile(w *tabwriter.Writer, file *File, indent string) {
	if self.PrintDoc {
		self.printDoc(w, file.Doc, indent)
	}
	fmt.Fprintf(w, "%sFile\t\"%s\"\n", indent, file.Name)
	if file.Imports.Len() > 0 {
		fmt.Fprintf(w, "%s%sImports\n", indent, self.Indentation)
		for _, key := range file.Imports.Keys() {
			var i, _ = file.Imports.Get(key)
			fmt.Fprintf(w, "%s%s%s\"%s\"\t(%s)\n", indent, self.Indentation, self.Indentation, i.Name, i.Path)
		}
	}

	for _, decl := range file.Declarations.Values() {
		switch d := decl.(type) {
		case *Const:
			if self.PrintConsts {
				self.printConst(w, d, indent+self.Indentation)
			}
		case *Var:
			if self.PrintVars {
				self.printVar(w, d, indent+self.Indentation)
			}
		case *Type:
			if self.PrintTypes {
				self.printType(w, d, indent+self.Indentation)
			}
		case *Func:
			if self.PrintFuncs {
				self.printFunc(w, d, indent+self.Indentation)
			}
		case *Method:
			if self.PrintMethods {
				self.printMethod(w, d, indent+self.Indentation)
			}
		case *Struct:
			if self.PrintStructs {
				self.printStruct(w, d, indent+self.Indentation)
			}
		case *Interface:
			if self.PrintInterfaces {
				self.printInterface(w, d, indent+self.Indentation)
			}
		}
	}
}

func (self *Printer) printConst(w *tabwriter.Writer, c *Const, indent string) {
	if self.PrintDoc {
		self.printDoc(w, c.Doc, indent)
	}
	fmt.Fprintf(w, "%sConst\t\"%s\"\t(%s)\t'%s'\n", indent, c.Name, c.Type, c.Value)
}

func (self *Printer) printVar(w *tabwriter.Writer, v *Var, indent string) {
	if self.PrintDoc {
		self.printDoc(w, v.Doc, indent)
	}
	fmt.Fprintf(w, "%sVar\t\"%s\"\t(%s)\t'%s'\n", indent, v.Name, v.Type, v.Value)
}

func (self *Printer) printType(w *tabwriter.Writer, t *Type, indent string) {
	if self.PrintDoc {
		self.printDoc(w, t.Doc, indent)
	}
	fmt.Fprintf(w, "%sType\t\"%s\"\t(%s)\n", indent, t.Name, t.Type)
	self.printFields(w, t.TypeParams, "Type Param", indent+self.Indentation)
}

func (self *Printer) printFunc(w *tabwriter.Writer, f *Func, indent string) {
	if self.PrintDoc {
		self.printDoc(w, f.Doc, indent)
	}
	fmt.Fprintf(w, "%sFunc\t\"%s\"\n", indent, f.Name)
	self.printFields(w, f.TypeParams, "Type Param", indent+self.Indentation)
	self.printFields(w, f.Params, "Param", indent+self.Indentation)
	self.printFields(w, f.Results, "Result", indent+self.Indentation)
}

func (self *Printer) printMethod(w *tabwriter.Writer, m *Method, indent string) {
	if self.PrintDoc {
		self.printDoc(w, m.Doc, indent)
	}
	fmt.Fprintf(w, "%sMethod\t\"%s\"\n", indent, m.Name)
	if m.Receiver != nil {
		if self.PrintDoc {
			self.printDoc(w, m.Receiver.Doc, indent+self.Indentation)
		}
		var receiverType = m.Receiver.Type
		if m.Receiver.Pointer {
			receiverType = "*" + receiverType
		}
		fmt.Fprintf(w, "%s%sReceiver\t\"%s\"\t(%s)\n", indent, self.Indentation, m.Receiver.Name, receiverType)
	}
	self.printFields(w, m.TypeParams, "Type Param", indent+self.Indentation)
	self.printFields(w, m.Params, "Param", indent+self.Indentation)
	self.printFields(w, m.Results, "Result", indent+self.Indentation)
}

func (self *Printer) printStruct(w *tabwriter.Writer, s *Struct, indent string) {
	if self.PrintDoc {
		self.printDoc(w, s.Doc, indent)
	}
	fmt.Fprintf(w, "%sStruct\t\"%s\"\n", indent, s.Name)
	for _, field := range s.Fields.Values() {
		if self.PrintDoc {
			self.printDoc(w, field.Doc, indent+self.Indentation)
		}
		fmt.Fprintf(w, "%s%sField\t\"%s\"\t(%s)\t%s\n", indent, self.Indentation, field.Name, field.Type, field.Tag)
	}
	self.printFields(w, s.TypeParams, "Type Param", indent+self.Indentation)
}

func (self *Printer) printInterface(w *tabwriter.Writer, i *Interface, indent string) {
	if self.PrintDoc {
		self.printDoc(w, i.Doc, indent)
	}
	fmt.Fprintf(w, "%sInterface\t\"%s\"\n", indent, i.Name)
	for _, method := range i.Methods.Values() {
		self.printMethod(w, method, indent+self.Indentation)
	}
	for _, intf := range i.Interfaces.Values() {
		self.printInterface(w, intf, indent+self.Indentation)
	}
}

func (self *Printer) printFields(w *tabwriter.Writer, fields *FieldMap, label, indent string) {
	for _, field := range fields.Values() {
		if self.PrintDoc {
			self.printDoc(w, field.Doc, indent)
		}
		fmt.Fprintf(w, "%s%s\t\"%s\"\t(%s)\n", indent, label, field.Name, field.Type)
	}
}

func (self *Printer) printDoc(w *tabwriter.Writer, doc []string, indent string) {
	for _, line := range doc {
		fmt.Fprintf(w, "%s%s\n", indent, line)
	}
}
