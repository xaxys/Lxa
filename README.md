# Lxa

A new programming language based on Lua vm.

I developed it for some learning purpose and It haven't been used in any formal project.

WARNING! Please expect breaking changes and unstable APIs. Most of them are currently at an early, experimental stage.

## Syntax

Lxa has c-formed syntax with lua-based variable and data structure. Anyone who familiar with C/C++/Java/Go can easily use it.

You can follow these examples or look EBNF below directly.

### Lexical Conventions

Basically the same as Lua.

But `&&`,`||`,`!` can be also used as `and`, `or`,`not`

Removed `::`, `goto`, `repeat` ,`until` ,`do`,`elseif`,`end`,`then`

Modified `function` to `func`, `//`(Integer Divide) to `~/`, `~=`(Not Equal) to `!=`, `^`(Pow) to `**`, `~`(Xor) to `^`

Added `+=`,  `-=`, `*=`, `/=`, `~/=`, `%=`, `**=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, `++`, `--`.

Added `?` to judge if a number = 0 or a string length = 0.

(e.g.`if a := 0; a? { print("is 0")}`, or `if a := ""; a? { print("a length is 0")}`)

Use `{` and `}` to recognize code block.

Use `//` and `/* */` to comment.

### Assignment

```lua
a, b, c = 1, 2, 3
a, b = somefunction()
```

The same as Lua, global variable can be assigned directly.

Multiple value assignment is supported.

#### Local Variable Declaration

```lua
local a, b, c = 1, 2, 3
a, b, c := 1, 2, 3 //the same
```

The first statement is the same as Lua.

The second golang-like statement is also supported which performs the same.

### Statement

#### If

```lua
if a := 1; b := 2; testfunc(); a < b {
	a++
} else if c := a; c != d {
	print("c != d")
} else {
	print('else')
}
```

'If' statement can contain multiple assignment(or function call) before the expression, separated by `;`.

The local variable declared by assignments can only be used inside 'if'.

#### While

```lua
while a := 0; a < 100 {
	a++
    if a > 50 {
		break
	}
} 
```

The same as 'if' statement, a 'while' statement can contain multiple assignment(or function call) before the expression, separated by `;`.

#### For

```lua
for i := 0; i < 100; i++ {
	print(i)
    continue
    print("unreachable")
}
```

The same as C/C++/Java/Go, a 'for' statement can only contain one assignment(or function call) before the expression and one after the expression, separated by `;`.

#### ForIn

```lua
for k, v in ipairs(map0) {
    print("Key: "..k.." Value: "..v)
}
```

The same as Lua, a 'forin' statement invoke the function to iterate.

#### Tips

Lxa removed `goto` and `repeat-until` from syntax. But `continue` is added to syntax.

### Function declaration

```lua
func add(a, b) {
	return a + b
}
local sub(a, b) {
	return a - b
}
mul = func(a, b) {
	return a * b
}
div := (a, b) => {
	return a / b
}
```

All Above function declaration are supported(lambda included).

### Other

Other Lua feature are supported.





