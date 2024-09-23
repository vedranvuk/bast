package main

import (
	"fmt"

	m "github.com/vedranvuk/bast/_testproject/pkg/models"
)

type Name string

var s = m.TestStruct2{}

func main() {
	_ = s
	fmt.Println("Hello from main")
}