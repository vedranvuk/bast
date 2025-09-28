// Package errortest contains valid Go code for testing edge case parsing
package errortest

import (
	"fmt"
	"io"
	"unsafe"
)

// This file contains valid but complex Go constructs to test parser robustness

// Complex struct with all field types
type ComplexStruct struct {
	// Basic types
	IntField    int
	StringField string
	BoolField   bool
	
	// Pointers
	IntPtr    *int
	StringPtr *string
	
	// Arrays and slices
	IntArray  [10]int
	IntSlice  []int
	ByteSlice []byte
	
	// Maps
	StringMap map[string]int
	IntMap    map[int]string
	
	// Channels
	IntChan    chan int
	StringChan <-chan string
	SendChan   chan<- int
	
	// Functions
	SimpleFunc func()
	ComplexFunc func(int, string) (bool, error)
	
	// Interfaces
	ReaderField io.Reader
	WriterField io.Writer
	
	// Unsafe pointer
	UnsafePtr unsafe.Pointer
	
	// Embedded types
	io.Reader
	io.Writer
	
	// Anonymous structs
	Nested struct {
		InnerField string
		InnerInt   int
	}
	
	// Tagged fields
	TaggedField string `json:"tagged_field" xml:"tagged" db:"tagged_column"`
}

// Generic types with various constraints
type GenericContainer[T any] struct {
	Value T
	Items []T
	Map   map[string]T
}

type ConstrainedGeneric[T comparable] struct {
	Key   T
	Value string
	Set   map[T]bool
}

type MultiConstraint[T any, U comparable, V ~int | ~string] struct {
	First  T
	Second U
	Third  V
}

// Complex function signatures
func ComplexFunction(
	basicParam int,
	sliceParam []string,
	mapParam map[string]int,
	funcParam func(int) string,
	chanParam chan int,
	variadicParam ...interface{},
) (result string, count int, err error) {
	return "", 0, nil
}

// Generic function with complex constraints
func GenericFunction[T comparable, U ~int | ~string](
	param1 T,
	param2 U,
	param3 []T,
) (T, U, error) {
	var zero T
	var zeroU U
	return zero, zeroU, nil
}

// Methods with complex receivers
func (c *ComplexStruct) PointerMethod() {}
func (c ComplexStruct) ValueMethod() {}

// Methods on generic types
func (g *GenericContainer[T]) Add(item T) {
	g.Items = append(g.Items, item)
}

func (g *GenericContainer[T]) Get(index int) T {
	if index < len(g.Items) {
		return g.Items[index]
	}
	var zero T
	return zero
}

// Complex interfaces
type ComplexInterface interface {
	BasicMethod()
	MethodWithParams(int, string) error
	MethodWithResults() (int, error)
	MethodWithInterface(interface{}) interface{}
}

type EmbeddedInterface interface {
	io.Reader
	io.Writer
	ComplexInterface
	
	AdditionalMethod() string
}

// Type aliases and definitions
type (
	StringAlias = string
	IntType     int
	SliceType   []string
	MapType     map[string]int
	FuncType    func(int) string
	ChanType    chan string
)

// Constants with various types
const (
	StringConst   = "test"
	IntConst      = 42
	FloatConst    = 3.14
	BoolConst     = true
	ComplexConst  = 1 + 2i
	RuneConst     = 'A'
	
	// Iota constants
	IotaFirst = iota
	IotaSecond
	IotaThird = iota + 10
)

// Variables with complex initialization
var (
	GlobalInt     int = 42
	GlobalString  string = "test"
	GlobalSlice   = []int{1, 2, 3}
	GlobalMap     = map[string]int{"key": 42}
	GlobalStruct  = ComplexStruct{IntField: 1, StringField: "test"}
	GlobalFunc    = func(x int) int { return x * 2 }
	GlobalChan    = make(chan int, 10)
	GlobalPointer = &GlobalInt
)

// Functions with all possible parameter/return combinations
func NoParamsNoReturn() {}
func WithParamsNoReturn(x int, y string) {}
func NoParamsWithReturn() int { return 0 }
func WithParamsWithReturn(x int) (int, error) { return x, nil }
func NamedReturns(x int) (result int, err error) { return x, nil }
func VariadicFunction(base string, args ...interface{}) string {
	return fmt.Sprintf(base, args...)
}

// Init function
func init() {
	GlobalInt = 100
	GlobalString = "initialized"
}
