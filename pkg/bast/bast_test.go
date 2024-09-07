package bast

import (
	"os"
	"testing"
)

func TestBast(t *testing.T) {
	var cfg = DefaultParseConfig()
	cfg.Dir = "./../../_testproject"
	var bast, err = ParsePackages(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	Print(os.Stdout, bast)
}
