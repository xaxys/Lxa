package lexer

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	. "lxa/compiler/token"
)

var keywords = map[string]TokenType{
	"and":      TOKEN_OP_AND,
	"break":    TOKEN_KW_BREAK,
	"continue": TOKEN_KW_CONTINUE,
	"else":     TOKEN_KW_ELSE,
	"false":    TOKEN_KW_FALSE,
	"for":      TOKEN_KW_FOR,
	"func":     TOKEN_KW_FUNC,
	"if":       TOKEN_KW_IF,
	"in":       TOKEN_KW_IN,
	"local":    TOKEN_KW_LOCAL,
	"nil":      TOKEN_KW_NIL,
	"not":      TOKEN_OP_NOT,
	"or":       TOKEN_OP_OR,
	"return":   TOKEN_KW_RETURN,
	"true":     TOKEN_KW_TRUE,
	"while":    TOKEN_KW_WHILE,
}

type Lexer struct {
	chunk       string
	chunkName   string
	currentLine int
	line        int
	peekPos     int
	tokenCache  []*Token
	syntaxError []string
}

func New(chunk string, chunkName string) *Lexer {
	return &Lexer{
		chunk:       chunk,
		chunkName:   chunkName,
		currentLine: 1,
		line:        1,
	}
}

func (l *Lexer) Line() int {
	return l.currentLine
}

func (l *Lexer) ChunkName() string {
	return l.chunkName
}

func (l *Lexer) SyntaxError() []string {
	return l.syntaxError
}

func (l *Lexer) NextIdentifier() *Token {
	return l.NextTokenOfType(TOKEN_IDENTIFIER)
}

func (l *Lexer) NextTokenOfType(tokenType ...TokenType) *Token {
	token := l.NextToken()
	if !token.Is(tokenType...) {
		l.Error("unexpected symbol near '%s', expect '%s', but got '%s'",
			token.Literal, tokenTypeString(tokenType), token.Literal)
	}
	return token
}

// PeekTokenN returns a slice of token before specified type
func (l *Lexer) PeekTokenOfType(tokenType ...TokenType) []*Token {
	for i, t := range l.tokenCache {
		if t.Is(tokenType...) {
			return l.tokenCache[0:i]
		}
	}
	i := len(l.tokenCache) + 1
	token := l.PeekTokenN(i)
	for !token.Is(tokenType...) && !token.Is(TOKEN_EOF) {
		i += 1
		token = l.PeekTokenN(i)
	}
	if token.Is(TOKEN_EOF) && !token.Is(tokenType...) {
		l.Error("unexpected symbol near '%s', expect '%s', but got '%s'",
			token.Literal, tokenTypeString(tokenType), token.Literal)
	}
	return l.tokenCache[0:i]
}

// PeekTokenN returns next nth token
func (l *Lexer) PeekTokenN(n int) *Token {
	if len(l.tokenCache) >= n {
		return l.tokenCache[n-1]
	}
	for len(l.tokenCache) < n {
		token := l.nextToken(false)
		l.tokenCache = append(l.tokenCache, token)
		if token.Type.Is(TOKEN_EOF) {
			return token
		}
	}
	return l.tokenCache[n-1]
}

// PeekToken returns next one Token.
func (l *Lexer) PeekToken() *Token {
	if len(l.tokenCache) > 0 {
		return l.tokenCache[0]
	}
	l.tokenCache = append(l.tokenCache, l.nextToken(false))
	return l.tokenCache[0]
}

func (l *Lexer) NextToken() *Token {
	token := l.nextToken(true)
	l.currentLine = token.Line
	return token
}

func (l *Lexer) nextToken(useCache bool) *Token {
	if useCache && len(l.tokenCache) > 0 {
		token := l.tokenCache[0]
		l.tokenCache = l.tokenCache[1:]
		return token
	}

	l.skipWhitespaceAndComment()
	if len(l.chunk) == 0 {
		return &Token{l.line, TOKEN_EOF, "EOF"}
	}

	ch := l.peekChar()

	switch ch {
	case '\n': // peek: \n
		l.read(1)
		l.line++
		return &Token{l.line - 1, TOKEN_SEP_EOLN, "<end-of-line>"}
	case ';': // peek: ;
		l.read(1)
		return &Token{l.line, TOKEN_SEP_SEMI, ";"}
	case ',': // peek: ,
		l.read(1)
		return &Token{l.line, TOKEN_SEP_COMMA, ","}
	case '(': // peek: (
		l.read(1)
		return &Token{l.line, TOKEN_SEP_LPAREN, "("}
	case ')': // peek: )
		l.read(1)
		return &Token{l.line, TOKEN_SEP_RPAREN, ")"}
	case '[': // peek: [
		l.read(1)
		return &Token{l.line, TOKEN_SEP_LBRACK, "["}
	case ']': // peek: ]
		l.read(1)
		return &Token{l.line, TOKEN_SEP_RBRACK, "]"}
	case '{': // peek: {
		l.read(1)
		return &Token{l.line, TOKEN_SEP_LCURLY, "{"}
	case '}': // peek: }
		l.read(1)
		return &Token{l.line, TOKEN_SEP_RCURLY, "}"}
	case ':':
		if l.peekChar() == '=' { // peek: :=
			l.read(2)
			return &Token{l.line, TOKEN_OP_LOCASSIGN, ":="}
		} else { // peek: :
			l.read(1)
			return &Token{l.line, TOKEN_SEP_COLON, ":"}
		}
	case '+':
		switch l.peekChar() {
		case '+': // peek: ++
			l.read(2)
			return &Token{l.line, TOKEN_OP_ADDSELF, "++"}
		case '=': // peek: +=
			l.read(2)
			return &Token{l.line, TOKEN_OP_ADDEQ, "+="}
		default: // peek: +
			l.read(1)
			return &Token{l.line, TOKEN_OP_ADD, "+"}
		}
	case '-':
		switch l.peekChar() {
		case '-': // peek: --
			l.read(2)
			return &Token{l.line, TOKEN_OP_SUBSELF, "--"}
		case '=': // peek: -=
			l.read(2)
			return &Token{l.line, TOKEN_OP_SUBEQ, "-="}
		default: // peek: -
			l.read(1)
			return &Token{l.line, TOKEN_OP_MINUS, "-"}
		}
	case '*':
		switch l.peekChar() {
		case '*':
			if l.peekChar() == '=' { // peek: **=
				l.read(3)
				return &Token{l.line, TOKEN_OP_POWEQ, "**="}
			} else { // peek: **
				l.read(2)
				return &Token{l.line, TOKEN_OP_POW, "**"}
			}
		case '=': // peak: *=
			l.read(2)
			return &Token{l.line, TOKEN_OP_MULEQ, "*="}
		default: // peek: *
			l.read(1)
			return &Token{l.line, TOKEN_OP_MUL, "*"}
		}
	case '/':
		if l.peekChar() == '=' { // peek: /=
			l.read(2)
			return &Token{l.line, TOKEN_OP_DIVEQ, "/="}
		} else { // peek: /
			l.read(1)
			return &Token{l.line, TOKEN_OP_DIV, "/"}
		}
	case '~':
		if l.peekChar() == '/' {
			if l.peekChar() == '=' { // peek: ~/=
				l.read(3)
				return &Token{l.line, TOKEN_OP_IDIVEQ, "~/="}
			} else { // peek: ~/
				l.read(2)
				return &Token{l.line, TOKEN_OP_IDIV, "~/"}
			}
		} else { // peek: ~
			l.read(1)
			return &Token{l.line, TOKEN_OP_BNOT, "~"}
		}
	case '%':
		if l.peekChar() == '=' { // peek: %=
			l.read(2)
			return &Token{l.line, TOKEN_OP_MODEQ, "%="}
		} else { // peek: %
			l.read(1)
			return &Token{l.line, TOKEN_OP_MOD, "%"}
		}
	case '&':
		switch l.peekChar() {
		case '&': // peek: &&
			l.read(2)
			return &Token{l.line, TOKEN_OP_AND, "&&"}
		case '=': // peek: &=
			l.read(2)
			return &Token{l.line, TOKEN_OP_BANDEQ, "&="}
		default: // peek: &
			l.read(1)
			return &Token{l.line, TOKEN_OP_BAND, "&"}
		}
	case '|':
		switch l.peekChar() {
		case '|': // peek: ||
			l.read(2)
			return &Token{l.line, TOKEN_OP_OR, "||"}
		case '=': // peek: |=
			l.read(2)
			return &Token{l.line, TOKEN_OP_BOREQ, "|="}
		default: // peek: |
			l.read(1)
			return &Token{l.line, TOKEN_OP_BOR, "|"}
		}
	case '^':
		if l.peekChar() == '=' { // peek: ^=
			l.read(2)
			return &Token{l.line, TOKEN_OP_BXOREQ, "^="}
		} else { // peek: ^
			l.read(1)
			return &Token{l.line, TOKEN_OP_BXOR, "^"}
		}
	case '#': // peek: #
		l.read(1)
		return &Token{l.line, TOKEN_OP_LEN, "#"}
	case '!':
		if l.peekChar() == '=' { // peek: !=
			l.read(2)
			return &Token{l.line, TOKEN_OP_NE, "!="}
		} else { // peek: !
			l.read(1)
			return &Token{l.line, TOKEN_OP_NOT, "!"}
		}
	case '?': // peek: ?
		l.read(1)
		return &Token{l.line, TOKEN_OP_QST, "?"}
	case '=':
		switch l.peekChar() {
		case '>': // peek: =>
			l.read(2)
			return &Token{l.line, TOKEN_OP_ARROW, "=>"}
		case '=': // peek: ==
			l.read(2)
			return &Token{l.line, TOKEN_OP_EQ, "=="}
		default: // peek: =
			l.read(1)
			return &Token{l.line, TOKEN_OP_ASSIGN, "="}
		}
	case '<':
		switch l.peekChar() {
		case '<':
			if l.peekChar() == '=' { // peek: <<=
				l.read(3)
				return &Token{l.line, TOKEN_OP_SHLEQ, "<<="}
			} else { // peek: <<
				l.read(2)
				return &Token{l.line, TOKEN_OP_SHL, "<<"}
			}
		case '=': // peek: <=
			l.read(2)
			return &Token{l.line, TOKEN_OP_LE, "<="}
		default: // peek: <
			l.read(1)
			return &Token{l.line, TOKEN_OP_LT, "<"}
		}
	case '>':
		switch l.peekChar() {
		case '>':
			if l.peekChar() == '=' { // peek: >>=
				l.read(3)
				return &Token{l.line, TOKEN_OP_SHREQ, ">>="}
			} else { // peek: >>
				l.read(2)
				return &Token{l.line, TOKEN_OP_SHR, ">>"}
			}
		case '=': // peek: >=
			l.read(2)
			return &Token{l.line, TOKEN_OP_GE, ">="}
		default: // peek: >
			l.read(1)
			return &Token{l.line, TOKEN_OP_GT, ">"}
		}
	case '.':
		if l.peekChar() == '.' {
			if l.peekChar() == '.' { // peek: ...
				l.read(3)
				return &Token{l.line, TOKEN_VARARG, "..."}
			} else { // peek: ..
				l.read(2)
				return &Token{l.line, TOKEN_OP_CONCAT, ".."}
			}
		} else if len(l.chunk) == 1 || !unicode.IsDigit(l.peekChar()) { // peek: .
			l.read(1)
			return &Token{l.line, TOKEN_SEP_DOT, "."}
		}
	case '\'', '"': // peek: '[STRING]' "[STRING]"
		return &Token{l.line, TOKEN_STRING, l.readShortString(string(ch))}
	case '`': // peek: `[STRING]`
		return &Token{l.line, TOKEN_STRING, l.readLongString()}
	}

	if ch == '.' || unicode.IsDigit(ch) {
		token := l.readNumber()
		return &Token{l.line, TOKEN_NUMBER, token}
	}
	if ch == '_' || unicode.IsLetter(ch) {
		token := l.readIdentifier()
		if kind, ok := keywords[token]; ok {
			return &Token{l.line, kind, token} // keyword
		} else {
			return &Token{l.line, TOKEN_IDENTIFIER, token}
		}
	}

	l.read(1)
	l.Error("unexpected symbol near %s", string(ch))
	return &Token{l.line, TOKEN_ILLEGAL, string(ch)}
}

func (l *Lexer) skipWhitespaceAndComment() {
	f := true
	for len(l.chunk) > 0 && f {
		switch l.peekChar() {
		case ' ', '\t', '\r', '\f', '\v': // peek: ' ', \t, \r, \f, \v
			l.read(1)
		case '/':
			switch l.peekChar() {
			case '/': // peek: //
				l.read(2)
				l.skipLine()
			case '*': // peek: /*
				l.read(2)
				l.skipLongComment()
			default:
				f = false
			}
		default:
			f = false
		}
	}
	l.peekReset()
}

var reIdentifier = regexp.MustCompile("^[_\\d\\w\u0080-\u07FF\u0800-\uFFFF]+")

func (l *Lexer) readIdentifier() string {
	return l.scan(reIdentifier)
}

var reNumber = regexp.MustCompile(`^0[xX][0-9a-fA-F]*(\.[0-9a-fA-F]*)?([pP][+\-]?[0-9]+)?|^[0-9]*(\.[0-9]*)?([eE][+\-]?[0-9]+)?`)

func (l *Lexer) readNumber() string {
	return l.scan(reNumber)
}

func (l *Lexer) scan(re *regexp.Regexp) string {
	if token := re.FindString(l.chunk); token != "" {
		l.readRaw(len(token))
		return token
	}
	l.Error("unreachable!")
	return ""
}

func (l *Lexer) skipLine() {
	endlineIdx := strings.Index(l.chunk, "\n")
	l.readRaw(endlineIdx + 1)
	l.line += 1
}

func (l *Lexer) skipLongComment() {
	closingIdx := strings.Index(l.chunk, "*/")
	if closingIdx < 0 {
		l.Error("unfinished comment")
	}
	s := l.chunk[0:closingIdx]
	l.readRaw(closingIdx + 2)
	l.line += strings.Count(s, "\n")
}

func (l *Lexer) peekChar() rune {
	c, n := utf8.DecodeRuneInString(l.chunk[l.peekPos:])
	if n == 0 {
		l.Error("invalid character!")
	}
	l.peekPos++
	return c
}

func (l *Lexer) peekReset() {
	l.peekPos = 0
}

func (l *Lexer) read(n int) {
	for i := 0; i < n; i++ {
		l.readChar()
	}
}

func (l *Lexer) readRaw(n int) {
	l.chunk = l.chunk[n:]
	l.peekReset()
}

func (l *Lexer) readChar() rune {
	c, n := utf8.DecodeRuneInString(l.chunk)
	if n == 0 {
		l.Error("invalid character!")
	}
	l.readRaw(n)
	return c
}

func (l *Lexer) readLongString() string {
	l.readRaw(1)
	closingIdx := strings.Index(l.chunk, "`")
	if closingIdx < 0 {
		l.Error("unfinished long string")
	}
	s := l.chunk[0:closingIdx]
	l.readRaw(closingIdx + 1)
	l.line += strings.Count(s, "\n")
	if len(s) > 0 && s[0] == '\n' {
		s = s[1:]
	}
	return s
}

var reNewLine = regexp.MustCompile("\r\n|\n\r|\n|\r")
var reShortStr = regexp.MustCompile(`(?s)(^'(\\\\|\\'|\\\n|\\z\s*|[^'\n])*')|(^"(\\\\|\\"|\\\n|\\z\s*|[^"\n])*")`)

func (l *Lexer) readShortString(end string) string {
	if s := reShortStr.FindString(l.chunk); s != "" {
		l.readRaw(len(s))
		s = s[1 : len(s)-1]
		if strings.Index(s, `\`) >= 0 {
			l.line += len(reNewLine.FindAllString(s, -1))
			s = l.escape(s)
		}
		return s
	}
	l.Error("unfinished string")
	return ""
}

var reDecEscapeSeq = regexp.MustCompile(`^\\[0-9]{1,3}`)
var reHexEscapeSeq = regexp.MustCompile(`^\\x[0-9a-fA-F]{2}`)
var reUnicodeEscapeSeq = regexp.MustCompile(`^\\u\{[0-9a-fA-F]+\}`)

func (l *Lexer) escape(s string) string {
	var buf bytes.Buffer

	for len(s) > 0 {
		if s[0] != '\\' {
			buf.WriteByte(s[0])
			s = s[1:]
			continue
		}

		if len(s) == 1 {
			l.Error("unfinished string")
		}

		switch s[1] {
		case 'a':
			buf.WriteByte('\a')
			s = s[2:]
			continue
		case 'b':
			buf.WriteByte('\b')
			s = s[2:]
			continue
		case 'f':
			buf.WriteByte('\f')
			s = s[2:]
			continue
		case 'n', '\n':
			buf.WriteByte('\n')
			s = s[2:]
			continue
		case 'r':
			buf.WriteByte('\r')
			s = s[2:]
			continue
		case 't':
			buf.WriteByte('\t')
			s = s[2:]
			continue
		case 'v':
			buf.WriteByte('\v')
			s = s[2:]
			continue
		case '"':
			buf.WriteByte('"')
			s = s[2:]
			continue
		case '\'':
			buf.WriteByte('\'')
			s = s[2:]
			continue
		case '\\':
			buf.WriteByte('\\')
			s = s[2:]
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // \ddd
			if found := reDecEscapeSeq.FindString(s); found != "" {
				d, _ := strconv.ParseInt(found[1:], 10, 32)
				if d <= 0xFF {
					buf.WriteByte(byte(d))
					s = s[len(found):]
					continue
				}
				l.Error("decimal escape too large near '%s'", found)
			}
		case 'x': // \xXX
			if found := reHexEscapeSeq.FindString(s); found != "" {
				d, _ := strconv.ParseInt(found[2:], 16, 32)
				buf.WriteByte(byte(d))
				s = s[len(found):]
				continue
			}
		case 'u': // \u{XXX}
			if found := reUnicodeEscapeSeq.FindString(s); found != "" {
				d, err := strconv.ParseInt(found[3:len(found)-1], 16, 32)
				if err == nil && d <= 0x10FFFF {
					buf.WriteRune(rune(d))
					s = s[len(found):]
					continue
				}
				l.Error("UTF-8 value too large near '%s'", found)
			}
		case 'z':
			s = s[2:]
			for len(s) > 0 && isWhiteSpace(s[0]) { // todo
				s = s[1:]
			}
			continue
		}
		l.Error("invalid escape sequence near '\\%c'", s[1])
	}

	return buf.String()
}

func isWhiteSpace(c byte) bool {
	switch c {
	case '\t', '\n', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func (l *Lexer) Error(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	fmt.Fprintln(os.Stderr, "Syntax Error Occurred:")
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error @ %s:%d: %s", l.ChunkName(), l.Line(), err))
	os.Exit(0)
}

func tokenTypeString(tokenType []TokenType) string {
	var s string
	for _, t := range tokenType {
		s += fmt.Sprint(t)
	}
	return s
}
