package bast

import (
	"fmt"
	"os"
	"testing"
)

func TestBast(t *testing.T) {
	var cfg = DefaultParseConfig()
	cfg.ParsedElements = NoElements.Add(Structs)
	// cfg.ParsedElements = AllElements.Remove(Types, Funcs)

	
	cfg.Dir = "./../../_testproject"
	var bast, err = ParsePackages(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	Print(os.Stdout, bast)

	fmt.Printf(
		"Underlying type of CustomType is: %s\n",
		bast.ResolveBasicType("CustomType"),
	)
}
