# Function reference

## std functions

### and

Computes the Boolean AND of its arguments, returning the first false argument it
encounters, or the last argument.

```
1 and 1 = 1
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
<p>Hello World!</p> = &lt;p&gt;Hello World!&lt;/p&gt;
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
function(name) { return name + 3.14 } = function(name) { return name + 3.14 }
```

### len

Returns the length of the item, with an error if it has no defined length.

```
Length of Vars: 1
```

### not

Returns the Boolean negation of its argument.

```
0 = false
```

### or

Computes the Boolean OR of its arguments, returning the first true argument it
encounters, or the last argument.

```
1 or 0 = 1
```

### print

Maps to fmt.Sprint.

```
print
```

### printf

Maps to fmt.Sprintf.

```
printf
```

### println

Maps to fmt.Sprintln.

```
println

```

### urlquery

Returns the escaped value of the textual representation of its arguments in a
form suitable for embedding in a URL query.

```
http://example.com/user?id=42&theme=pink = http%3A%2F%2Fexample.com%2Fuser%3Fid%3D42%26theme%3Dpink
```

## Comparisons

### eq

Evaluates the comparison a == b || a == c || ...

```
2 == 2 = true
```

### ge

Evaluates the comparison a >= b.

```
5 >= 4 = true
```

### gt

Evaluates the comparison a > b.

```
5 > 4 = true
```

### le

Evaluates the comparison <= b.

```
4 <= 5 = true
```

### lt

Evaluates the comparison a < b.

```
4 <= 5 = true
```

### ne

Evaluates the comparison a != b.

```
4 != 5 = true
```

## BAST

### var

Retrieves a variable from a package by name.

```
Variable = "Value"
```

### const

Retrieves a constant from a package by name.

```
Constant = 42
```

### type

Retrieves a type from a package by name.

```
MyInt
```

### func

Retrieves a function from a package by name.

```
Echo
[// Echo echoes input to output.]
```

### method

Retrieves a method from a package by name.

```
Method (self Struct)
PointerMethod (self *Struct)
```

### interface

Retrieves an interface from a package by name.

```
Interface

MethodA
MethodB
MethodC
```

### struct

Retrieves a struct from a package by name.

```
Struct

Name: string
Age: int
```

### trimpfx

Trims a prefix from a string.

```
World!
```

### trimsfx

Trims a suffix from a string.

```
Hello
```

### split

Splits a string by separator.

```
One
Two
Three
```

### join

Joins strings with separator.

```
One, Two, Three
```

```

db.Exec("INSERT INTO table VALUES (?, ?)",
	input.Name,
	input.Age,
)
```

### repeat

Repeats a string with optional delimiter n times.

```
INSERT INTO table VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
```

### vars

Returns all vars from a package.

```

Variable
```

### consts

Returns all consts from a package.

```

Constant
```

### types

Returns all types from a package.

```

MyInt
MyArray
```

### funcs

Retrieves all functions from a package.

```

main
Echo

```

### methods

Retrieves all methods from a package.

```

Method
PointerMethod

```

### interfaces

Retrieves all interfaces from a package.

```

Interface

```

### structs

Retrieves all structs from a package.

```

Struct
```

### allvars

Returns all vars from a package.

```

Variable
```

### allconsts

Returns all consts from a package.

```

Constant
```

### alltypes

Returns all types from a package.

```

MyInt
MyArray
```

### allfuncs

Retrieves all functions from a package.

```

main
Echo
```

### allmethods

Retrieves all methods from a package.

```

Method
PointerMethod
```

### allinterfaces

Retrieves all interfaces from a package.

```

Interface
```

### allstructs

Retrieves all structs from a package.

```

Struct
```
