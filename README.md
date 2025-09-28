# BAST

[![Go Reference](https://pkg.go.dev/badge/github.com/vedranvuk/bast.svg)](https://pkg.go.dev/github.com/vedranvuk/bast)

BAST is a Go package that provides a simplified intermediate representation of Go source code declarations. Built on top of Go's standard `go/ast` and `go/types` packages, BAST transforms complex AST structures into an intuitive model optimized for code analysis and generation tools.

## Features

- **Simplified API**: Clean, intuitive interface for accessing Go declarations without navigating complex AST structures
- **Template-Friendly**: Designed specifically for use with Go's `text/template` package for code generation
- **Type Resolution**: Automatic resolution of basic types and package imports with optional type checking
- **Modern Go Support**: Full support for generics, type parameters, and modern Go language features  
- **Cross-Package Analysis**: Query declarations across multiple packages with unified methods
- **Flexible Loading**: Support for standard Go package patterns (`./...`, import paths, etc.)
- **Rich Metadata**: Preserve documentation, comments, and structural relationships

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/vedranvuk/bast"
)

func main() {
    // Load Go packages from current directory
    b, err := bast.Load(bast.Default(), "./...")
    if err != nil {
        log.Fatal(err)
    }
    
    // Find all structs across packages
    for _, pkg := range b.Packages() {
        for _, s := range pkg.Structs.Values() {
            fmt.Printf("Struct: %s.%s\n", pkg.Name, s.Name)
            
            // Print struct fields
            for _, field := range s.Fields.Values() {
                fmt.Printf("  %s %s\n", field.Name, field.Type)
            }
            
            // Print methods
            for _, method := range s.Methods() {
                fmt.Printf("  func %s(...)\n", method.Name)
            }
        }
    }
}
```

## Use Cases

**Code Generation**: Generate boilerplate code, interfaces, mocks, or serialization logic from existing types.

```go
// Find all structs implementing a specific interface pattern
structs := b.AllStructs()
for _, s := range structs {
    if hasValidateMethod(s) {
        generateValidator(s)
    }
}
```

**API Documentation**: Extract and format API documentation from source code.

```go
// Document all public functions in a package
pkg := b.PackageByPath("github.com/myorg/mypackage")
for _, fn := range pkg.Funcs.Values() {
    if isPublic(fn.Name) {
        generateDoc(fn.Name, fn.Doc, fn.Params, fn.Results)
    }
}
```

**Static Analysis**: Analyze code patterns, dependencies, and structural relationships.

```go
// Find all types that embed a specific struct
embedders := b.TypesOfType("mypackage", "BaseStruct")
for _, typ := range embedders {
    analyzeEmbedding(typ)
}
```

**Refactoring Tools**: Build tools that understand and transform Go code structures.

## Configuration

```go
cfg := &bast.Config{
    Dir:                ".",              // Base directory
    Tests:              true,             // Include test files
    TypeChecking:       true,             // Enable type resolution
    TypeCheckingErrors: false,            // Ignore type errors
    BuildFlags:         []string{"-tags", "integration"},
}

b, err := bast.Load(cfg, "./cmd/...", "./pkg/...")
```

## Package Structure

BAST organizes code into a hierarchical model:

- **Bast**: Root container with cross-package query methods
- **Package**: Represents a Go package with its declarations
- **File**: Individual source files with imports and file-level declarations
- **Declarations**: Vars, Consts, Funcs, Methods, Types, Structs, Interfaces
- **Fields**: Function parameters, struct fields, interface methods

Each level provides both local queries (`pkg.Struct("MyType")`) and global searches (`b.AnyStruct("MyType")`).

## Type Resolution

BAST can resolve type aliases and named types to their underlying basic types:

```go
// With type checking enabled
basicType := b.ResolveBasicType("MyIntAlias") // Returns "int"
basicType = b.ResolveBasicType("pkg.CustomString") // Returns "string"
```

## Template Integration

BAST structures are designed to work seamlessly with Go templates:

```go
const tmpl = `
{{range .AllStructs}}
type {{.Name}} struct {
    {{range .Fields.Values}}{{.Name}} {{.Type}}{{end}}
}
{{range .Methods}}
func ({{.Receiver.Name}} {{.Receiver.Type}}) {{.Name}}({{range .Params.Values}}{{.Name}} {{.Type}}{{end}})
{{end}}
{{end}}
`

t := template.Must(template.New("").Parse(tmpl))
t.Execute(os.Stdout, bast)
```

## Status

Production ready. The API is stable and follows semantic versioning.

## License

MIT. See [LICENSE](LICENSE) file for details.
