// Package generics tests comprehensive generic type scenarios
package generics

import (
	"context"
	"io"
	"github.com/vedranvuk/bast/_testproject/pkg/types"
)

// Basic generic constraints
type (
	// Ordered represents types that can be ordered
	Ordered interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
			~float32 | ~float64 | ~string
	}

	// Addable represents types that support addition
	Addable interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
			~float32 | ~float64 | ~complex64 | ~complex128 | ~string
	}

	// SignedInteger represents signed integer types
	SignedInteger interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}
)

// Generic data structures
type (
	// Node represents a node in a linked list or tree
	Node[T any] struct {
		Value T
		Next  *Node[T]
		Children []*Node[T]
	}

	// Pair holds two values of potentially different types
	Pair[T, U any] struct {
		First  T
		Second U
	}

	// Triple holds three values
	Triple[T, U, V any] struct {
		First  T
		Second U
		Third  V
	}

	// Container with multiple constraints
	Container[T Ordered, U Addable] struct {
		Key   T
		Value U
		Items []Pair[T, U]
	}

	// SelfReferential generic type
	SelfRef[T comparable] struct {
		Value T
		Self  *SelfRef[T]
		Map   map[T]*SelfRef[T]
	}

	// Generic interface
	Processor[T any] interface {
		Process(T) (T, error)
		Batch([]T) ([]T, error)
		~struct{ Value T } | ~*struct{ Value T } // type constraint with approximation
	}

	// Interface with method constraints
	Comparable[T comparable] interface {
		Compare(T) int
		Equal(T) bool
	}

	// Complex nested generics  
	NestedContainer[T comparable, U Comparable[T]] struct {
		Items map[T]U
		ProcessorFuncs []func(T) T
	}
)

// Generic functions
func SimpleGeneric[T any](value T) T {
	return value
}

func GenericWithConstraint[T Ordered](a, b T) bool {
	return a < b // This won't compile without the constraint
}

func MultipleConstraints[T Ordered, U Addable](key T, value U) Pair[T, U] {
	return Pair[T, U]{First: key, Second: value}
}

func GenericWithTypeInference[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[0]
}

// Methods on generic types
func (n *Node[T]) Add(value T) *Node[T] {
	if n.Next == nil {
		n.Next = &Node[T]{Value: value}
		return n.Next
	}
	return n.Next.Add(value)
}

func (n *Node[T]) Find(value T, compare func(T, T) bool) *Node[T] {
	if compare(n.Value, value) {
		return n
	}
	if n.Next != nil {
		return n.Next.Find(value, compare)
	}
	return nil
}

func (p Pair[T, U]) Swap() Pair[U, T] {
	return Pair[U, T]{First: p.Second, Second: p.First}
}

func (c *Container[T, U]) Add(key T, value U) {
	c.Items = append(c.Items, Pair[T, U]{First: key, Second: value})
}

// Generic methods with additional type parameters
func Transform[T Ordered, U Addable, V any](c *Container[T, U], fn func(Pair[T, U]) V) []V {
	result := make([]V, len(c.Items))
	for i, item := range c.Items {
		result[i] = fn(item)
	}
	return result
}

// Type aliases with generics
type (
	StringIntPair = Pair[string, int]
	IntNode = Node[int]
	StringProcessor = Processor[string]
	IDNode = Node[types.ID] // Cross-package generic usage
)

// Generic type with embedded interface
type Reader[T any] struct {
	io.Reader
	data T
}

func (r *Reader[T]) ReadTyped() (T, error) {
	return r.data, nil
}

// Generic variadic functions
func Collect[T any](items ...T) []T {
	return items
}

func Process[T any](processor func(T) T, items ...T) []T {
	result := make([]T, len(items))
	for i, item := range items {
		result[i] = processor(item)
	}
	return result
}

// Context with generics
func ProcessWithContext[T any](ctx context.Context, data T, processor func(context.Context, T) (T, error)) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	default:
		return processor(ctx, data)
	}
}

// Higher-order generic functions
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Generic constants and variables
var (
	DefaultPair = Pair[string, int]{First: "default", Second: 0}
	EmptyNode   = &Node[interface{}]{}
)

const (
	DefaultCapacity = 100
)

// Complex type constraints
type (
	// Constraint with multiple type sets
	Number interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
			~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
			~float32 | ~float64
	}

	// Constraint with methods and types
	Stringer[T any] interface {
		String() string
		comparable
		*T // pointer constraint
	}

	// Recursive constraint - simplified
	TreeNode[T any] interface {
		GetChildren() []T
		GetParent() T
		SetParent(T)
	}
)

// Implementation of complex constraints
type MyInt int

func (m MyInt) String() string {
	return "MyInt"
}

type MyString string

func (m *MyString) String() string {
	return string(*m)
}

// Generic receiver methods
func GenericMethod[T Number](value T) T {
	return value + 1
}

// Test instantiation of complex types
var (
	IntStringPair     = Pair[int, string]{First: 1, Second: "one"}
	NodeTree          = Node[string]{Value: "root", Children: []*Node[string]{{Value: "child"}}}
	OrderedContainer  = Container[int, string]{Key: 1, Value: "test"}
	ProcessorInstance = Reader[types.ID]{data: types.ID(42)}
)