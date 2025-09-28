// Package crosspkg tests cross-package type resolution and dependencies
package crosspkg

import (
	// Standard library imports
	"context"
	"fmt"
	"io"

	// Cross-package imports with various patterns
	"github.com/vedranvuk/bast/_testproject/pkg/types"
	"github.com/vedranvuk/bast/_testproject/pkg/models"
	"github.com/vedranvuk/bast/_testproject/pkg/edgecases"
	"github.com/vedranvuk/bast/_testproject/pkg/generics"

	// Aliased imports
	baseTypes "github.com/vedranvuk/bast/_testproject/pkg/types"
	modelsAlias "github.com/vedranvuk/bast/_testproject/pkg/models"
)

// Cross-package type usage in variables
var (
	// Basic cross-package type
	TypesID types.ID = 42
	
	// Aliased cross-package type
	AliasedID baseTypes.ID = 100
	
	// Struct from another package with explicit type
	TestStruct models.TestStruct2 = models.TestStruct2{
		FooField: "cross-package",
		BarField: 25,
	}
	
	// Aliased struct
	AliasedStruct modelsAlias.TestStruct1
	
	// Generic types from another package with explicit type
	GenericPair generics.Pair[types.ID, string] = generics.Pair[types.ID, string]{
		First:  types.ID(1),
		Second: "value",
	}
	
	// Complex nested cross-package types with explicit type
	GenericContainer generics.Container[types.ID, string] = generics.Container[types.ID, string]{
		Key:   types.ID(42),
		Value: "test",
	}
	
	// Cross-package embedded types
	EdgeCaseVar edgecases.EmbeddedStruct
)

// Cross-package type usage in constants
const (
	// Using cross-package type as constant type
	MaxID types.ID = 9999
	
	// String constants referencing cross-package values
	PackageName = "crosspkg"
)

// Cross-package type usage in type declarations
type (
	// Type alias to cross-package type
	LocalID = types.ID
	
	// Type declaration with cross-package underlying type  
	WrappedID types.ID
	
	// Struct with cross-package field types
	CrossPackageStruct struct {
		ID       types.ID
		AliasID  baseTypes.ID
		Model    models.TestStruct1
		Generic  generics.Pair[types.ID, string]
		Embedded edgecases.EmbeddedStruct
	}
	
	// Interface with cross-package method signatures
	CrossPackageInterface interface {
		GetID() types.ID
		SetID(types.ID)
		ProcessModel(models.TestStruct2) error
		GetGeneric() generics.Pair[types.ID, string]
	}
	
	// Generic type with cross-package constraints and types
	CrossGeneric[T any, U types.ID] struct {
		Key   U
		Value T
		Pair  generics.Pair[U, T]
	}
)

// Functions with cross-package parameters and returns
func ProcessID(id types.ID) types.ID {
	return id + 1
}

func ProcessAliasedID(id baseTypes.ID) baseTypes.ID {
	return id * 2
}

func ProcessModel(m models.TestStruct2) models.TestStruct1 {
	return models.TestStruct1{}
}

func ProcessGeneric(p generics.Pair[types.ID, string]) generics.Pair[string, types.ID] {
	return generics.Pair[string, types.ID]{
		First:  p.Second,
		Second: p.First,
	}
}

func CreateCrossGeneric[T any](value T, id types.ID) CrossGeneric[T, types.ID] {
	return CrossGeneric[T, types.ID]{
		Key:   id,
		Value: value,
		Pair:  generics.Pair[types.ID, T]{First: id, Second: value},
	}
}

// Methods on cross-package types (not possible, but methods using them)
func (c *CrossPackageStruct) UpdateID(newID types.ID) {
	c.ID = newID
	c.AliasID = baseTypes.ID(newID)
}

func (c *CrossPackageStruct) GetModel() models.TestStruct1 {
	return c.Model
}

func (c *CrossPackageStruct) SetGeneric(first types.ID, second string) {
	c.Generic = generics.Pair[types.ID, string]{
		First:  first,
		Second: second,
	}
}

// Implementation of cross-package interface
type CrossImplementation struct {
	id    types.ID
	model models.TestStruct2
	pair  generics.Pair[types.ID, string]
}

func (c *CrossImplementation) GetID() types.ID {
	return c.id
}

func (c *CrossImplementation) SetID(id types.ID) {
	c.id = id
}

func (c *CrossImplementation) ProcessModel(m models.TestStruct2) error {
	c.model = m
	return nil
}

func (c *CrossImplementation) GetGeneric() generics.Pair[types.ID, string] {
	return c.pair
}

// Functions that resolve types across packages
func ResolveTypeChain() interface{} {
	// This creates a complex type resolution chain:
	// crosspkg -> generics -> types
	var container = generics.Container[types.ID, int]{
		Key:   types.ID(42),
		Value: 100,
	}
	
	return container
}

// Complex cross-package generic instantiation
func ComplexGenericUsage() interface{} {
	// Using multiple generics from different packages
	node := generics.Node[types.ID]{
		Value: types.ID(1),
		Next: &generics.Node[types.ID]{
			Value: types.ID(2),
		},
	}
	
	// Transform using cross-package types
	pair := generics.Pair[types.ID, models.TestStruct1]{
		First:  types.ID(42),
		Second: models.TestStruct1{},
	}
	
	return struct {
		Node generics.Node[types.ID]
		Pair generics.Pair[types.ID, models.TestStruct1]
	}{
		Node: node,
		Pair: pair,
	}
}

// Context usage with cross-package types
func ProcessWithContext(ctx context.Context, id types.ID, model models.TestStruct2) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("Processing ID %d with model %+v\n", id, model)
		return nil
	}
}

// Interface satisfaction across packages
func ProcessReader(r io.Reader, id types.ID) ([]byte, types.ID, error) {
	data := make([]byte, 100)
	n, err := r.Read(data)
	return data[:n], id, err
}

// Variable declarations that test type resolution
var (
	// Complex nested type with cross-package dependencies and explicit type
	ComplexNested struct {
		IDs   []types.ID
		Pairs []generics.Pair[types.ID, models.TestStruct1]
		Maps  map[types.ID]models.TestStruct2
	} = struct {
		IDs   []types.ID
		Pairs []generics.Pair[types.ID, models.TestStruct1]
		Maps  map[types.ID]models.TestStruct2
	}{
		IDs:   []types.ID{1, 2, 3},
		Pairs: []generics.Pair[types.ID, models.TestStruct1]{{First: 1}},
		Maps:  map[types.ID]models.TestStruct2{1: {}},
	}
	
	// Function variable with cross-package signature
	ProcessFunc func(types.ID, models.TestStruct2) (generics.Pair[types.ID, string], error)
	
	// Channel with cross-package type
	IDChannel chan types.ID = make(chan types.ID, 10)
	
	// Interface variable
	Processor CrossPackageInterface = &CrossImplementation{}
)

// Init function with cross-package initialization
func init() {
	TypesID = types.ID(100)
	TestStruct.FooField = "initialized"
	
	ProcessFunc = func(id types.ID, model models.TestStruct2) (generics.Pair[types.ID, string], error) {
		return generics.Pair[types.ID, string]{
			First:  id,
			Second: model.FooField,
		}, nil
	}
}