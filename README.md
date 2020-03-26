# Lxa

A new programming language based on Lua vm.

I developed it for some learning purpose, but it haven't been used in any formal project.

WARNING! Please expect breaking changes and unstable APIs. Most of them are currently at an early, experimental stage.

## Update

* 2020/03/23 Released Lxa v0.1.0. 
  * Use Azure/golua as vm.
* 2020/03/25 Released Lxa v0.2.1. 
  * Use Official lua 5.3.5 vm (written in c) as default vm.
  * Added inner go lua vm as a option (several stdlib unsupported yet).
  * Added 'Compile Only' option to output compiled lua bytecode (since v0.2.0).
  * Added linux support and x86 support (auto select static lib when compiling) (untested). 
* 2020/03/26 Released Lxa v0.2.4.
  * Added Debug option to display details in running.
  * Added 'Parse Only' option to watch bytecode instructions.
  * Optimized Logical Expression generation, now multiple `and` and `or` expressions are specially handled apart from Binary Opration Expression.
  * Fixed bug of 0 length lua string in bytecode.

## Syntax

Lxa has c-formed syntax with lua-based variable and data structure. Anyone who is familiar with C/C++/Java/Go can easily use it.

You can follow these examples or look EBNF below directly.

### Lexical Conventions

Free-style Code, basically the same as Lua.

`;` is not necessary in the end of a sentence.

`\n` will be recognized as equal as `;`. 

`&&`,`||`,`!` can be also used as `and`, `or`,`not`

Removed `::`, `goto`, `repeat` ,`until` ,`do`,`elseif`,`end`,`then`

Modified `function` to `func`, `//`(Integer Divide) to `~/`, `~=`(Not Equal) to `!=`, `^`(Pow) to `**`, `~`(Xor) to `^`

Added `+=`,  `-=`, `*=`, `/=`, `~/=`, `%=`, `**=`, `&=`, `|=`, `^=`, `<<=`, `>>=`, `++`, `--`.

Added `?` to judge if a number = 0 or a string length = 0.

(e.g.`if a := 0; a? { print("is 0")}`, or `if a := ""; a? { print("a length is 0")}`)

Use `{` and `}` to recognize code block.

(WARNING: Because `\n` will be recognized as equal as `;`, `{` is not allowed appeared in the next line)

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

### EBNF

```
chunk ::= block

block ::= {stat} [retstat]
retstat ::= return [explist] [';']
explist ::= exp {',' exp}
namelist ::= Name {',' Name}
varlist ::= var {',' var}
var ::=  Name | prefixexp '[' exp ']' | prefixexp '.' Name

assignment ::= assign | locvardecl | functioncall

assign ::= varlist ('+=' | '-=' | '*=' | '/=' | '~/=' | '%='
		| '&=' | '^=' | '|=' | '**=' | '<<=' | '>>=' | '=') explist

locvardecl ::= namelist ':=' explist | local namelist [':=' explist]

exp ::=  nil
	| false
	| true
	| Numeral
	| LiteralString
	| '...'
	| functiondef
	| functioncall
	| prefixexp
	| tableconstructor
	| exp binop exp
	| unop exp
	| varlist ['=' explist]
	| namelist ':=' explist

prefixexp ::= var
	| functioncall
	| '(' exp ')'
	| prefixexp ':' Name args
	| prefixexp args

functioncall ::=  prefixexp args | prefixexp ':' Name args

tableconstructor ::= '{' [fieldlist] '}'
fieldlist ::= field {fieldsep field} [fieldsep]
field ::= '[' exp ']' '=' exp | Name '=' exp | exp
fieldsep ::= ',' | ';'

functiondef ::= func funcbody | lambda
lambda ::= '(' [parlist] ')' '=>' '{' block '}'
funcbody ::= '(' [parlist] ')' '{' block '}'
parlist ::= namelist [',' '...'] | '...'
namelist ::= Name {',' Name}

stat ::= ';'
	| assignment ';'
	| break
	| continue
	| while {assignment ';'} exp '{' block '}'
	| if {assignment ';'} exp '{' block '}' {else if {assignment ';'} exp '{' block '}'} [else '{' block '}']
	| for assignment ';' exp ';' assignment '{' block '}'
	| for namelist in explist '{' block '}'
	| func funcname funcbody
	| local func Name funcbody
```

## About

Contact me: E-mail: gz@oasis.run, QQ: 963796543, WebSite: http://www.oasis.run