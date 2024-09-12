package main

import (
	"fmt"

	"github.com/vedranvuk/bast/_testproject/pkg/models"
)

type Name string

var s = models.TestStruct2{}

func main() {
	_ = s
	fmt.Println("Hello from main")
}