package lexer

import (
	"fmt"
	"testing"

	. "lxa/compiler/token"
)

func TestLexer(t *testing.T) {
	fmt.Println("lexer test")
	code := `



	firsttoken secound_token 3rdToken
	func aa:dd(a) {
		d = d + 1
		c += 4 << 5
		e /= 7 ~/ 8.8; f = g != h
		p = .5
		eee = 1E-10
		for k, v in table {
			blah = (4 * 5) ^ 7 ** 8 ? 55 % 3
		}
		if a and b {
			c = 0x7f7f7f7f
			o = "中文字 符串"
			oo = '的ddd d d ddd △☺…×☀'
			ooo = '1\t2\t3\t4'
			oooo = 18'b'"79g'\' \
			888\
			999"
		    if 可以 {
				print "可以"
			} else {
				print '不可以'
			}
		}
	}

	c := b ? 0
	not
	nto

	range
	identifer
	for end
	is 
	!
	
	
	`
	l := New(code, "testcode")
	peektoken := l.PeekToken()
	fmt.Println("Peek First Token:", peektoken)
	fmt.Println("Again Peek First Token:", l.PeekToken())
	fmt.Println("Again Peek First Token:", l.PeekToken())
	fmt.Println("Again Peek First Token:", l.PeekToken())
	fmt.Println("Again Peek First Token:", l.PeekToken())

	fmt.Println("Peek Second Token:", l.PeekTokenN(2))
	fmt.Println("Peek Forth Token:", l.PeekTokenN(4))
	fmt.Println("Again Peek Forth Token:", l.PeekTokenN(4))
	fmt.Println("Peek Third Token:", l.PeekTokenN(3))
	fmt.Println("Again Peek First Token:", l.PeekToken())

	fmt.Println("Peek Token Of { :", l.PeekTokenOfType(TOKEN_SEP_LCURLY))
	fmt.Println("Peek Token Of EOF :", l.PeekTokenOfType(TOKEN_EOF))

	fmt.Println("Peek Next 17th Token:", l.PeekTokenN(17))
	for token := l.NextToken(); token.Type != TOKEN_EOF; token = l.NextToken() {
		fmt.Println(token)
	}
	fmt.Println("finished")
}
