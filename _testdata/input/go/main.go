package main

func main() {}

var (
	One   string
	Two          = "two"
	Three string = "three"
)

const (
	Jack string
	Jill        = "jill"
	John string = "john"
)

type Alphabet int

const (
	Alpha Alphabet = iota
	Beta
	Gamma
	Delta
)

func Print(text string) error { return nil }

type AnInterface interface {
	Print(text string) error
}

type MyStruct struct {
	Name string
	Age  int
}

func (self *MyStruct) Print() error { return nil }
