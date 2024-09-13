package bast

import "testing"

func TestImportSpecForTypeSelector(t *testing.T) {
	var cfg = DefaultConfig()
	cfg.Dir = "_testproject"
	var bast, err = Load(cfg, "./...")
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range bast.Packages() {
		for _, file := range pkg.Files.Values() {
			if decl, ok := file.Declarations.Get("PackageType"); ok {
				
			} else {
				t.Fatal("ImportSpecForTypeSelector: declaration not found")
			}
		}
	}
}
