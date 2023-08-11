# Function reference

## std functions

### and

Computes the Boolean AND of its arguments, returning the first false argument it
encounters, or the last argument.

```
1 and 1 = {{and 1 1}}
```

### call

Returns the result of evaluating the first argument as a function. The function
must return 1 result, or 2 results, the second of which is an error.

```
// TODO call example
```

### html

Returns the escaped HTML equivalent of the textual representation of its arguments.

```
{{- $html := "<p>Hello World!</p>"}}
<p>Hello World!</p> = {{html $html}}
```

### index

Returns the result of indexing its first argument by the following arguments.
Thus "index x 1 2 3" is, in Go syntax, x[1][2][3]. Each indexed item must be a
map, slice, or array.

```
// TODO index example
```

### slice

Returns the result of slicing its first argument by the remaining arguments.
Thus "slice x 1 2" is, in Go syntax, x[1:2], while "slice x" is x[:], "slice x 1"
is x[1:], and "slice x 1 2 3" is x[1:2:3]. The first argument must be a string,
slice, or array.

```
// TODO slice example
```

### js

Returns the escaped JavaScript equivalent of the textual representation of its
arguments.

```
{{- $js := "function(name) { return name + 3.14 }"}}
function(name) { return name + 3.14 } = {{js $js}}
```

### len

Returns the length of the item, with an error if it has no defined length.

```
Length of Vars: {{len allvars }}
```

### not

Returns the Boolean negation of its argument.

```
0 = {{ not 1}}
```

### or

Computes the Boolean OR of its arguments, returning the first true argument it
encounters, or the last argument.

```
1 or 0 = {{or 1 0}}
```

### print

Maps to fmt.Sprint.

```
{{print "print"}}
```

### printf

Maps to fmt.Sprintf.

```
{{printf "%s" "printf"}}
```

### println

Maps to fmt.Sprintln.

```
{{println "println"}}
```

### urlquery

Returns the escaped value of the textual representation of its arguments in a
form suitable for embedding in a URL query.

```
{{- $url := "http://example.com/user?id=42&theme=pink"}}
http://example.com/user?id=42&theme=pink = {{ urlquery $url}}
```

## Comparisons

### eq

Evaluates the comparison a == b || a == c || ...

```
2 == 2 = {{eq 2 2}}
```

### ge

Evaluates the comparison a >= b.

```
5 >= 4 = {{ge 5 4}}
```

### gt

Evaluates the comparison a > b.

```
5 > 4 = {{gt 5 4 }}
```

### le

Evaluates the comparison <= b.

```
4 <= 5 = {{le 4 5}}
```

### lt

Evaluates the comparison a < b.

```
4 <= 5 = {{lt 4 5}}
```

### ne

Evaluates the comparison a != b.

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
{{$names := structfieldnames "main" "Struct"}}
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
