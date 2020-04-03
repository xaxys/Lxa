package token

import "fmt"

type Token struct {
	Line    int
	Type    TokenType
	Literal string
}

type TokenType int

// token type
const (
	TOKEN_ILLEGAL      TokenType        = iota - 1 // illegal
	TOKEN_EOF                                      // end-of-file
	TOKEN_VARARG                                   // ...
	TOKEN_SEP_EOLN                                 // end-of-line
	TOKEN_SEP_SEMI                                 // ;
	TOKEN_SEP_COMMA                                // ,
	TOKEN_SEP_DOT                                  // .
	TOKEN_SEP_COLON                                // :
	TOKEN_SEP_LPAREN                               // (
	TOKEN_SEP_RPAREN                               // )
	TOKEN_SEP_LBRACK                               // [
	TOKEN_SEP_RBRACK                               // ]
	TOKEN_SEP_LCURLY                               // {
	TOKEN_SEP_RCURLY                               // }
	TOKEN_OP_ARROW                                 // =>
	TOKEN_OP_CONCAT                                // ..
	TOKEN_OP_ASSIGN                                // =
	TOKEN_OP_LOCASSIGN                             // :=
	TOKEN_OP_MINUS                                 // - (sub or unm)
	TOKEN_OP_SUBEQ                                 // -=
	TOKEN_OP_SUBSELF                               // --
	TOKEN_OP_ADD                                   // +
	TOKEN_OP_ADDEQ                                 // +=
	TOKEN_OP_ADDSELF                               // ++
	TOKEN_OP_MUL                                   // *
	TOKEN_OP_MULEQ                                 // *=
	TOKEN_OP_DIV                                   // /
	TOKEN_OP_DIVEQ                                 // /=
	TOKEN_OP_IDIV                                  // ~/
	TOKEN_OP_IDIVEQ                                // ~/=
	TOKEN_OP_POW                                   // **
	TOKEN_OP_POWEQ                                 // **=
	TOKEN_OP_MOD                                   // %
	TOKEN_OP_MODEQ                                 // %=
	TOKEN_OP_BAND                                  // &
	TOKEN_OP_BANDEQ                                // &=
	TOKEN_OP_BNOT                                  // ~
	TOKEN_OP_BOR                                   // |
	TOKEN_OP_BOREQ                                 // |=
	TOKEN_OP_BXOR                                  // ^
	TOKEN_OP_BXOREQ                                // ^=
	TOKEN_OP_SHR                                   // >>
	TOKEN_OP_SHREQ                                 // >>=
	TOKEN_OP_SHL                                   // <<
	TOKEN_OP_SHLEQ                                 // <<=
	TOKEN_OP_LT                                    // <
	TOKEN_OP_LE                                    // <=
	TOKEN_OP_GT                                    // >
	TOKEN_OP_GE                                    // >=
	TOKEN_OP_EQ                                    // ==
	TOKEN_OP_NE                                    // !=
	TOKEN_OP_QST                                   // ?
	TOKEN_OP_LEN                                   // #
	TOKEN_OP_AND                                   // and &&
	TOKEN_OP_OR                                    // or ||
	TOKEN_OP_NOT                                   // not !
	TOKEN_KW_BREAK                                 // break
	TOKEN_KW_CONTINUE                              // continue
	TOKEN_KW_ELSE                                  // else
	TOKEN_KW_FALSE                                 // false
	TOKEN_KW_FOR                                   // for
	TOKEN_KW_FUNC                                  // func
	TOKEN_KW_IF                                    // if
	TOKEN_KW_IN                                    // in
	TOKEN_KW_LOCAL                                 // local
	TOKEN_KW_NIL                                   // nil
	TOKEN_KW_RETURN                                // return
	TOKEN_KW_TRUE                                  // true
	TOKEN_KW_WHILE                                 // while
	TOKEN_IDENTIFIER                               // identifier
	TOKEN_NUMBER                                   // number literal
	TOKEN_STRING                                   // string literal
	TOKEN_OP_UNM       = TOKEN_OP_MINUS            // unary minus
	TOKEN_OP_SUB       = TOKEN_OP_MINUS
)

func (t *Token) Change() (*Token, bool) {
	switch t.Type {
	case TOKEN_OP_ADDEQ, TOKEN_OP_ADDSELF:
		return &Token{t.Line, TOKEN_OP_ADD, t.Literal}, true
	case TOKEN_OP_SUBEQ, TOKEN_OP_SUBSELF:
		return &Token{t.Line, TOKEN_OP_SUB, t.Literal}, true
	case TOKEN_OP_MULEQ:
		return &Token{t.Line, TOKEN_OP_MUL, t.Literal}, true
	case TOKEN_OP_DIVEQ:
		return &Token{t.Line, TOKEN_OP_DIV, t.Literal}, true
	case TOKEN_OP_IDIVEQ:
		return &Token{t.Line, TOKEN_OP_IDIV, t.Literal}, true
	case TOKEN_OP_POWEQ:
		return &Token{t.Line, TOKEN_OP_POW, t.Literal}, true
	case TOKEN_OP_MODEQ:
		return &Token{t.Line, TOKEN_OP_MOD, t.Literal}, true
	case TOKEN_OP_BANDEQ:
		return &Token{t.Line, TOKEN_OP_BAND, t.Literal}, true
	case TOKEN_OP_BOREQ:
		return &Token{t.Line, TOKEN_OP_BOR, t.Literal}, true
	case TOKEN_OP_BXOREQ:
		return &Token{t.Line, TOKEN_OP_BXOR, t.Literal}, true
	case TOKEN_OP_SHLEQ:
		return &Token{t.Line, TOKEN_OP_SHL, t.Literal}, true
	case TOKEN_OP_SHREQ:
		return &Token{t.Line, TOKEN_OP_SHR, t.Literal}, true
	default:
		return nil, false
	}
}

func (t *Token) Is(tokenType ...TokenType) bool { return t.Type.Is(tokenType...) }

func (tt TokenType) Is(tokenType ...TokenType) bool {
	for _, t := range tokenType {
		if tt == t {
			return true
		}
	}
	return false
}

func (tt TokenType) In(token ...*Token) bool {
	for _, t := range token {
		if tt == t.Type {
			return true
		}
	}
	return false
}

func (tt TokenType) CountIn(token ...*Token) int {
	count := 0
	for _, t := range token {
		if tt == t.Type {
			count++
		}
	}
	return count
}

func (t *Token) IsAssignment() bool { return t.Type.IsAssignment() }

func (tt TokenType) IsAssignment() bool {
	return tt.Is(
		TOKEN_OP_ASSIGN,    // =
		TOKEN_OP_LOCASSIGN, // :=
		TOKEN_OP_SUBEQ,     // -=
		TOKEN_OP_ADDEQ,     // +=
		TOKEN_OP_MULEQ,     // *=
		TOKEN_OP_DIVEQ,     // /=
		TOKEN_OP_IDIVEQ,    // ~/=
		TOKEN_OP_POWEQ,     // **=
		TOKEN_OP_MODEQ,     // %=
		TOKEN_OP_BANDEQ,    // &=
		TOKEN_OP_BOREQ,     // |=
		TOKEN_OP_BXOREQ,    // ^=
		TOKEN_OP_SHREQ,     // >>=
		TOKEN_OP_SHLEQ)     // <<=
}

func (t *Token) IsReturnOrBlockEnd() bool { return t.Type.IsReturnOrBlockEnd() }

func (tt TokenType) IsReturnOrBlockEnd() bool {
	return tt.Is(
		TOKEN_KW_RETURN,  // return
		TOKEN_EOF,        // EOF
		TOKEN_SEP_RCURLY) // }
}

func (t *Token) String() string {
	return fmt.Sprintf("<%s>(Line: %d, Literal: '%s')", t.Type, t.Line, t.Literal)
}

func (t TokenType) String() string {
	switch t {
	case TOKEN_ILLEGAL:
		return "ILLEGAL"
	case TOKEN_EOF:
		return "EOF"
	case TOKEN_VARARG:
		return "..."
	case TOKEN_SEP_EOLN:
		return "\\n"
	case TOKEN_SEP_SEMI:
		return ";"
	case TOKEN_SEP_COMMA:
		return ","
	case TOKEN_SEP_DOT:
		return "."
	case TOKEN_SEP_COLON:
		return ":"
	case TOKEN_SEP_LPAREN:
		return "("
	case TOKEN_SEP_RPAREN:
		return ")"
	case TOKEN_SEP_LBRACK:
		return "["
	case TOKEN_SEP_RBRACK:
		return "]"
	case TOKEN_SEP_LCURLY:
		return "{"
	case TOKEN_SEP_RCURLY:
		return "}"
	case TOKEN_OP_ARROW:
		return "=>"
	case TOKEN_OP_CONCAT:
		return ".."
	case TOKEN_OP_ASSIGN:
		return "="
	case TOKEN_OP_LOCASSIGN:
		return ":="
	case TOKEN_OP_MINUS:
		return "-"
	case TOKEN_OP_SUBEQ:
		return "-="
	case TOKEN_OP_SUBSELF:
		return "--"
	case TOKEN_OP_ADD:
		return "++"
	case TOKEN_OP_ADDEQ:
		return "+="
	case TOKEN_OP_ADDSELF:
		return "++"
	case TOKEN_OP_MUL:
		return "*"
	case TOKEN_OP_MULEQ:
		return "*="
	case TOKEN_OP_DIV:
		return "/"
	case TOKEN_OP_DIVEQ:
		return "/="
	case TOKEN_OP_IDIV:
		return "~/"
	case TOKEN_OP_IDIVEQ:
		return "~/="
	case TOKEN_OP_POW:
		return "**"
	case TOKEN_OP_POWEQ:
		return "**-"
	case TOKEN_OP_MOD:
		return "%"
	case TOKEN_OP_MODEQ:
		return "%="
	case TOKEN_OP_BAND:
		return "&"
	case TOKEN_OP_BANDEQ:
		return "&="
	case TOKEN_OP_BNOT:
		return "~"
	case TOKEN_OP_BOR:
		return "|"
	case TOKEN_OP_BOREQ:
		return "|="
	case TOKEN_OP_BXOR:
		return "^"
	case TOKEN_OP_BXOREQ:
		return "^="
	case TOKEN_OP_SHR:
		return ">>"
	case TOKEN_OP_SHREQ:
		return ">>="
	case TOKEN_OP_SHL:
		return "<<"
	case TOKEN_OP_SHLEQ:
		return "<<="
	case TOKEN_OP_LT:
		return "<"
	case TOKEN_OP_LE:
		return "<="
	case TOKEN_OP_GT:
		return ">"
	case TOKEN_OP_GE:
		return ">="
	case TOKEN_OP_EQ:
		return "=="
	case TOKEN_OP_NE:
		return "!="
	case TOKEN_OP_QST:
		return "?"
	case TOKEN_OP_LEN:
		return "#"
	case TOKEN_OP_AND:
		return "(&&, and)"
	case TOKEN_OP_OR:
		return "(||, or)"
	case TOKEN_OP_NOT:
		return "(!, not)"
	case TOKEN_KW_BREAK:
		return "BREAK"
	case TOKEN_KW_CONTINUE:
		return "CONTINUE"
	case TOKEN_KW_ELSE:
		return "ELSE"
	case TOKEN_KW_FALSE:
		return "FALSE"
	case TOKEN_KW_FOR:
		return "FOR"
	case TOKEN_KW_FUNC:
		return "FUNC"
	case TOKEN_KW_IF:
		return "IF"
	case TOKEN_KW_IN:
		return "IN"
	case TOKEN_KW_LOCAL:
		return "LOCAL"
	case TOKEN_KW_NIL:
		return "NIL"
	case TOKEN_KW_RETURN:
		return "RETURN"
	case TOKEN_KW_TRUE:
		return "TRUE"
	case TOKEN_KW_WHILE:
		return "WHILE"
	case TOKEN_IDENTIFIER:
		return "IDENTIFIER"
	case TOKEN_NUMBER:
		return "NUMBER"
	case TOKEN_STRING:
		return "STRING"
	default:
		return "UNKNOWN"
	}
}
