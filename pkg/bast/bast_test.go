package bast

import (
	"os"
	"testing"
)

func TestBast(t *testing.T) {
	var bast, err = ParsePackage("./../../_testdata/pkg/test", nil)
	// var bast, err = ParsePackage(".", nil)
	if err != nil {
		t.Fatal(err)
	}
	Print(os.Stdout, bast)
}