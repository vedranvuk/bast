package bast

import "testing"

func TestImportSpecForTypeSelector(t *testing.T) {
	var cfg = DefaultConfig()
	cfg.Dir = "_testproject"
	var bast, err = Load(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	_ = bast
	// bast.ImportSpecForTypeSelector("types.ID")
}