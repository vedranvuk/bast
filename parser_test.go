package bast

import (
	"strings"
	"testing"
	"golang.org/x/tools/go/packages"
)

// TestParserInternals tests the internal parser functionality
func TestParserInternals(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	
	// Create parser
	parser := NewParser(cfg)
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}
	
	if parser.config != cfg {
		t.Error("Expected parser config to match input config")
	}
}

// TestParseWithErrors tests parsing behavior when errors occur
func TestParseWithErrors(t *testing.T) {
	// Test with type checking errors disabled
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	cfg.TypeCheckingErrors = false
	
	// Load packages that might have errors
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Should not error with TypeCheckingErrors=false: %v", err)
	}
	if bast == nil {
		t.Fatal("Expected valid bast even with potential errors")
	}
}

// TestParserFieldParsing tests detailed field parsing scenarios
func TestParserFieldParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	t.Run("EmbeddedFields", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct3")
		if s == nil {
			t.Fatal("Expected to find TestStruct3")
		}
		
		// Should have embedded TestStruct2
		embeddedField, ok := s.Fields.Get("TestStruct2")
		if !ok {
			t.Fatal("Expected embedded TestStruct2 field")
		}
		if embeddedField.Type != "TestStruct2" {
			t.Errorf("Expected embedded field type 'TestStruct2', got '%s'", embeddedField.Type)
		}
		if embeddedField.Name != "TestStruct2" {
			t.Errorf("Expected embedded field name 'TestStruct2', got '%s'", embeddedField.Name)
		}
	})

	t.Run("FieldTags", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct2")
		if s == nil {
			t.Fatal("Expected to find TestStruct2")
		}
		
		// Test field with tag
		barField, ok := s.Fields.Get("BarField")
		if !ok {
			t.Fatal("Expected BarField")
		}
		expectedTag := "`tag:\"value\"`"
		if barField.Tag != expectedTag {
			t.Errorf("Expected field tag '%s', got '%s'", expectedTag, barField.Tag)
		}
	})

	t.Run("UnnamedFields", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct2")
		if s == nil {
			t.Fatal("Expected to find TestStruct2")
		}
		
		// CustomType should be unnamed embedded field
		customTypeField, ok := s.Fields.Get("CustomType")
		if !ok {
			t.Fatal("Expected CustomType field")
		}
		if !customTypeField.Unnamed {
			t.Error("Expected CustomType field to be unnamed")
		}
	})
}

// TestParserFunctionParsing tests function parsing edge cases
func TestParserFunctionParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	testCases := []struct {
		funcName        string
		expectedParams  int
		expectedResults int
		hasTypeParams   bool
	}{
		{"TestFunc1", 0, 0, false},
		{"TestFunc2", 0, 1, false},
		{"TestFunc3", 0, 2, false},
		{"TestFunc4", 0, 1, false},
		{"TestFunc5", 0, 2, false},
		{"TestFunc6", 0, 3, false}, // Multiple named results of same type are now parsed correctly as separate fields
		{"TestFunc7", 1, 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.funcName, func(t *testing.T) {
			f := bast.AnyFunc(tc.funcName)
			if f == nil {
				t.Fatalf("Expected to find function '%s'", tc.funcName)
			}
			
			if f.Params.Len() != tc.expectedParams {
				t.Errorf("Expected %d params, got %d", tc.expectedParams, f.Params.Len())
			}
			
			if f.Results.Len() != tc.expectedResults {
				t.Errorf("Expected %d results, got %d", tc.expectedResults, f.Results.Len())
			}
			
			hasTypeParams := f.TypeParams.Len() > 0
			if hasTypeParams != tc.hasTypeParams {
				t.Errorf("Expected hasTypeParams=%v, got %v", tc.hasTypeParams, hasTypeParams)
			}
		})
	}
}

// TestParserMethodParsing tests method parsing with receivers
func TestParserMethodParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	testCases := []struct {
		methodName     string
		receiverType   string
		isPointer      bool
		hasResults     bool
	}{
		{"TestMethod1", "TestStruct1", false, false},
		{"TestMethod2", "TestStruct1", true, false},
		{"TestMethod3", "TestStruct4", false, true},
		{"TestMethod4", "TestStruct4", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.methodName, func(t *testing.T) {
			m := bast.AnyMethod(tc.methodName)
			if m == nil {
				t.Fatalf("Expected to find method '%s'", tc.methodName)
			}
			
			if m.Receiver == nil {
				t.Fatal("Expected method to have receiver")
			}
			
			if m.Receiver.Type != tc.receiverType {
				t.Errorf("Expected receiver type '%s', got '%s'", tc.receiverType, m.Receiver.Type)
			}
			
			if m.Receiver.Pointer != tc.isPointer {
				t.Errorf("Expected receiver pointer=%v, got %v", tc.isPointer, m.Receiver.Pointer)
			}
			
			hasResults := m.Results.Len() > 0
			if hasResults != tc.hasResults {
				t.Errorf("Expected hasResults=%v, got %v", tc.hasResults, hasResults)
			}
		})
	}
}

// TestParserInterfaceParsing tests interface parsing including embedded interfaces
func TestParserInterfaceParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	t.Run("EmptyInterface", func(t *testing.T) {
		i := bast.AnyInterface("Interface1")
		if i == nil {
			t.Fatal("Expected to find Interface1")
		}
		
		if i.Methods.Len() > 0 {
			t.Error("Expected empty interface to have no methods")
		}
		if i.Interfaces.Len() > 0 {
			t.Error("Expected empty interface to have no embedded interfaces")
		}
	})

	t.Run("InterfaceWithMethods", func(t *testing.T) {
		i := bast.AnyInterface("Interface2")
		if i == nil {
			t.Fatal("Expected to find Interface2")
		}
		
		if i.Methods.Len() != 1 {
			t.Errorf("Expected 1 method, got %d", i.Methods.Len())
		}
		
		method, ok := i.Methods.Get("IntfMethod1")
		if !ok {
			t.Fatal("Expected IntfMethod1")
		}
		if method.Receiver != nil {
			t.Error("Interface methods should not have receivers")
		}
	})

	t.Run("EmbeddedInterface", func(t *testing.T) {
		i := bast.AnyInterface("Interface3")
		if i == nil {
			t.Fatal("Expected to find Interface3")
		}
		
		// Should embed Interface2
		_, hasEmbedded := i.Interfaces.Get("Interface2")
		if !hasEmbedded {
			t.Error("Expected Interface3 to embed Interface2")
		}
		
		// Should have its own method too
		_, hasOwnMethod := i.Methods.Get("IntfMethod2")
		if !hasOwnMethod {
			t.Error("Expected Interface3 to have IntfMethod2")
		}
	})
}

// TestParserTypeParsing tests type declaration parsing
func TestParserTypeParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	testCases := []struct {
		typeName     string
		expectedType string
		hasTypeParams bool
	}{
		{"CustomType", "int", false},
		{"ParametrisedType", "int", true},
		{"PackageType", "types.ID", false},
	}

	for _, tc := range testCases {
		t.Run(tc.typeName, func(t *testing.T) {
			ty := bast.AnyType(tc.typeName)
			if ty == nil {
				t.Fatalf("Expected to find type '%s'", tc.typeName)
			}
			
			if ty.Type != tc.expectedType {
				t.Errorf("Expected underlying type '%s', got '%s'", tc.expectedType, ty.Type)
			}
			
			hasTypeParams := ty.TypeParams.Len() > 0
			if hasTypeParams != tc.hasTypeParams {
				t.Errorf("Expected hasTypeParams=%v, got %v", tc.hasTypeParams, hasTypeParams)
			}
		})
	}
}

// TestParserConstAndVarParsing tests constant and variable parsing
func TestParserConstAndVarParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Test variables
	t.Run("Variables", func(t *testing.T) {
		testCases := []struct {
			varName      string
			expectedType string
			expectedValue string
		}{
			{"a", "", "0"}, // implicit type
			{"b", "int", "1"}, // explicit type with value
			{"c", "int", ""}, // explicit type without value
		}

		for _, tc := range testCases {
			t.Run(tc.varName, func(t *testing.T) {
				v := bast.AnyVar(tc.varName)
				if v == nil {
					t.Fatalf("Expected to find variable '%s'", tc.varName)
				}
				
				if v.Type != tc.expectedType {
					t.Errorf("Expected type '%s', got '%s'", tc.expectedType, v.Type)
				}
				
				if v.Value != tc.expectedValue {
					t.Errorf("Expected value '%s', got '%s'", tc.expectedValue, v.Value)
				}
			})
		}
	})

	// Test constants
	t.Run("Constants", func(t *testing.T) {
		testCases := []struct {
			constName     string
			expectedType  string
			expectedValue string
		}{
			{"d", "", "0"}, // implicit type
			{"e", "int", "1"}, // explicit type
		}

		for _, tc := range testCases {
			t.Run(tc.constName, func(t *testing.T) {
				c := bast.AnyConst(tc.constName)
				if c == nil {
					t.Fatalf("Expected to find constant '%s'", tc.constName)
				}
				
				if c.Type != tc.expectedType {
					t.Errorf("Expected type '%s', got '%s'", tc.expectedType, c.Type)
				}
				
				if c.Value != tc.expectedValue {
					t.Errorf("Expected value '%s', got '%s'", tc.expectedValue, c.Value)
				}
			})
		}
	})
}

// TestModelMethods tests the Model struct methods
func TestModelMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	t.Run("GetFile", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct1")
		if s == nil {
			t.Fatal("Expected to find TestStruct1")
		}
		
		file := s.GetFile()
		if file == nil {
			t.Fatal("Expected struct to have a file")
		}
		
		if !strings.HasSuffix(file.Name, "models.go") {
			t.Errorf("Expected file name to end with 'models.go', got '%s'", file.Name)
		}
	})

	t.Run("GetPackage", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct1")
		if s == nil {
			t.Fatal("Expected to find TestStruct1")
		}
		
		pkg := s.GetPackage()
		if pkg == nil {
			t.Fatal("Expected struct to have a package")
		}
		
		if pkg.Name != "models" {
			t.Errorf("Expected package name 'models', got '%s'", pkg.Name)
		}
	})

	t.Run("ImportSpecBySelectorExpr", func(t *testing.T) {
		// Find PackageType which uses types.ID
		ty := bast.AnyType("PackageType")
		if ty == nil {
			t.Fatal("Expected to find PackageType")
		}
		
		// Test valid selector
		imp := ty.ImportSpecBySelectorExpr("types.ID")
		if imp == nil {
			t.Fatal("Expected to find import spec for types.ID")
		}
		if !strings.Contains(imp.Path, "/pkg/types") {
			t.Errorf("Expected import path to contain '/pkg/types', got '%s'", imp.Path)
		}
		
		// Test invalid selectors
		testCases := []string{
			"invalid", // no dot
			"", // empty
			".", // just dot
			"nonexistent.Type", // non-existent package
		}
		
		for _, selector := range testCases {
			imp := ty.ImportSpecBySelectorExpr(selector)
			if imp != nil {
				t.Errorf("Expected nil for selector '%s', got %v", selector, imp)
			}
		}
	})

	t.Run("ResolveBasicType", func(t *testing.T) {
		ty := bast.AnyType("CustomType")
		if ty == nil {
			t.Fatal("Expected to find CustomType")
		}
		
		// Test basic type resolution
		resolved := ty.ResolveBasicType("int")
		if resolved != "int" {
			t.Errorf("Expected 'int', got '%s'", resolved)
		}
		
		// Test custom type resolution
		resolved = ty.ResolveBasicType("CustomType")
		if resolved != "int" {
			t.Errorf("Expected 'int' for CustomType resolution, got '%s'", resolved)
		}
		
		// Test qualified type resolution
		resolved = ty.ResolveBasicType("types.ID")
		if resolved != "int" {
			t.Errorf("Expected 'int' for types.ID resolution, got '%s'", resolved)
		}
		
		// Test non-existent type
		resolved = ty.ResolveBasicType("NonExistentType")
		if resolved != "" {
			t.Errorf("Expected empty string for non-existent type, got '%s'", resolved)
		}
	})
}

// TestMethodSet tests the MethodSet functionality
func TestMethodSet(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
	
	t.Run("TestStruct1Methods", func(t *testing.T) {
		methods := bast.MethodSet(modelsPath, "TestStruct1")
		if len(methods) != 2 {
			t.Errorf("Expected 2 methods for TestStruct1, got %d", len(methods))
		}
		
		methodNames := make(map[string]bool)
		for _, m := range methods {
			methodNames[m.Name] = true
		}
		
		expectedMethods := []string{"TestMethod1", "TestMethod2"}
		for _, expected := range expectedMethods {
			if !methodNames[expected] {
				t.Errorf("Expected method '%s' not found", expected)
			}
		}
	})

	t.Run("TestStruct4Methods", func(t *testing.T) {
		methods := bast.MethodSet(modelsPath, "TestStruct4")
		if len(methods) != 2 {
			t.Errorf("Expected 2 methods for TestStruct4, got %d", len(methods))
		}
		
		// Test that pointer methods are also included when searching for value type
		foundPointerMethod := false
		for _, m := range methods {
			if m.Receiver.Pointer {
				foundPointerMethod = true
				break
			}
		}
		if !foundPointerMethod {
			t.Error("Expected to find pointer method when searching for TestStruct4")
		}
	})

	t.Run("NonExistentType", func(t *testing.T) {
		methods := bast.MethodSet(modelsPath, "NonExistentType")
		if len(methods) != 0 {
			t.Errorf("Expected 0 methods for non-existent type, got %d", len(methods))
		}
	})

	t.Run("NonExistentPackage", func(t *testing.T) {
		methods := bast.MethodSet("non/existent/package", "TestStruct1")
		if len(methods) != 0 {
			t.Errorf("Expected 0 methods for non-existent package, got %d", len(methods))
		}
	})
}

// TestVarsOfType and related methods
func TestTypeFiltering(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
	
	t.Run("VarsOfType", func(t *testing.T) {
		// Test int vars
		intVars := bast.VarsOfType(modelsPath, "int")
		if len(intVars) < 2 { // b and c at minimum
			t.Errorf("Expected at least 2 int vars, got %d", len(intVars))
		}
		
		// Verify all are int type
		for _, v := range intVars {
			if v.Type != "int" {
				t.Errorf("Expected var type 'int', got '%s'", v.Type)
			}
		}
	})

	t.Run("ConstsOfType", func(t *testing.T) {
		// Test int constants
		intConsts := bast.ConstsOfType(modelsPath, "int")
		if len(intConsts) < 1 { // e at minimum
			t.Errorf("Expected at least 1 int const, got %d", len(intConsts))
		}
	})

	t.Run("TypesOfType", func(t *testing.T) {
		// Test types based on int
		intTypes := bast.TypesOfType(modelsPath, "int")
		if len(intTypes) < 1 { // CustomType at minimum
			t.Errorf("Expected at least 1 int-based type, got %d", len(intTypes))
		}
	})
}

// TestFieldNames tests the FieldNames method
func TestFieldNames(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"

	// Test with correct package path and struct name
	fieldNames := bast.FieldNames(modelsPath, "TestStruct2")
	
	expectedFields := []string{"CustomType", "NamedCustomType", "FooField", "BarField", "Baz", "Bat"}
	
	if len(fieldNames) < len(expectedFields) {
		t.Errorf("Expected at least %d field names, got %d: %v", len(expectedFields), len(fieldNames), fieldNames)
	}
}

// TestParser tests the Parser struct methods directly
func TestParser(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	
	parser := NewParser(cfg)
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}

	// Load packages using go/packages
	mode := packages.NeedSyntax | packages.NeedCompiledGoFiles | packages.NeedName
	if cfg.TypeChecking {
		mode |= packages.NeedTypes | packages.NeedDeps | packages.NeedImports
	}

	pkgCfg := &packages.Config{
		Mode: mode,
		Dir:  cfg.Dir,
		Tests: cfg.Tests,
	}

	pkgs, err := packages.Load(pkgCfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	// Parse using the parser
	bast, err := parser.Parse(pkgs)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if bast == nil {
		t.Fatal("Expected valid bast from parser")
	}

	// Verify basic functionality
	if len(bast.PackageNames()) == 0 {
		t.Error("Expected packages to be parsed")
	}
}

// TestLoad tests the main Load function with various edge cases
func TestLoadEdgeCases(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		// Should use default config
		bast, err := Load(nil, ".")
		if err != nil {
			// May fail due to no valid go files in current dir, but should not panic
			if !strings.Contains(err.Error(), "failed to load packages") {
				t.Errorf("Unexpected error type: %v", err)
			}
		} else if bast == nil {
			t.Error("Expected valid bast or error")
		}
	})

	t.Run("EmptyPatterns", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		bast, err := Load(cfg)
		
		// Should still try to load something or return appropriate error
		if err != nil {
			if !strings.Contains(err.Error(), "failed to load packages") &&
			   !strings.Contains(err.Error(), "no Go files") {
				t.Errorf("Unexpected error type: %v", err)
			}
		} else if bast == nil {
			t.Error("Expected valid bast or error")
		}
	})

	t.Run("MultiplePatterns", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		bast, err := Load(cfg, "./pkg/models", "./pkg/types")
		
		if err != nil {
			t.Fatalf("Failed to load multiple patterns: %v", err)
		}
		if bast == nil {
			t.Fatal("Expected valid bast")
		}

		// Should have loaded both packages
		packages := bast.PackageNames()
		expectedPackages := []string{"models", "types"}
		
		for _, expected := range expectedPackages {
			found := false
			for _, pkg := range packages {
				if pkg == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected package '%s' not found in: %v", expected, packages)
			}
		}
	})
}

// TestParserErrorCases tests parser behavior with packages that have errors
func TestParserErrorCases(t *testing.T) {
	t.Run("TypeCheckingErrorsDisabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		cfg.TypeCheckingErrors = false
		
		// Should succeed even if there are type checking errors
		bast, err := Load(cfg, "./...")
		if err != nil {
			t.Fatalf("Should not fail with TypeCheckingErrors=false: %v", err)
		}
		if bast == nil {
			t.Fatal("Expected valid bast")
		}
	})
	
	t.Run("TypeCheckingDisabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		cfg.TypeChecking = false
		cfg.TypeCheckingErrors = false
		
		bast, err := Load(cfg, "./...")
		if err != nil {
			t.Fatalf("Should not fail with TypeChecking=false: %v", err)
		}
		if bast == nil {
			t.Fatal("Expected valid bast")
		}
		
		// Type resolution should not work without type checking
		resolved := bast.ResolveBasicType("CustomType")
		if resolved != "" {
			t.Errorf("Expected empty string for type resolution without type checking, got '%s'", resolved)
		}
	})
}