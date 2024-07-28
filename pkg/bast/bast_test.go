package bast

import (
	"os"
	"testing"
)

func TestBast(t *testing.T) {
	var bast, err = LoadPackage(".")
	if err != nil {
		t.Fatal(err)
	}
	Print(os.Stdout, bast)
}