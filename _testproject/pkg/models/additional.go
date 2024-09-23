package models

var (
	// x is a var with implicit type.
	x = 1
	// y is a var with explicit type and a value.
	y int = 2
	// z is a var with explicit type.
	z int
)

const (
	// u is a const with implicit type.
	u = 1
	// w is a const with explicit type and a value.
	w int = 2
)

type (
	// GrouppedType is a type declared in a type group.
	GrouppedType = int

	// GrouppedStruct is a struc declared in a type group.
	GrouppedStruct struct {
		// Name is the name field.
		Name string
	}
)