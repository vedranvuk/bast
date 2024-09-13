package bast

import (
	"fmt"
	"os"
	"testing"
)

func TestBast(t *testing.T) {
	var cfg = DefaultConfig()
	cfg.Dir = "_testproject"
	var bast, err = Load(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	Print(os.Stdout, bast)

	fmt.Printf(
		"Underlying type of CustomType is: %s\n",
		bast.ResolveBasicType("CustomType"),
	)
	fmt.Printf(
		"Underlying type of string is: %s\n",
		bast.ResolveBasicType("string"),
	)
	fmt.Printf(
		"Underlying type of PackageType is: %s\n",
		bast.ResolveBasicType("PackageType"),
	)
}

func TestImportSpecForTypeSelector(t *testing.T) {
	var cfg = DefaultConfig()
	cfg.Dir = "_testproject"
	var bast, err = Load(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	var typ = bast.AnyType("PackageType")
	var imp = typ.PackageImportBySelectorExpr(typ.Type)
	fmt.Println(imp.Path)
}
