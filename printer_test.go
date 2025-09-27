package bast

import (
	"bytes"
	"strings"
	"testing"
)

// TestPrinter tests the printer functionality
func TestPrinter(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	t.Run("DefaultPrinter", func(t *testing.T) {
		printer := DefaultPrinter()
		
		// Test all flags are enabled by default
		if !printer.PrintDoc {
			t.Error("Expected PrintDoc to be true")
		}
		if !printer.PrintComments {
			t.Error("Expected PrintComments to be true")
		}
		if !printer.PrintConsts {
			t.Error("Expected PrintConsts to be true")
		}
		if !printer.PrintVars {
			t.Error("Expected PrintVars to be true")
		}
		if !printer.PrintTypes {
			t.Error("Expected PrintTypes to be true")
		}
		if !printer.PrintFuncs {
			t.Error("Expected PrintFuncs to be true")
		}
		if !printer.PrintMethods {
			t.Error("Expected PrintMethods to be true")
		}
		if !printer.PrintStructs {
			t.Error("Expected PrintStructs to be true")
		}
		if !printer.PrintInterfaces {
			t.Error("Expected PrintInterfaces to be true")
		}
		if printer.Indentation != "\t" {
			t.Errorf("Expected default indentation to be tab, got '%s'", printer.Indentation)
		}
	})

	t.Run("PrintFunction", func(t *testing.T) {
		var buf bytes.Buffer
		Print(&buf, bast)
		
		output := buf.String()
		if len(output) == 0 {
			t.Error("Expected non-empty output from Print")
		}
		
		// Should contain package information
		if !strings.Contains(output, "Package") {
			t.Error("Expected output to contain 'Package'")
		}
		
		// Should contain some of our test declarations
		expectedStrings := []string{
			"TestStruct1",
			"TestFunc1", 
			"CustomType",
			"Interface1",
		}
		
		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain '%s'", expected)
			}
		}
	})

	t.Run("CustomPrinter", func(t *testing.T) {
		// Create printer with selective output
		printer := &Printer{
			PrintDoc:        false,
			PrintComments:   false,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      true,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    true,
			PrintInterfaces: false,
			Indentation:     "  ", // spaces instead of tabs
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		if len(output) == 0 {
			t.Error("Expected non-empty output from custom printer")
		}
		
		// Should contain structs and types
		if !strings.Contains(output, "TestStruct1") {
			t.Error("Expected output to contain structs")
		}
		if !strings.Contains(output, "CustomType") {
			t.Error("Expected output to contain types")
		}
		
		// Should not contain functions (disabled)
		if strings.Contains(output, "TestFunc1") {
			t.Error("Expected output to not contain functions (disabled)")
		}
		
		// Should not contain interfaces (disabled)
		if strings.Contains(output, "Interface1") {
			t.Error("Expected output to not contain interfaces (disabled)")
		}
	})

	t.Run("OnlyDocumentation", func(t *testing.T) {
		// Printer that only shows documentation
		printer := &Printer{
			PrintDoc:        true,
			PrintComments:   true,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    false,
			PrintInterfaces: false,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should contain documentation comments
		if !strings.Contains(output, "Package test description goes here") {
			t.Error("Expected output to contain package documentation")
		}
	})

	t.Run("NoDocumentation", func(t *testing.T) {
		// Printer that excludes all documentation
		printer := &Printer{
			PrintDoc:        false,
			PrintComments:   false,
			PrintConsts:     true,
			PrintVars:       true,
			PrintTypes:      true,
			PrintFuncs:      true,
			PrintMethods:    true,
			PrintStructs:    true,
			PrintInterfaces: true,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should not contain doc comments
		if strings.Contains(output, "Package test description goes here") {
			t.Error("Expected output to not contain package documentation")
		}
		
		// Should still contain declarations
		if !strings.Contains(output, "TestStruct1") {
			t.Error("Expected output to contain declarations without docs")
		}
	})

	t.Run("CustomIndentation", func(t *testing.T) {
		printer := &Printer{
			PrintDoc:        false,
			PrintComments:   false,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    true,
			PrintInterfaces: false,
			Indentation:     "    ", // 4 spaces
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Check that custom indentation is used
		lines := strings.Split(output, "\n")
		foundIndentedLine := false
		for _, line := range lines {
			if strings.HasPrefix(line, "    ") && len(strings.TrimSpace(line)) > 0 {
				foundIndentedLine = true
				break
			}
		}
		if !foundIndentedLine {
			t.Error("Expected to find lines with 4-space indentation")
		}
	})
}

// TestPrinterEdgeCases tests printer behavior with edge cases
func TestPrinterEdgeCases(t *testing.T) {
	t.Run("EmptyBast", func(t *testing.T) {
		// Create empty bast
		emptyBast := new()
		
		var buf bytes.Buffer
		Print(&buf, emptyBast)
		
		output := buf.String()
		// Should not panic and should produce some output
		_ = len(output) // Just checking it doesn't panic
	})

	t.Run("AllFlagsDisabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		bast, err := Load(cfg, "./...")
		if err != nil {
			t.Fatalf("Failed to load test project: %v", err)
		}

		// Printer with all flags disabled
		printer := &Printer{
			PrintDoc:        false,
			PrintComments:   false,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    false,
			PrintInterfaces: false,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should still show package structure but minimal content
		if !strings.Contains(output, "Package") {
			t.Error("Expected to still show package headers even with all flags disabled")
		}
	})

	t.Run("LargeOutput", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Dir = "_testproject"
		bast, err := Load(cfg, "./...")
		if err != nil {
			t.Fatalf("Failed to load test project: %v", err)
		}

		var buf bytes.Buffer
		Print(&buf, bast)
		
		output := buf.String()
		
		// Verify comprehensive output
		expectedElements := []string{
			"Package", "File", "Var", "Const", "Func", "Method", "Struct", "Interface",
			"Type Param", "Param", "Result", "Field",
		}
		
		for _, element := range expectedElements {
			if !strings.Contains(output, element) {
				t.Errorf("Expected comprehensive output to contain '%s'", element)
			}
		}
	})
}

// TestPrinterComponents tests individual printer components
func TestPrinterComponents(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		t.Fatalf("Failed to load test project: %v", err)
	}

	t.Run("StructPrinting", func(t *testing.T) {
		printer := &Printer{
			PrintDoc:        true,
			PrintComments:   true,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    true,
			PrintInterfaces: false,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should contain struct information
		structElements := []string{
			"TestStruct1",
			"TestStruct2", 
			"TestStruct3",
			"TestStruct4",
			"Field",
		}
		
		for _, element := range structElements {
			if !strings.Contains(output, element) {
				t.Errorf("Expected struct output to contain '%s'", element)
			}
		}
	})

	t.Run("FunctionPrinting", func(t *testing.T) {
		printer := &Printer{
			PrintDoc:        true,
			PrintComments:   true,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      true,
			PrintMethods:    false,
			PrintStructs:    false,
			PrintInterfaces: false,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should contain function information
		functionElements := []string{
			"TestFunc1",
			"TestFunc2",
			"TestFunc7",
		}
		
		for _, element := range functionElements {
			if !strings.Contains(output, element) {
				t.Errorf("Expected function output to contain '%s'", element)
			}
		}
		
		// Should contain parameter/result information for complex functions
		if !strings.Contains(output, "Param") && !strings.Contains(output, "Result") {
			t.Error("Expected function output to contain parameter or result information")
		}
	})

	t.Run("InterfacePrinting", func(t *testing.T) {
		printer := &Printer{
			PrintDoc:        true,
			PrintComments:   true,
			PrintConsts:     false,
			PrintVars:       false,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    false,
			PrintInterfaces: true,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should contain interface information
		interfaceElements := []string{
			"Interface1",
			"Interface2",
			"Interface3",
		}
		
		for _, element := range interfaceElements {
			if !strings.Contains(output, element) {
				t.Errorf("Expected interface output to contain '%s'", element)
			}
		}
	})

	t.Run("VarConstPrinting", func(t *testing.T) {
		printer := &Printer{
			PrintDoc:        true,
			PrintComments:   true,
			PrintConsts:     true,
			PrintVars:       true,
			PrintTypes:      false,
			PrintFuncs:      false,
			PrintMethods:    false,
			PrintStructs:    false,
			PrintInterfaces: false,
			Indentation:     "\t",
		}
		
		var buf bytes.Buffer
		printer.Print(&buf, bast)
		
		output := buf.String()
		
		// Should contain var and const information
		if !strings.Contains(output, "Var") {
			t.Error("Expected output to contain 'Var'")
		}
		if !strings.Contains(output, "Const") {
			t.Error("Expected output to contain 'Const'")
		}
	})
}

// Benchmark printer performance
func BenchmarkPrint(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		b.Fatalf("Failed to load test project: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		Print(&buf, bast)
	}
}

func BenchmarkCustomPrint(b *testing.B) {
	cfg := DefaultConfig()
	cfg.Dir = "_testproject"
	bast, err := Load(cfg, "./...")
	if err != nil {
		b.Fatalf("Failed to load test project: %v", err)
	}

	printer := &Printer{
		PrintDoc:        false,
		PrintComments:   false,
		PrintConsts:     false,
		PrintVars:       false,
		PrintTypes:      true,
		PrintFuncs:      true,
		PrintMethods:    false,
		PrintStructs:    true,
		PrintInterfaces: false,
		Indentation:     "\t",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		printer.Print(&buf, bast)
	}
}