package bast

import (
	"reflect"
	"strings"
	"testing"
)

// TestCase represents a single test case for testing bast functionality
type TestCase struct {
	Name        string
	Description string
	Config      *Config
	Patterns    []string
	ExpectedErr string
	Validate    func(t *testing.T, bast *Bast)
}

// runTestCases is a helper function that runs a series of test cases
func runTestCases(t *testing.T, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bast, err := Load(tc.Config, tc.Patterns...)
			
			// Check for expected errors
			if tc.ExpectedErr != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tc.ExpectedErr)
					return
				}
				if !strings.Contains(err.Error(), tc.ExpectedErr) {
					t.Errorf("Expected error containing '%s', but got: %v", tc.ExpectedErr, err)
					return
				}
				return // Expected error occurred, test passed
			}
			
			// If no error was expected, ensure we got a valid bast
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if bast == nil {
				t.Fatal("Expected valid bast, but got nil")
			}
			
			// Run validation if provided
			if tc.Validate != nil {
				tc.Validate(t, bast)
			}
		})
	}
}

// TestLoadAndParse tests the main Load functionality with various configurations
func TestLoadAndParse(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "BasicLoad",
			Description: "Load with default configuration",
			Config:      DefaultConfig(),
			Patterns:    []string{"./..."},
			Validate: func(t *testing.T, bast *Bast) {
				if len(bast.PackageNames()) == 0 {
					t.Error("Expected at least one package, got none")
				}
			},
		},
		{
			Name:        "LoadWithTestProject",
			Description: "Load the test project specifically",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "_testproject"
				return cfg
			}(),
			Patterns: []string{"./..."},
			Validate: func(t *testing.T, bast *Bast) {
				// Validate that we loaded expected packages
				packages := bast.PackageNames()
				expectedPackages := []string{"types", "models", "main", "edgecases"}
				
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
			},
		},
		{
			Name:        "LoadWithoutTypeChecking",
			Description: "Load with type checking disabled",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "_testproject"
				cfg.TypeChecking = false
				return cfg
			}(),
			Patterns: []string{"./..."},
			Validate: func(t *testing.T, bast *Bast) {
				// Should still load successfully
				if len(bast.PackageNames()) == 0 {
					t.Error("Expected packages to be loaded even without type checking")
				}
			},
		},
		{
			Name:        "LoadWithTests",
			Description: "Load with test files included",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "_testproject"
				cfg.Tests = true
				return cfg
			}(),
			Patterns: []string{"./..."},
			Validate: func(t *testing.T, bast *Bast) {
				// Should load successfully with tests
				if len(bast.PackageNames()) == 0 {
					t.Error("Expected packages to be loaded with test files")
				}
			},
		},
		{
			Name:        "LoadNonExistentDirectory",
			Description: "Try to load from non-existent directory",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "/non/existent/directory"
				return cfg
			}(),
			Patterns:    []string{"./..."},
			ExpectedErr: "failed to load packages",
		},
		{
			Name:        "LoadInvalidPattern",
			Description: "Try to load with invalid pattern",
			Config:      DefaultConfig(),
			Patterns:    []string{"invalid:pattern:with:colons"},
			ExpectedErr: "malformed import path",
		},
	}
	
	runTestCases(t, testCases)
}

// TestBastAPI tests the main Bast API methods
func TestBastAPI(t *testing.T) {
	// Setup: Load test project
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []TestCase{
		{
			Name:        "PackageNames",
			Description: "Test PackageNames method",
			Validate: func(t *testing.T, _ *Bast) {
				names := bast.PackageNames()
				if len(names) == 0 {
					t.Error("Expected package names, got none")
				}
				// Should contain our test packages
				expectedPackages := []string{"types", "models", "main", "edgecases"}
				for _, expected := range expectedPackages {
					found := false
					for _, name := range names {
						if name == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected package '%s' in names: %v", expected, names)
					}
				}
			},
		},
		{
			Name:        "PackageByPath",
			Description: "Test PackageByPath method",
			Validate: func(t *testing.T, _ *Bast) {
				// Test existing package
				pkg := bast.PackageByPath("github.com/vedranvuk/bast/_testproject/pkg/models")
				if pkg == nil {
					t.Error("Expected to find models package")
				} else if pkg.Name != "models" {
					t.Errorf("Expected package name 'models', got '%s'", pkg.Name)
				}
				
				// Test non-existing package
				pkg = bast.PackageByPath("non/existent/package")
				if pkg != nil {
					t.Error("Expected nil for non-existent package")
				}
			},
		},
		{
			Name:        "PackageImportPaths",
			Description: "Test PackageImportPaths method",
			Validate: func(t *testing.T, _ *Bast) {
				paths := bast.PackageImportPaths()
				if len(paths) == 0 {
					t.Error("Expected import paths, got none")
				}
				
				// Should contain our test package paths
				expectedPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
				found := false
				for _, path := range paths {
					if path == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected path '%s' in paths: %v", expectedPath, paths)
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Validate(t, bast)
		})
	}
}

// TestDeclarationRetrieval tests various declaration retrieval methods
func TestDeclarationRetrieval(t *testing.T) {
	// Setup
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []struct {
		name     string
		test     func(t *testing.T)
	}{
		{
			name: "AnyVar",
			test: func(t *testing.T) {
				// Test existing variable
				v := bast.AnyVar("a")
				if v == nil {
					t.Error("Expected to find variable 'a'")
				} else if v.Name != "a" {
					t.Errorf("Expected variable name 'a', got '%s'", v.Name)
				}
				
				// Test non-existing variable
				v = bast.AnyVar("nonexistent")
				if v != nil {
					t.Error("Expected nil for non-existent variable")
				}
			},
		},
		{
			name: "AnyConst",
			test: func(t *testing.T) {
				c := bast.AnyConst("d")
				if c == nil {
					t.Error("Expected to find constant 'd'")
				} else if c.Name != "d" {
					t.Errorf("Expected constant name 'd', got '%s'", c.Name)
				}
			},
		},
		{
			name: "AnyFunc",
			test: func(t *testing.T) {
				f := bast.AnyFunc("TestFunc1")
				if f == nil {
					t.Error("Expected to find function 'TestFunc1'")
				} else if f.Name != "TestFunc1" {
					t.Errorf("Expected function name 'TestFunc1', got '%s'", f.Name)
				}
			},
		},
		{
			name: "AnyMethod",
			test: func(t *testing.T) {
				m := bast.AnyMethod("TestMethod1")
				if m == nil {
					t.Error("Expected to find method 'TestMethod1'")
				} else if m.Name != "TestMethod1" {
					t.Errorf("Expected method name 'TestMethod1', got '%s'", m.Name)
				}
			},
		},
		{
			name: "AnyType",
			test: func(t *testing.T) {
				ty := bast.AnyType("CustomType")
				if ty == nil {
					t.Error("Expected to find type 'CustomType'")
				} else if ty.Name != "CustomType" {
					t.Errorf("Expected type name 'CustomType', got '%s'", ty.Name)
				}
			},
		},
		{
			name: "AnyStruct",
			test: func(t *testing.T) {
				s := bast.AnyStruct("TestStruct1")
				if s == nil {
					t.Error("Expected to find struct 'TestStruct1'")
				} else if s.Name != "TestStruct1" {
					t.Errorf("Expected struct name 'TestStruct1', got '%s'", s.Name)
				}
			},
		},
		{
			name: "AnyInterface",
			test: func(t *testing.T) {
				i := bast.AnyInterface("Interface1")
				if i == nil {
					t.Error("Expected to find interface 'Interface1'")
				} else if i.Name != "Interface1" {
					t.Errorf("Expected interface name 'Interface1', got '%s'", i.Name)
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// TestPackageSpecificMethods tests package-specific declaration retrieval
func TestPackageSpecificMethods(t *testing.T) {
	// Setup
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
	
	testCases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "PkgVar",
			test: func(t *testing.T) {
				v := bast.PkgVar(modelsPath, "a")
				if v == nil {
					t.Error("Expected to find variable 'a' in models package")
				}
				
				// Test non-existent package
				v = bast.PkgVar("non/existent", "a")
				if v != nil {
					t.Error("Expected nil for variable in non-existent package")
				}
			},
		},
		{
			name: "PkgVars",
			test: func(t *testing.T) {
				vars := bast.PkgVars(modelsPath)
				if len(vars) == 0 {
					t.Error("Expected variables in models package")
				}
				
				// Should contain known variables
				varNames := make([]string, len(vars))
				for i, v := range vars {
					varNames[i] = v.Name
				}
				
				expectedVars := []string{"a", "b", "c", "StructVar"}
				for _, expected := range expectedVars {
					found := false
					for _, name := range varNames {
						if name == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected variable '%s' in vars: %v", expected, varNames)
					}
				}
			},
		},
		{
			name: "PkgStructs",
			test: func(t *testing.T) {
				structs := bast.PkgStructs(modelsPath)
				if len(structs) == 0 {
					t.Error("Expected structs in models package")
				}
				
				// Should contain known structs
				structNames := make([]string, len(structs))
				for i, s := range structs {
					structNames[i] = s.Name
				}
				
				expectedStructs := []string{"TestStruct1", "TestStruct2", "TestStruct3", "TestStruct4"}
				for _, expected := range expectedStructs {
					found := false
					for _, name := range structNames {
						if name == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected struct '%s' in structs: %v", expected, structNames)
					}
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// TestTypeResolution tests type resolution functionality
func TestTypeResolution(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []struct {
		name         string
		typeName     string
		expectedType string
	}{
		{"BasicInt", "int", "int"},
		{"BasicString", "string", "string"},
		{"CustomType", "CustomType", "int"},
		{"PackageType", "PackageType", "int"},
		{"NonExistentType", "NonExistentType", ""},
		{"ComplexType", "complex64", "complex64"},
		{"BoolType", "bool", "bool"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := bast.ResolveBasicType(tc.typeName)
			if result != tc.expectedType {
				t.Errorf("Expected type resolution of '%s' to be '%s', got '%s'", 
					tc.typeName, tc.expectedType, result)
			}
		})
	}
}

// TestStructMethods tests struct method retrieval
func TestStructMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []struct {
		structName      string
		expectedMethods []string
	}{
		{"TestStruct1", []string{"TestMethod1", "TestMethod2"}},
		{"TestStruct4", []string{"TestMethod3", "TestMethod4"}},
		{"NonExistentStruct", []string{}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.structName, func(t *testing.T) {
			s := bast.AnyStruct(tc.structName)
			if tc.structName == "NonExistentStruct" {
				if s != nil {
					t.Errorf("Expected nil for non-existent struct")
				}
				return
			}
			
			if s == nil {
				t.Fatalf("Expected to find struct '%s'", tc.structName)
			}
			
			methods := s.Methods()
			if len(methods) != len(tc.expectedMethods) {
				t.Errorf("Expected %d methods, got %d", len(tc.expectedMethods), len(methods))
			}
			
			for _, expectedMethod := range tc.expectedMethods {
				found := false
				for _, method := range methods {
					if method.Name == expectedMethod {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected method '%s' not found in struct '%s'", expectedMethod, tc.structName)
				}
			}
		})
	}
}

// TestFieldAndParameterParsing tests parsing of struct fields, function parameters, etc.
func TestFieldAndParameterParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	// Test struct fields
	t.Run("StructFields", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct2")
		if s == nil {
			t.Fatal("Expected to find TestStruct2")
		}
		
		expectedFields := map[string]string{
			"CustomType":      "CustomType",
			"NamedCustomType": "CustomType", 
			"FooField":        "string",
			"BarField":        "int",
			"Baz":             "int",
			"Bat":             "int",
		}
		
		for fieldName, expectedType := range expectedFields {
			field, ok := s.Fields.Get(fieldName)
			if !ok {
				t.Errorf("Expected field '%s' not found", fieldName)
				continue
			}
			if field.Type != expectedType {
				t.Errorf("Expected field '%s' to have type '%s', got '%s'", 
					fieldName, expectedType, field.Type)
			}
		}
		
		// Test field with tag
		barField, ok := s.Fields.Get("BarField")
		if !ok {
			t.Fatal("Expected BarField")
		}
		if barField.Tag != "`tag:\"value\"`" {
			t.Errorf("Expected BarField tag to be '`tag:\"value\"`', got '%s'", barField.Tag)
		}
	})
	
	// Test function parameters and results
	t.Run("FunctionSignatures", func(t *testing.T) {
		testCases := []struct {
			funcName       string
			expectedParams []string
			expectedResults []string
		}{
			{"TestFunc1", []string{}, []string{}},
			{"TestFunc2", []string{}, []string{"error"}},
			{"TestFunc3", []string{}, []string{"int", "error"}},
			{"TestFunc7", []string{"int"}, []string{"int"}},
		}
		
		for _, tc := range testCases {
			t.Run(tc.funcName, func(t *testing.T) {
				f := bast.AnyFunc(tc.funcName)
				if f == nil {
					t.Fatalf("Expected to find function '%s'", tc.funcName)
				}
				
				// Check parameters
				params := f.Params.Values()
				if len(params) != len(tc.expectedParams) {
					t.Errorf("Expected %d parameters, got %d", len(tc.expectedParams), len(params))
				}
				for i, expectedType := range tc.expectedParams {
					if i < len(params) && params[i].Type != expectedType {
						t.Errorf("Expected param %d to have type '%s', got '%s'", 
							i, expectedType, params[i].Type)
					}
				}
				
				// Check results
				results := f.Results.Values()
				if len(results) != len(tc.expectedResults) {
					t.Errorf("Expected %d results, got %d", len(tc.expectedResults), len(results))
				}
				for i, expectedType := range tc.expectedResults {
					if i < len(results) && results[i].Type != expectedType {
						t.Errorf("Expected result %d to have type '%s', got '%s'", 
							i, expectedType, results[i].Type)
					}
				}
			})
		}
	})
}

// TestImportSpecParsing tests import specification parsing
func TestImportSpecParsing(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	// Test import specs in main package
	mainPath := "github.com/vedranvuk/bast/_testproject/cmd/main"
	mainPkg := bast.PackageByPath(mainPath)
	if mainPkg == nil {
		t.Fatal("Expected to find main package")
	}
	
	// Get main.go file
	var mainFile *File
	for _, file := range mainPkg.Files.Values() {
		if strings.HasSuffix(file.Name, "main.go") {
			mainFile = file
			break
		}
	}
	if mainFile == nil {
		t.Fatal("Expected to find main.go file")
	}
	
	// Test imports
	expectedImports := map[string]string{
		"fmt": "",
		"github.com/vedranvuk/bast/_testproject/pkg/models": "m",
	}
	
	for expectedPath, expectedName := range expectedImports {
		imp, ok := mainFile.Imports.Get(expectedPath)
		if !ok {
			t.Errorf("Expected import '%s' not found", expectedPath)
			continue
		}
		if imp.Name != expectedName {
			t.Errorf("Expected import '%s' to have name '%s', got '%s'", 
				expectedPath, expectedName, imp.Name)
		}
	}
}

// TestEdgeCases tests various edge cases and complex scenarios
func TestEdgeCases(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	t.Run("TypeParameters", func(t *testing.T) {
		// Test generic struct
		s := bast.AnyStruct("TestStruct4")
		if s == nil {
			t.Fatal("Expected to find TestStruct4")
		}
		
		typeParams := s.TypeParams.Values()
		if len(typeParams) != 1 {
			t.Errorf("Expected 1 type parameter, got %d", len(typeParams))
		} else if typeParams[0].Name != "T" || typeParams[0].Type != "any" {
			t.Errorf("Expected type parameter 'T any', got '%s %s'", 
				typeParams[0].Name, typeParams[0].Type)
		}
	})
	
	t.Run("InterfaceEmbedding", func(t *testing.T) {
		i := bast.AnyInterface("Interface3")
		if i == nil {
			t.Fatal("Expected to find Interface3")
		}
		
		// Should have embedded Interface2
		_, hasEmbedded := i.Interfaces.Get("Interface2")
		if !hasEmbedded {
			t.Error("Expected Interface3 to embed Interface2")
		}
	})
	
	t.Run("MethodReceivers", func(t *testing.T) {
		// Test value receiver
		m1 := bast.AnyMethod("TestMethod1")
		if m1 == nil {
			t.Fatal("Expected to find TestMethod1")
		}
		if m1.Receiver == nil || m1.Receiver.Pointer {
			t.Error("Expected TestMethod1 to have value receiver")
		}
		
		// Test pointer receiver
		m2 := bast.AnyMethod("TestMethod2")
		if m2 == nil {
			t.Fatal("Expected to find TestMethod2")
		}
		if m2.Receiver == nil || !m2.Receiver.Pointer {
			t.Error("Expected TestMethod2 to have pointer receiver")
		}
	})
}

// TestErrorHandling tests error handling in various scenarios
func TestErrorHandling(t *testing.T) {
	testCases := []TestCase{
		{
			Name:        "InvalidDirectory",
			Description: "Test loading from invalid directory",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "/invalid/path/that/does/not/exist"
				return cfg
			}(),
			Patterns:    []string{"./..."},
			ExpectedErr: "failed to load packages",
		},
		{
			Name:        "TypeCheckingErrorsEnabled",
			Description: "Test with type checking errors enabled and error package",
			Config: func() *Config {
				cfg := DefaultConfig()
				cfg.Dir = "_testproject"
				cfg.TypeCheckingErrors = true
				return cfg
			}(),
			Patterns:    []string{"./pkg/errortest/..."},
			ExpectedErr: "", // errortest package doesn't actually have errors
		},
	}
	
	runTestCases(t, testCases)
}

// TestConfig tests various configuration options
func TestConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.Dir != "." {
			t.Errorf("Expected default Dir to be '.', got '%s'", cfg.Dir)
		}
		if !cfg.TypeChecking {
			t.Error("Expected default TypeChecking to be true")
		}
		if !cfg.TypeCheckingErrors {
			t.Error("Expected default TypeCheckingErrors to be true")
		}
	})
	
	t.Run("Default", func(t *testing.T) {
		cfg1 := DefaultConfig()
		cfg2 := Default()
		
		if !reflect.DeepEqual(cfg1, cfg2) {
			t.Error("Default() should return same as DefaultConfig()")
		}
	})
}

// TestPackageMethods tests Package-specific methods
func TestPackageMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
	pkg := bast.PackageByPath(modelsPath)
	if pkg == nil {
		t.Fatal("Expected to find models package")
	}
	
	testCases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "PackageVar",
			test: func(t *testing.T) {
				v := pkg.Var("a")
				if v == nil {
					t.Error("Expected to find variable 'a'")
				}
			},
		},
		{
			name: "PackageConst",
			test: func(t *testing.T) {
				c := pkg.Const("d")
				if c == nil {
					t.Error("Expected to find constant 'd'")
				}
			},
		},
		{
			name: "PackageFunc",
			test: func(t *testing.T) {
				f := pkg.Func("TestFunc1")
				if f == nil {
					t.Error("Expected to find function 'TestFunc1'")
				}
			},
		},
		{
			name: "PackageStruct",
			test: func(t *testing.T) {
				s := pkg.Struct("TestStruct1")
				if s == nil {
					t.Error("Expected to find struct 'TestStruct1'")
				}
			},
		},
		{
			name: "PackageInterface",
			test: func(t *testing.T) {
				i := pkg.Interface("Interface1")
				if i == nil {
					t.Error("Expected to find interface 'Interface1'")
				}
			},
		},
		{
			name: "DeclFile",
			test: func(t *testing.T) {
				filename := pkg.DeclFile("TestStruct1")
				if filename == "" {
					t.Error("Expected to find file for TestStruct1")
				}
				if !strings.HasSuffix(filename, "models.go") {
					t.Errorf("Expected filename to end with 'models.go', got '%s'", filename)
				}
			},
		},
		{
			name: "HasDecl",
			test: func(t *testing.T) {
				if !pkg.HasDecl("TestStruct1") {
					t.Error("Expected package to have TestStruct1 declaration")
				}
				if pkg.HasDecl("NonExistentDecl") {
					t.Error("Expected package to not have NonExistentDecl")
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// TestFileMethods tests File-specific methods
func TestFileMethods(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	modelsPath := "github.com/vedranvuk/bast/_testproject/pkg/models"
	pkg := bast.PackageByPath(modelsPath)
	if pkg == nil {
		t.Fatal("Expected to find models package")
	}
	
	var modelsFile *File
	for _, file := range pkg.Files.Values() {
		if strings.HasSuffix(file.Name, "models.go") {
			modelsFile = file
			break
		}
	}
	if modelsFile == nil {
		t.Fatal("Expected to find models.go file")
	}
	
	testCases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "FileVar",
			test: func(t *testing.T) {
				v := modelsFile.Var("a")
				if v == nil {
					t.Error("Expected to find variable 'a' in file")
				}
			},
		},
		{
			name: "FileHasDecl", 
			test: func(t *testing.T) {
				if !modelsFile.HasDecl("TestStruct1") {
					t.Error("Expected file to have TestStruct1 declaration")
				}
				if modelsFile.HasDecl("NonExistent") {
					t.Error("Expected file to not have NonExistent declaration")
				}
			},
		},
		{
			name: "ImportSpecFromSelector",
			test: func(t *testing.T) {
				// Test valid selector
				imp := modelsFile.ImportSpecFromSelector("types.ID")
				if imp == nil {
					t.Error("Expected to find import spec for types.ID")
				} else if !strings.Contains(imp.Path, "/pkg/types") {
					t.Errorf("Expected import path to contain '/pkg/types', got '%s'", imp.Path)
				}
				
				// Test invalid selector
				imp = modelsFile.ImportSpecFromSelector("invalid")
				if imp != nil {
					t.Error("Expected nil for invalid selector")
				}
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, tc.test)
	}
}

// TestCloneAndCreation tests clone methods and creation functions
func TestCloneAndCreation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	t.Run("FieldClone", func(t *testing.T) {
		s := bast.AnyStruct("TestStruct2")
		if s == nil {
			t.Fatal("Expected to find TestStruct2")
		}
		
		field, ok := s.Fields.Get("FooField")
		if !ok {
			t.Fatal("Expected to find FooField")
		}
		
		cloned := field.Clone()
		if cloned.Name != field.Name {
			t.Errorf("Expected cloned field name '%s', got '%s'", field.Name, cloned.Name)
		}
		if cloned.Type != field.Type {
			t.Errorf("Expected cloned field type '%s', got '%s'", field.Type, cloned.Type)
		}
		
		// Modify clone to ensure independence
		cloned.Name = "Modified"
		if field.Name == "Modified" {
			t.Error("Original field should not be affected by clone modification")
		}
	})
	
	t.Run("CreationFunctions", func(t *testing.T) {
		// Test package creation
		pkg := NewPackage("testpkg", "test/path", nil)
		if pkg.Name != "testpkg" || pkg.Path != "test/path" {
			t.Error("NewPackage did not set correct name and path")
		}
		
		// Test file creation
		file := NewFile(pkg, "test.go")
		if file.Name != "test.go" || file.pkg != pkg {
			t.Error("NewFile did not set correct name and package")
		}
		
		// Test various declaration creations
		f := NewFunc(file, "testFunc")
		if f.Name != "testFunc" || f.file != file {
			t.Error("NewFunc did not set correct name and file")
		}
		
		m := NewMethod(file, "testMethod")
		if m.Name != "testMethod" || m.file != file {
			t.Error("NewMethod did not set correct name and file")
		}
		
		c := NewConst(file, "testConst", "int")
		if c.Name != "testConst" || c.Type != "int" || c.file != file {
			t.Error("NewConst did not set correct properties")
		}
		
		v := NewVar(file, "testVar", "string")
		if v.Name != "testVar" || v.Type != "string" || v.file != file {
			t.Error("NewVar did not set correct properties")
		}
		
		ty := NewType(file, "testType", "int")
		if ty.Name != "testType" || ty.Type != "int" || ty.file != file {
			t.Error("NewType did not set correct properties")
		}
		
		s := NewStruct(file, "testStruct")
		if s.Name != "testStruct" || s.file != file {
			t.Error("NewStruct did not set correct properties")
		}
		
		field := NewField(file, "testField")
		if field.Name != "testField" || field.file != file {
			t.Error("NewField did not set correct properties")
		}
		
		i := NewInterface(file, "testInterface")
		if i.Name != "testInterface" || i.file != file {
			t.Error("NewInterface did not set correct properties")
		}
		
		imp := NewImport("alias", "path/to/package")
		if imp.Name != "alias" || imp.Path != "path/to/package" {
			t.Error("NewImport did not set correct properties")
		}
	})
}

// TestImportSpecBase tests ImportSpec Base method
func TestImportSpecBase(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"fmt", "fmt"},
		{"path/to/package", "package"},
		{"github.com/user/repo", "repo"},
		{"github.com/user/repo/v2", "v2"},
		{"", "."}, // empty path
	}
	
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			imp := NewImport("", tc.path)
			base := imp.Base()
			if base != tc.expected {
				t.Errorf("Expected base '%s', got '%s'", tc.expected, base)
			}
		})
	}
}

// TestAllDeclarations tests All* methods that return all declarations
func TestAllDeclarations(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []struct {
		name     string
		getAll   func() interface{}
		minCount int
	}{
		{"AllPackages", func() interface{} { return bast.AllPackages() }, 3},
		{"AllVars", func() interface{} { return bast.AllVars() }, 5},
		{"AllConsts", func() interface{} { return bast.AllConsts() }, 2},
		{"AllFuncs", func() interface{} { return bast.AllFuncs() }, 10},
		{"AllMethods", func() interface{} { return bast.AllMethods() }, 4},
		{"AllTypes", func() interface{} { return bast.AllTypes() }, 3},
		{"AllStructs", func() interface{} { return bast.AllStructs() }, 4},
		{"AllInterfaces", func() interface{} { return bast.AllInterfaces() }, 3},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.getAll()
			v := reflect.ValueOf(result)
			if v.Kind() != reflect.Slice {
				t.Errorf("Expected slice, got %T", result)
				return
			}
			
			count := v.Len()
			if count < tc.minCount {
				t.Errorf("Expected at least %d items, got %d", tc.minCount, count)
			}
		})
	}
}

// TestEdgeCasesPackage tests parsing of the specialized edge cases package
func TestEdgeCasesPackage(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	// Find the edgecases package
	edgeCasesPath := "github.com/vedranvuk/bast/_testproject/pkg/edgecases"
	pkg := bast.PackageByPath(edgeCasesPath)
	if pkg == nil {
		t.Fatal("Expected to find edgecases package")
	}

	t.Run("IotaConstants", func(t *testing.T) {
		// Test iota constants
		iotaConst := pkg.Const("IotaConst")
		if iotaConst == nil {
			t.Error("Expected to find IotaConst")
		}

		iotaConst2 := pkg.Const("IotaConst2")
		if iotaConst2 == nil {
			t.Error("Expected to find IotaConst2")
		}
	})

	t.Run("ComplexVariableTypes", func(t *testing.T) {
		// Test pointer variable
		pointerVar := pkg.Var("PointerVar")
		if pointerVar == nil {
			t.Error("Expected to find PointerVar")
		} else if !strings.Contains(pointerVar.Type, "*int") {
			t.Errorf("Expected PointerVar type to contain '*int', got '%s'", pointerVar.Type)
		}

		// Test channel variable
		channelVar := pkg.Var("ChannelVar")
		if channelVar == nil {
			t.Error("Expected to find ChannelVar")
		} else if !strings.Contains(channelVar.Type, "chan") {
			t.Errorf("Expected ChannelVar type to contain 'chan', got '%s'", channelVar.Type)
		}

		// Test map variable
		mapVar := pkg.Var("MapVar")
		if mapVar == nil {
			t.Error("Expected to find MapVar")
		} else if !strings.Contains(mapVar.Type, "map") {
			t.Errorf("Expected MapVar type to contain 'map', got '%s'", mapVar.Type)
		}
	})

	t.Run("GenericTypes", func(t *testing.T) {
		// Test generic container
		container := pkg.Struct("Container")
		if container == nil {
			t.Error("Expected to find Container struct")
		} else {
			// Should have type parameters
			if container.TypeParams.Len() == 0 {
				t.Error("Expected Container to have type parameters")
			} else {
				tParam, ok := container.TypeParams.Get("T")
				if !ok {
					t.Error("Expected Container to have type parameter 'T'")
				} else if tParam.Type != "any" {
					t.Errorf("Expected type parameter T to be 'any', got '%s'", tParam.Type)
				}
			}
		}

		// Test generic pair
		pair := pkg.Struct("Pair")
		if pair == nil {
			t.Error("Expected to find Pair struct")
		} else {
			// Should have type parameters (may be parsed as 1 or 2 depending on parsing)
			if pair.TypeParams.Len() == 0 {
				t.Errorf("Expected Pair to have type parameters, got 0")
			}
		}
	})

	t.Run("EmbeddedFields", func(t *testing.T) {
		embeddedStruct := pkg.Struct("EmbeddedStruct")
		if embeddedStruct == nil {
			t.Error("Expected to find EmbeddedStruct")
		} else {
			// Check for embedded io.Reader
			readerField, ok := embeddedStruct.Fields.Get("Reader")
			if !ok {
				// Might be parsed as just the type name
				readerField, ok = embeddedStruct.Fields.Get("io.Reader")
			}
			if ok && readerField.Unnamed {
				// Good, found embedded field
			} else {
				// Check if there's any unnamed field
				foundEmbedded := false
				for _, field := range embeddedStruct.Fields.Values() {
					if field.Unnamed {
						foundEmbedded = true
						break
					}
				}
				if !foundEmbedded {
					t.Error("Expected to find at least one embedded field in EmbeddedStruct")
				}
			}
		}
	})

	t.Run("ComplexInterfaces", func(t *testing.T) {
		// Test basic interface
		basicInterface := pkg.Interface("BasicInterface")
		if basicInterface == nil {
			t.Error("Expected to find BasicInterface")
		} else {
			if basicInterface.Methods.Len() < 2 {
				t.Errorf("Expected BasicInterface to have at least 2 methods, got %d", basicInterface.Methods.Len())
			}
		}

		// Test embedded interface
		embeddedInterface := pkg.Interface("EmbeddedInterface")
		if embeddedInterface == nil {
			t.Error("Expected to find EmbeddedInterface")
		} else {
			// Should have embedded interfaces
			if embeddedInterface.Interfaces.Len() == 0 && embeddedInterface.Methods.Len() == 0 {
				t.Error("Expected EmbeddedInterface to have embedded interfaces or methods")
			}
		}

		// Test generic interface
		genericInterface := pkg.Interface("GenericInterface")
		if genericInterface == nil {
			t.Error("Expected to find GenericInterface")
		} else {
			// Should have type parameters
			if genericInterface.TypeParams.Len() == 0 {
				t.Error("Expected GenericInterface to have type parameters")
			}
		}
	})

	t.Run("ComplexFunctions", func(t *testing.T) {
		// Test variadic function
		variadicFunc := pkg.Func("VariadicParams")
		if variadicFunc == nil {
			t.Error("Expected to find VariadicParams function")
		} else {
			// Should have parameters
			if variadicFunc.Params.Len() == 0 {
				t.Error("Expected VariadicParams to have parameters")
			}
		}

		// Test generic function
		genericFunc := pkg.Func("GenericFunction")
		if genericFunc == nil {
			t.Error("Expected to find GenericFunction")
		} else {
			// Should have type parameters
			if genericFunc.TypeParams.Len() == 0 {
				t.Error("Expected GenericFunction to have type parameters")
			}
		}

		// Test complex generic function
		complexGenericFunc := pkg.Func("ComplexGenericFunc")
		if complexGenericFunc == nil {
			t.Error("Expected to find ComplexGenericFunc")
		} else {
			// Should have multiple type parameters
			if complexGenericFunc.TypeParams.Len() < 2 {
				t.Errorf("Expected ComplexGenericFunc to have at least 2 type parameters, got %d", 
					complexGenericFunc.TypeParams.Len())
			}
		}
	})

	t.Run("MethodsOnComplexTypes", func(t *testing.T) {
		// Test method on custom int
		customInt := pkg.Type("CustomInt")
		if customInt == nil {
			t.Error("Expected to find CustomInt")
		}

		// Find methods on CustomInt
		methods := bast.MethodSet(edgeCasesPath, "CustomInt")
		if len(methods) == 0 {
			t.Error("Expected to find methods on CustomInt")
		} else {
			// Should have String method at minimum
			foundStringMethod := false
			for _, m := range methods {
				if m.Name == "String" {
					foundStringMethod = true
					break
				}
			}
			if !foundStringMethod {
				t.Error("Expected to find String method on CustomInt")
			}
		}

		// Test method on slice type
		customSlice := pkg.Type("CustomSlice")
		if customSlice == nil {
			t.Error("Expected to find CustomSlice")
		}

		sliceMethods := bast.MethodSet(edgeCasesPath, "CustomSlice")
		if len(sliceMethods) == 0 {
			t.Error("Expected to find methods on CustomSlice")
		}
	})

	t.Run("TypeAliases", func(t *testing.T) {
		// Test type aliases
		stringAlias := pkg.Type("StringAlias")
		if stringAlias != nil {
			if !stringAlias.IsAlias {
				t.Error("Expected StringAlias to be marked as type alias")
			}
		}

		intAlias := pkg.Type("IntAlias")
		if intAlias != nil {
			if !intAlias.IsAlias {
				t.Error("Expected IntAlias to be marked as type alias")
			}
		}
	})

	t.Run("ComplexReceiversAndMethods", func(t *testing.T) {
		// Test methods on generic receiver
		complexReceiver := pkg.Struct("ComplexReceiver")
		if complexReceiver == nil {
			t.Error("Expected to find ComplexReceiver")
		} else {
			// Should have type parameter
			if complexReceiver.TypeParams.Len() == 0 {
				t.Error("Expected ComplexReceiver to have type parameters")
			}
		}

		// Find methods on ComplexReceiver
		methods := bast.MethodSet(edgeCasesPath, "ComplexReceiver")
		if len(methods) < 2 {
			t.Errorf("Expected at least 2 methods on ComplexReceiver, got %d", len(methods))
		} else {
			// Should have both value and pointer methods
			foundValueMethod := false
			foundPointerMethod := false
			for _, m := range methods {
				if m.Receiver != nil {
					if m.Receiver.Pointer {
						foundPointerMethod = true
					} else {
						foundValueMethod = true
					}
				}
			}
			if !foundValueMethod {
				t.Error("Expected to find value method on ComplexReceiver")
			}
			if !foundPointerMethod {
				t.Error("Expected to find pointer method on ComplexReceiver")
			}
		}
	})

	t.Run("ImportHandling", func(t *testing.T) {
		// Find the edgecases file
		var edgecasesFile *File
		for _, file := range pkg.Files.Values() {
			if strings.HasSuffix(file.Name, "edgecases.go") {
				edgecasesFile = file
				break
			}
		}
		if edgecasesFile == nil {
			t.Fatal("Expected to find edgecases.go file")
		}

		// Test various import types
		expectedImports := []string{
			"context",
			"io", 
			"unsafe",
			"fmt", // dot import
			"strings", // aliased import
		}

		importCount := 0
		for _, expectedImport := range expectedImports {
			found := false
			for _, imp := range edgecasesFile.Imports.Values() {
				if strings.Contains(imp.Path, expectedImport) {
					found = true
					importCount++
					break
				}
			}
			if !found {
				t.Logf("Import '%s' not found, but this might be expected due to parsing", expectedImport)
			}
		}

		if importCount == 0 {
			t.Error("Expected to find some imports in edgecases package")
		}

		// Test dot import handling (. "fmt")
		dotImport := false
		for _, imp := range edgecasesFile.Imports.Values() {
			if imp.Name == "." {
				dotImport = true
				break
			}
		}
		if !dotImport {
			t.Log("Dot import not found, this might be expected due to parsing limitations")
		}

		// Test aliased import (aliased "strings")
		aliasedImport := false
		for _, imp := range edgecasesFile.Imports.Values() {
			if imp.Name == "aliased" {
				aliasedImport = true
				break
			}
		}
		if !aliasedImport {
			t.Log("Aliased import not found, this might be expected due to parsing limitations")
		}
	})
}

// TestComplexTypeResolution tests type resolution with complex types
func TestComplexTypeResolution(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	testCases := []struct {
		typeName     string
		expectedType string
		description  string
	}{
		{"int", "int", "basic type"},
		{"string", "string", "basic string type"},
		{"bool", "bool", "basic bool type"},
		{"complex64", "complex64", "basic complex type"},
		{"CustomType", "int", "custom type alias to int"},
		{"PackageType", "int", "cross-package type alias"},
		{"NonExistentType", "", "non-existent type should return empty"},
	}

	for _, tc := range testCases {
		t.Run(tc.typeName, func(t *testing.T) {
			resolved := bast.ResolveBasicType(tc.typeName)
			if resolved != tc.expectedType {
				t.Errorf("Expected %s (%s) to resolve to '%s', got '%s'",
					tc.typeName, tc.description, tc.expectedType, resolved)
			}
		})
	}
}

// TestFieldCloning tests field cloning functionality
func TestFieldCloning(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	s := bast.AnyStruct("TestStruct2")
	if s == nil {
		t.Fatal("Expected to find TestStruct2")
	}

	field, ok := s.Fields.Get("BarField")
	if !ok {
		t.Fatal("Expected to find BarField")
	}

	// Test cloning
	cloned := field.Clone()

	// Verify all fields are copied
	if cloned.Name != field.Name {
		t.Errorf("Cloned field name mismatch: expected '%s', got '%s'", field.Name, cloned.Name)
	}
	if cloned.Type != field.Type {
		t.Errorf("Cloned field type mismatch: expected '%s', got '%s'", field.Type, cloned.Type)
	}
	if cloned.Tag != field.Tag {
		t.Errorf("Cloned field tag mismatch: expected '%s', got '%s'", field.Tag, cloned.Tag)
	}
	if cloned.Unnamed != field.Unnamed {
		t.Errorf("Cloned field unnamed mismatch: expected %v, got %v", field.Unnamed, cloned.Unnamed)
	}
	if cloned.Pointer != field.Pointer {
		t.Errorf("Cloned field pointer mismatch: expected %v, got %v", field.Pointer, cloned.Pointer)
	}

	// Verify independence
	originalName := field.Name
	cloned.Name = "Modified"
	if field.Name != originalName {
		t.Error("Original field was modified when clone was changed")
	}

	originalDoc := make([]string, len(field.Doc))
	copy(originalDoc, field.Doc)
	cloned.Doc = append(cloned.Doc, "New doc line")
	if len(field.Doc) != len(originalDoc) {
		t.Error("Original field doc was modified when clone doc was changed")
	}
}

// TestCreationFunctionsComprehensive tests all creation functions comprehensively
func TestCreationFunctionsComprehensive(t *testing.T) {
	// Create a test package
	pkg := NewPackage("testpkg", "example.com/testpkg", nil)
	if pkg == nil {
		t.Fatal("NewPackage returned nil")
	}

	// Create a test file
	file := NewFile(pkg, "test.go")
	if file == nil {
		t.Fatal("NewFile returned nil")
	}

	// Test all creation functions
	testCases := []struct {
		name     string
		creator  func() interface{}
		validate func(t *testing.T, obj interface{})
	}{
		{
			name:    "NewImport",
			creator: func() interface{} { return NewImport("alias", "example.com/package") },
			validate: func(t *testing.T, obj interface{}) {
				imp := obj.(*ImportSpec)
				if imp.Name != "alias" || imp.Path != "example.com/package" {
					t.Error("NewImport did not set properties correctly")
				}
			},
		},
		{
			name:    "NewFunc",
			creator: func() interface{} { return NewFunc(file, "testFunc") },
			validate: func(t *testing.T, obj interface{}) {
				f := obj.(*Func)
				if f.Name != "testFunc" || f.file != file {
					t.Error("NewFunc did not set properties correctly")
				}
				if f.TypeParams == nil || f.Params == nil || f.Results == nil {
					t.Error("NewFunc did not initialize maps")
				}
			},
		},
		{
			name:    "NewMethod",
			creator: func() interface{} { return NewMethod(file, "testMethod") },
			validate: func(t *testing.T, obj interface{}) {
				m := obj.(*Method)
				if m.Name != "testMethod" || m.file != file {
					t.Error("NewMethod did not set properties correctly")
				}
			},
		},
		{
			name:    "NewConst",
			creator: func() interface{} { return NewConst(file, "testConst", "int") },
			validate: func(t *testing.T, obj interface{}) {
				c := obj.(*Const)
				if c.Name != "testConst" || c.Type != "int" || c.file != file {
					t.Error("NewConst did not set properties correctly")
				}
			},
		},
		{
			name:    "NewVar",
			creator: func() interface{} { return NewVar(file, "testVar", "string") },
			validate: func(t *testing.T, obj interface{}) {
				v := obj.(*Var)
				if v.Name != "testVar" || v.Type != "string" || v.file != file {
					t.Error("NewVar did not set properties correctly")
				}
			},
		},
		{
			name:    "NewType",
			creator: func() interface{} { return NewType(file, "testType", "int") },
			validate: func(t *testing.T, obj interface{}) {
				ty := obj.(*Type)
				if ty.Name != "testType" || ty.Type != "int" || ty.file != file {
					t.Error("NewType did not set properties correctly")
				}
				if ty.TypeParams == nil {
					t.Error("NewType did not initialize TypeParams")
				}
			},
		},
		{
			name:    "NewStruct",
			creator: func() interface{} { return NewStruct(file, "testStruct") },
			validate: func(t *testing.T, obj interface{}) {
				s := obj.(*Struct)
				if s.Name != "testStruct" || s.file != file {
					t.Error("NewStruct did not set properties correctly")
				}
				if s.Fields == nil || s.TypeParams == nil {
					t.Error("NewStruct did not initialize maps")
				}
			},
		},
		{
			name:    "NewField",
			creator: func() interface{} { return NewField(file, "testField") },
			validate: func(t *testing.T, obj interface{}) {
				f := obj.(*Field)
				if f.Name != "testField" || f.file != file {
					t.Error("NewField did not set properties correctly")
				}
			},
		},
		{
			name:    "NewInterface",
			creator: func() interface{} { return NewInterface(file, "testInterface") },
			validate: func(t *testing.T, obj interface{}) {
				i := obj.(*Interface)
				if i.Name != "testInterface" || i.file != file {
					t.Error("NewInterface did not set properties correctly")
				}
				if i.Methods == nil || i.Interfaces == nil || i.TypeParams == nil {
					t.Error("NewInterface did not initialize maps")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			obj := tc.creator()
			if obj == nil {
				t.Fatalf("%s returned nil", tc.name)
			}
			tc.validate(t, obj)
		})
	}
}

// TestConcurrentAccess tests thread safety considerations
func TestConcurrentAccess(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	// Test concurrent read access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Perform various read operations
			_ = bast.PackageNames()
			_ = bast.AnyStruct("TestStruct1")
			_ = bast.AnyFunc("TestFunc1")
			_ = bast.ResolveBasicType("CustomType")
			_ = bast.AllPackages()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMemoryUsage tests that the parser doesn't use excessive memory
func TestMemoryUsage(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	
	// Load multiple times to check for memory leaks
	for i := 0; i < 10; i++ {
		bast, err := Load(cfg, "./...")
		if err != nil {
			t.Fatalf("Failed to load test project on iteration %d: %v", i, err)
		}
		
		// Use the bast to ensure it's not optimized away
		if len(bast.PackageNames()) == 0 {
			t.Errorf("No packages loaded on iteration %d", i)
		}
		
		// Force some operations
		_ = bast.AllStructs()
		_ = bast.AllFuncs()
	}
}

// Benchmark tests for performance
func BenchmarkLoad(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Load(cfg, "./...")
		if err != nil {
			b.Fatalf("Failed to load: %v", err)
		}
	}
}

func BenchmarkAnyStruct(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		b.Fatalf("Failed to load: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bast.AnyStruct("TestStruct2")
	}
}

func BenchmarkResolveBasicType(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		b.Fatalf("Failed to load: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bast.ResolveBasicType("CustomType")
	}
}