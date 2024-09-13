// Package comments go here.

// Package test description goes here.
package models

import "github.com/vedranvuk/bast/_testproject/pkg/types"

// a is a variable with implicit type and a value.
var a = 0

// b is a var with explicit type and a value.
var b int = 1

// c is a var with explicit type and no value.
var c int

// d is a constant with implicit type and a value.
const d = 0

// e is a constant with explicit type and a value.
const e int = 1

// StructVar is a variable of an inplace defined struct literal.
var StructVar = struct {
	Name string
}{
	Name: "bast",
}

// FuncType is a func type.
type FuncType[K any] func(K) bool

// TestFunc returns no values.
func TestFunc1() {}

// TestFunc2 returns a single unnamed value.
func TestFunc2() error { return nil }

// TestFunc3 returns two values
func TestFunc3() (int, error) { return 0, nil }

// TestFunc4 returns a single named value.
func TestFunc4() (err error) { return nil }

// TestFunc5 returns two named values.
func TestFunc5() (i int, err error) { return 0, nil }

// TestFunc6 returns three named values of same type.
func TestFunc6() (a, b, c int) { return 0, 1, 2 }

// TestFunc7 is a func with a type parameter.
func TestFunc7[T any](in int) (out int) { return 0 }

// CustomType is a custom type.
type CustomType int

// ParametrisedType is a custom type with type parameters.
type ParametrisedType[K any] int

// TestStruct is an empty struct.
type TestStruct1 struct{}

// TestStruct2 Ha sfields.
type TestStruct2 struct {
	// CustomType is an unnamed field of custom type.
	CustomType
	// NamedCustomType is a field of custom type with a name.
	NamedCustomType CustomType
	// FooField is a struct field.
	FooField string
	// BarField is also a field but has a tag.
	BarField int `tag:"value"`
	// Baz and Bat are inline and described by this single line of text.
	Baz, Bat int
}

// TestStruct3 embedds other structs.
type TestStruct3 struct {
	Description string
	TestStruct2
}

// TestStruct4 is a struct with a type parameter.
type TestStruct4[T any] struct{}

// TestMethod1 is a methd on TestStruct4 with a generic method.
func (self *TestStruct4[T]) TestMethod1() (out int) { return 0 }

// Interface1 is an empty interface.
type Interface1 interface{}

// Interface2 is an interface with a single method.
type Interface2 interface {
	// IntfMethod1 is a method in Interface2.
	IntfMethod1() string
}

// Interface3 is an interface that embeds Interface2.
type Interface3 interface {
	// Interface2 is the inherited interface.
	Interface2
	// IntfMethod2 is a method in Interface3.
	IntfMethod2(in int) (out bool)
}

// PackageType is a type whose parent is in another package.
type PackageType types.ID
