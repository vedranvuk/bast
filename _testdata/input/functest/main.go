package main

func main() {}

var Variable string = "Value"

const Constant = 42

type MyInt int

type MyArray [5]bool

// Interface defines an interface.
type Interface interface {
	MethodA()
	MethodB()
	MethodC()
}

type Struct struct {
	Name string `json="name" db="name"`
	Age  int
}

func (self Struct) Method(name string) error                                          {}
func (self *Struct) PointerMethod(name, surname string, age int) (ok bool, err error) {}

// Echo echoes input to output.
func Echo(in string) (out string) { return name }
