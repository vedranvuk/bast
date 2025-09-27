# BAST

BAST (Bastard AST) is a lightweight model of top-level Go declarations, constructed from the standard `go/ast` package. It is designed to simplify source code analysis and code generation.

While the standard `go/ast` and `go/types` packages are powerful, they can be complex to navigate. BAST provides a simpler, more direct API to access basic type and declaration information from Go packages.

## Getting Started

Here is a quick example of how to use BAST to inspect the methods of a struct in a project.

Assuming you are running this from the `bast` project root, which contains the `_testproject` directory:

```go
package main

import (
	"fmt"
	"log"

	"github.com/vedranvuk/bast"
)

func main() {
	// Configure loading for a directory.
	cfg := bast.DefaultConfig()
	cfg.Dir = "./_testproject" // Path to the project to analyze.

	// Load and parse the packages.
	b, err := bast.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("failed to load: %v", err)
	}

	// Find a specific struct by name from any package.
	myStruct := b.AnyStruct("TestStruct4")
	if myStruct == nil {
		log.Fatal("struct not found")
	}

	// Print the methods of the struct.
	fmt.Printf("Methods of %s:\n", myStruct.Name)
	for _, method := range myStruct.Methods() {
		fmt.Printf("- %s\n", method.Name)
	}
}
```

Running this code will produce the following output:

```
Methods of TestStruct4:
- TestMethod1
- TestMethod2
```

# Status

Experimental. The API is subject to change.

## License

MIT.
