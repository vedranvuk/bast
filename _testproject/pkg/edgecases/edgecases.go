// Package edgecases contains various edge cases for testing the bast parser.package edgecases

package edgecases

import (
	"context"
	"io"
	"unsafe"
	
	. "fmt"
	_ "log"
	aliased "strings"
	
	"github.com/vedranvuk/bast/_testproject/pkg/types"
)

// Constants with various types and edge cases
const (
	// IotaConst uses iota
	IotaConst = iota
	IotaConst2
	IotaConst3 = iota + 1
	
	// StringConst with special characters
	StringConst = "test\nwith\"escapes\\"
	
	// ComplexConst is complex
	ComplexConst complex128 = 1 + 2i
)

// Variables with edge cases
var (
	// PointerVar is a pointer
	PointerVar *int = nil
	
	// ChannelVar is a channel
	ChannelVar chan string
	
	// MapVar is a map
	MapVar map[string]int = make(map[string]int)
	
	// SliceVar is a slice
	SliceVar []string
	
	// ArrayVar is an array
	ArrayVar [10]int
	
	// FunctionVar is a function variable
	FunctionVar func(int) string
	
	// InterfaceVar is an interface
	InterfaceVar io.Reader
	
	// UnsafePointer using unsafe package
	UnsafePointer unsafe.Pointer
)

// Function types with various signatures
type (
	// SimpleFunc is a simple function type
	SimpleFunc func()
	
	// ComplexFunc has complex signature
	ComplexFunc func(a int, b ...string) (result string, err error)
	
	// GenericFuncType is a generic function type
	GenericFuncType[T any, U comparable] func(T, U) bool
	
	// RecursiveType references itself
	RecursiveType *RecursiveType
)

// Generic types
type (
	// Container is a generic container
	Container[T any] struct {
		Value T
		Next  *Container[T]
	}
	
	// Pair holds two values
	Pair[T, U any] struct {
		First  T
		Second U
	}
	
	// Constrained uses type constraints
	Constrained[T comparable] struct {
		Key T
	}
)

// Embedded fields and anonymous types
type (
	// EmbeddedStruct has embedded fields
	EmbeddedStruct struct {
		io.Reader
		io.Writer
		string // anonymous field
		*types.ID // pointer to embedded type
		
		Name string `json:"name" yaml:"name" db:"name"`
	}
	
	// AnonymousStruct with anonymous struct field
	AnonymousStruct struct {
		Nested struct {
			Value int
		}
	}
)

// Interfaces with various edge cases
type (
	// EmptyInterface is empty
	EmptyInterface interface{}
	
	// BasicInterface has methods
	BasicInterface interface {
		Method1()
		Method2(int) string
	}
	
	// EmbeddedInterface embeds other interfaces
	EmbeddedInterface interface {
		io.Reader
		io.Writer
		BasicInterface
		Method3() error
	}
	
	// GenericInterface is generic
	GenericInterface[T any] interface {
		Process(T) T
		~int | ~string // type constraint
	}
	
	// ComplexInterface with type constraints
	ComplexInterface[T comparable, U any] interface {
		*T | ~string
		Compare(T, T) bool
		Transform(U) T
	}
)

// Functions with various signatures and edge cases

// NoParamNoReturn has no parameters or return values
func NoParamNoReturn() {}

// SingleParam has one parameter
func SingleParam(x int) {}

// MultipleParams has multiple parameters
func MultipleParams(a int, b string, c bool) {}

// VariadicParams has variadic parameters
func VariadicParams(prefix string, values ...int) {}

// SingleReturn returns one value
func SingleReturn() int { return 0 }

// MultipleReturn returns multiple values
func MultipleReturn() (int, string, error) { return 0, "", nil }

// NamedReturn has named return values
func NamedReturn() (result int, err error) { return 0, nil }

// MixedReturn has mixed named and unnamed returns
func MixedReturn() (result int, msg string, err error) { return 0, "", nil }

// GenericFunction is a generic function
func GenericFunction[T any](value T) T { return value }

// ComplexGenericFunc has complex generics
func ComplexGenericFunc[T comparable, U ~int | ~string](a T, b U) (T, U) {
	return a, b
}

// MethodWithComplexReceiver demonstrates complex receiver types
type ComplexReceiver[T any] struct {
	Value T
}

// ValueMethod has value receiver
func (cr ComplexReceiver[T]) ValueMethod() T {
	return cr.Value
}

// PointerMethod has pointer receiver
func (cr *ComplexReceiver[T]) PointerMethod(value T) {
	cr.Value = value
}

// Methods on basic types
type CustomInt int

func (ci CustomInt) String() string { return "custom" }
func (ci *CustomInt) Increment()    { *ci++ }

// Methods on slice types
type CustomSlice []int

func (cs CustomSlice) Len() int           { return len(cs) }
func (cs *CustomSlice) Append(v int)      { *cs = append(*cs, v) }

// Interface implementations
type StringReader struct {
	data string
}

func (sr *StringReader) Read(p []byte) (n int, err error) {
	n = copy(p, sr.data)
	sr.data = sr.data[n:]
	if n == 0 {
		err = io.EOF
	}
	return
}

// Type aliases
type (
	StringAlias = string
	IntAlias = int
	StructAlias = EmbeddedStruct
	GenericAlias[T any] = Container[T]
)

// Dot imports and aliased imports usage
var (
	PrinterFunc = Printf // from . "fmt"
	StringsFunc = aliased.Contains // from aliased "strings"
)

// Context usage
func ContextFunc(ctx context.Context) error {
	return ctx.Err()
}

// Anonymous functions and closures
var (
	AnonymousFunc = func(x int) int { return x * 2 }
	ClosureFunc   = func() func() int {
		count := 0
		return func() int {
			count++
			return count
		}
	}()
)

// Init function
func init() {
	// Initialization code
	SliceVar = make([]string, 0)
}