package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"lxa/compiler/lexer"
)

func TestParser(t *testing.T) {
	fmt.Println("parsr test")
	code := `
	
	print("Hello World! by lxa")
	
	for i := 1; i <= 100; i++ {
		print("yes!" + i)
	}

	for k, v in ipairs(map) {
		print(k + v)
	}

	local a = 4
	local m, n = 0, 0
	while d := 3; d < 10 {
		d++
		c <<= 1
	}

	func test(a, b) {
		return a + b, a - b
	}
	
	test(a, 10)

	local add = func(a, b) {
		print(a + b)
		return a + b
	}

	sub := func(a, b) {
		print(a - b)
		return a - b
	}

	div := (x, y) => {
		if test := 1; y == 0 {
			return nil, false
		} else if y < 0 {
			return -x / -y, true
		} else { return x / y, false }
	}

	mm, ok := div(100, 10)

	if ans, ok := div(5, 9); ok {
		print(ans)
	}

	// comment test

	/* in comment
	// lambda test
	sqr := (x) => {
		print(x**2)
		return x**2
	}
	*/

	// lambda test
	sqr := (x) => {
		print(x**2)
		return x**2
	}

	local func Varargfunc(...) {
		p = ...
		return nil
	}

	Varargfunc()

	g = 100
	`
	l := lexer.New(code, "testcode")
	p := New(l)
	ast := p.Parse()
	b, err := json.Marshal(ast)
	if err != nil {
		panic(err)
	}
	println(string(b))
}
