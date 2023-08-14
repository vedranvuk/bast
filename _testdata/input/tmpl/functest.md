# Function reference

This document lists functions available to a template file being executed by the `bast` command. For `text/template` reference and usage click [here](https://pkg.go.dev/text/template).

- [Standard library functions](#standard-library-functions)
  - [Utilities](#)
    - [call](#call)
    - [len](#len)
    - [index](#index)
    - [slice](#slice)
    - [print](#print)
    - [printf](#printf)
    - [println](#println)
  - [Escaping](#)
    - [html](#html)
    - [js](#len)
    - [urlquery](#urlquery)
  - [Logic](#)
    - [and](#and)
    - [or](#or)
    - [not](#not)
  - [Comparisons](#comparisons)
    - [eq](#eq)
    - [ge](#ge)
    - [gt](#gt)
    - [le](#le)
    - [lt](#lt)
    - [ne](#ne)
- [Bast functions](#bast)
  - [Declaration retrieval by declaration name](#)
    - [var](#var)
    - [const](#const)
    - [type](#type)
    - [func](#func)
    - [method](#method)
    - [interface](#interface)
    - [struct](#struct)
  - [Declaration retrieval by parent package](#)
    - [vars](#vars)
    - [consts](#consts)
    - [types](#types)
    - [funcs](#funcs)
    - [methods](#methods)
    - [interfaces](#interfaces)
    - [structs](#structs)
  - [Global declaration retrieval](#)
    - [allvars](#allvars)
    - [allconsts](#allconsts)
    - [alltypes](#alltypes)
    - [allfuncs](#allfuncs)
    - [allmethods](#allmethods)
    - [allinterfaces](#allinterfaces)
    - [allstructs](#allstructs)
  - [Utilities](#)
    - [varsoftype](#varsoftype)
    - [constsoftype](#constsoftype)
    - [structmethods](#structmethods)
    - [structfieldnames](#structfieldnames)

---

## Standard library functions

These functions are provided by `text-template` package and are always available.
They are listed here for quick reference.

### call

Returns the result of evaluating the first argument as a function. The function
must return 1 result, or 2 results, the second of which is an error.

```
{{call .SomeFunction "a" "b" "c"}}
```

### len

Returns the length of the item, with an error if it has no defined length.

```
Length of Vars: {{len allvars}}
```

### index

Returns the result of indexing its first argument by the following arguments.
Thus "index x 1 2 3" is, in Go syntax, x[1][2][3]. Each indexed item must be a
map, slice, or array.

```
Element at [1][2][3]: {{index .MyMultiDimArray 1 2 3}}
```

### slice

Returns the result of slicing its first argument by the remaining arguments.
Thus "slice x 1 2" is, in Go syntax, x[1:2], while "slice x" is x[:], "slice x 1"
is x[1:], and "slice x 1 2 3" is x[1:2:3]. The first argument must be a string,
slice, or array.

```
{{$args := join "a" "b" "c" "d"}}
{{slice $args 2 1}}
// Output: c
```

### print

Maps to fmt.Sprint and returns its result.

```
{{print "print"}}
```

### printf

Maps to fmt.Sprintf and returns its result.

```
{{printf "%s" "printf"}}
```

### println

Maps to fmt.Sprintln and returns its result.

```
{{println "println"}}
```

### html

Returns the escaped HTML equivalent of the textual representation of its arguments.

```
{{- $html := "<p>Hello World!</p>"}}
<p>Hello World!</p> = {{html $html}}
```

### js

Returns the escaped JavaScript equivalent of the textual representation of its
arguments.

```
{{- $js := "function(name) { return name + 3.14 }"}}
function(name) { return name + 3.14 } = {{js $js}}
```

### urlquery

Returns the escaped value of the textual representation of its arguments in a
form suitable for embedding in a URL query.

```
{{- $url := "http://example.com/user?id=42&theme=pink"}}
http://example.com/user?id=42&theme=pink = {{ urlquery $url}}
```

### and

Computes the Boolean AND of its arguments, returning the first false argument it
encounters, or the last argument.

```
1 and 1 = {{and 1 1}}
```

### or

Computes the Boolean OR of its arguments, returning the first true argument it
encounters, or the last argument.

```
1 or 0 = {{or 1 0}}
```

### not

Returns the Boolean negation of its argument.

```
0 = {{ not 1}}
```

## Comparisons

### eq

Returns true if both of its arguments have equal values.

```
2 == 2 = {{eq 2 2}}
```

### ge

Returns true if first argument is greater than or equal to its second argument.

```
5 >= 4 = {{ge 5 4}}
```

### gt

Returns true if first argument is greater than its second argument.

```
5 > 4 = {{gt 5 4 }}
```

### le

Returns true if first argument is less than or equal to its second argument.

```
4 <= 5 = {{le 4 5}}
```

### lt

Returns true if first argument is less than its second argument.

```
4 <= 5 = {{lt 4 5}}
```

### ne

Returns true if first argument is not equal to its second argument.

```
4 != 5 = {{ne 4 5}}
```

## BAST

### var

Retrieves a variable from a package by name.

```
{{ $var := var "main" "Variable"}}{{$var.Name}} = {{$var.Value}}
```

### const

Retrieves a constant from a package by name.

```
{{ $const := const "main" "Constant"}}{{$const.Name}} = {{$const.Value}}
```

### type

Retrieves a type from a package by name.

```
{{ $type := type "main" "MyInt"}}{{$type.Name}}
```

### func

Retrieves a function from a package by name.

```
{{ $func := func "main" "Echo" -}}
{{$func.Name}}
{{$func.Doc}}
```

### method

Retrieves a method from a package by name.

```
{{ $method := method "main" "Method" -}}
{{$receiver := index $method.Receivers 0 -}}
{{$method.Name}} ({{$receiver.Name}} {{$receiver.Type}})
{{ $method := method "main" "PointerMethod" -}}
{{$receiver := index $method.Receivers 0 -}}
{{$method.Name}} ({{$receiver.Name}} {{$receiver.Type}})
```

### interface

Retrieves an interface from a package by name.

```
{{$intf := interface "main" "Interface" -}}
{{$intf.Name}}
{{range $method := $intf.Methods}}
{{$method.Name}}{{end}}
```

### struct

Retrieves a struct from a package by name.

```
{{$struct := struct "main" "Struct" -}}
{{$struct.Name}}
{{range $field := $struct.Fields}}
{{$field.Name}}: {{$field.Type}}{{end}}
```

### trimpfx

Trims a prefix from a string.

```
{{trimpfx "Hello World!" "Hello "}}
```

### trimsfx

Trims a suffix from a string.

```
{{trimsfx "Hello World!" " World!"}}
```

### split

Splits a string by separator.

```
{{- $array := split "One\tTwo\tThree" "\t" -}}
{{range $string := $array}}
{{$string}}{{end}}
```

### join

Joins strings with separator.

```
{{join ", " "One" "Two" "Three"}}
```

```
{{$names := fieldnames "main" "Struct"}}
db.Exec("INSERT INTO table VALUES ({{len $names | repeat "?" ", "}})",{{range $names}}
	input.{{.}},{{end}}
)
```

### repeat

Repeats a string with optional delimiter n times.

```
INSERT INTO table VALUES ({{repeat "?" ", " 15}})
```

### vars

Returns all vars from a package.

```
{{range vars "main"}}
{{.Name}}{{end}}
```

### consts

Returns all consts from a package.

```
{{range consts "main"}}
{{.Name}}{{end}}
```

### types

Returns all types from a package.

```
{{range types "main"}}
{{.Name}}{{end}}
```

### funcs

Retrieves all functions from a package.

```
{{range funcs "main"}}
{{.Name}}{{end}}

```

### methods

Retrieves all methods from a package.

```
{{range methods "main"}}
{{.Name}}{{end}}

```

### interfaces

Retrieves all interfaces from a package.

```
{{range interfaces "main"}}
{{.Name}}{{end}}

```

### structs

Retrieves all structs from a package.

```
{{range structs "main"}}
{{.Name}}{{end}}
```

### allvars

Returns all vars from a package.

```
{{range allvars}}
{{.Name}}{{end}}
```

### allconsts

Returns all consts from a package.

```
{{range allconsts}}
{{.Name}}{{end}}
```

### alltypes

Returns all types from a package.

```
{{range alltypes}}
{{.Name}}{{end}}
```

### allfuncs

Retrieves all functions from a package.

```
{{range allfuncs}}
{{.Name}}{{end}}
```

### allmethods

Retrieves all methods from a package.

```
{{range allmethods}}
{{.Name}}{{end}}
```

### allinterfaces

Retrieves all interfaces from a package.

```
{{range allinterfaces}}
{{.Name}}{{end}}
```

### allstructs

Retrieves all structs from a package.

```
{{range allstructs}}
{{.Name}}{{end}}
```

### varsoftype

Returns all variables from a package that have specific type.

```
{{range varsoftype "int"}}
{{.Name}}{{end}}
```

### constsoftype

Returns all constants from a package that have specific type.

```
{{range constsoftype "int"}}
{{.Name}}{{end}}
```

### methodset

Returns methods for a type from a package by its name.

```
{{range methodset "main" "Struct"}}
{{.Name}}{{end}}
```

### fieldnames

Returns a slice of field names of a struct in a package.
```
{{fieldnames "main" "Struct"}}
```
