# BAST

Bastard AST is lightweight model of top level Go declarations constructed from [`go/ast`](https://pkg.go.dev/go/ast) and intended for code generation using [`text/template`](https://pkg.go.dev/text/template).

# Status

Experimental. API is subject to change.

## Example

Loading is straightforward like with `golang.org/x/tools/go/packages`. Specify
file, directory or module path one or more times to load into bast then pass it
to a template.

```Go
package main

import (
	"os"
	"text/template"
	"github.com/vedranvuk/bast"
)

var (
	err error
	data *bast.Bast
	file []byte
	tmpl *template.Template = template.New("bast")
)

func main() {
	if data, err = bast.Load("file.go", "./pkg/package", "net/http"); err != nil {
		fatal(err)
	}
	if file, err = os.ReadFile("template.tmpl"); err != nil {
		panic(err)
	}
	if tmpl, err = tmpl.Funcs(data.FuncMap()).Parse(string(file)); err != nil {
		panic(err)
	}
	if err = tmpl.ExecuteTemplate("bast", os.Stdout, data); err != nil {
		panic(err)
	}
}
```

## Reference command

Bast comes with a command that serves as use example. It can execute a single template file at a time. It also has a lot more docs but best docs are the source docs. 

To install it: `go install github.com/vedranvuk/bast`.

Tool reference is at `bast -f`.

For slightly more boilerplating power check out [boil](https://github.com/vedranvuk/boil).

## License

MIT.
