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
	TOKEN_ILLEGAL     TokenType        = iota - 1 // illegal
	TOKEN_EOF                                     // end-of-file
	TOKEN_VARARG                                  // ...
	TOKEN_SEP_EOLN                                // end-of-line
	TOKEN_SEP_SEMI                                // ;
	TOKEN_SEP_COMMA                               // ,
	TOKEN_SEP_DOT                                 // .
	TOKEN_SEP_COLON                               // :
	TOKEN_SEP_LPAREN                              // (
	TOKEN_SEP_RPAREN                              // )
	TOKEN_SEP_LBRACK                              // [
	TOKEN_SEP_RBRACK                              // ]
	TOKEN_SEP_LCURLY                              // {
	TOKEN_SEP_RCURLY                              // }
	TOKEN_OP_ARROW                                // =>
	TOKEN_OP_CONCAT                               // ..
	TOKEN_OP_ASSIGN                               // =
	TOKEN_OP_LASSIGN                              // :=
	TOKEN_OP_MINUS                                // - (sub or unm)
	TOKEN_OP_SUBEQ                                // -=
	TOKEN_OP_SUBSELF                              // --
	TOKEN_OP_ADD                                  // +
	TOKEN_OP_ADDEQ                                // +=
	TOKEN_OP_ADDSELF                              // ++
	TOKEN_OP_MUL                                  // *
	TOKEN_OP_MULEQ                                // *=
	TOKEN_OP_DIV                                  // /
	TOKEN_OP_DIVEQ                                // /=
	TOKEN_OP_IDIV                                 // ~/
	TOKEN_OP_IDIVEQ                               // ~/=
	TOKEN_OP_POW                                  // **
	TOKEN_OP_POWEQ                                // **=
	TOKEN_OP_MOD                                  // %
	TOKEN_OP_MODEQ                                // %=
	TOKEN_OP_BAND                                 // &
	TOKEN_OP_BANDEQ                               // &=
	TOKEN_OP_BNOT                                 // ~
	TOKEN_OP_BOR                                  // |
	TOKEN_OP_BOREQ                                // |=
	TOKEN_OP_BXOR                                 // ^
	TOKEN_OP_BXOREQ                               // ^=
	TOKEN_OP_SHR                                  // >>
	TOKEN_OP_SHREQ                                // >>=
	TOKEN_OP_SHL                                  // <<
	TOKEN_OP_SHLEQ                                // <<=
	TOKEN_OP_LT                                   // <
	TOKEN_OP_LE                                   // <=
	TOKEN_OP_GT                                   // >
	TOKEN_OP_GE                                   // >=
	TOKEN_OP_EQ                                   // ==
	TOKEN_OP_NE                                   // !=
	TOKEN_OP_QST                                  // ?
	TOKEN_OP_LEN                                  // #
	TOKEN_OP_AND                                  // and &&
	TOKEN_OP_OR                                   // or ||
	TOKEN_OP_NOT                                  // not !
	TOKEN_KW_BREAK                                // break
	TOKEN_KW_CONTINUE                             // continue
	TOKEN_KW_ELSE                                 // else
	TOKEN_KW_FALSE                                // false
	TOKEN_KW_FOR                                  // for
	TOKEN_KW_FUNC                                 // func
	TOKEN_KW_IF                                   // if
	TOKEN_KW_IN                                   // in
	TOKEN_KW_LOCAL                                // local
	TOKEN_KW_NIL                                  // nil
	TOKEN_KW_RETURN                               // return
	TOKEN_KW_TRUE                                 // true
	TOKEN_KW_WHILE                                // while
	TOKEN_IDENTIFIER                              // identifier
	TOKEN_NUMBER                                  // number literal
	TOKEN_STRING                                  // string literal
	TOKEN_OP_UNM      = TOKEN_OP_MINUS            // unary minus
	TOKEN_OP_SUB      = TOKEN_OP_MINUS
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
		TOKEN_OP_ASSIGN,  // =
		TOKEN_OP_LASSIGN, // :=
		TOKEN_OP_SUBEQ,   // -=
		TOKEN_OP_ADDEQ,   // +=
		TOKEN_OP_MULEQ,   // *=
		TOKEN_OP_DIVEQ,   // /=
		TOKEN_OP_IDIVEQ,  // ~/=
		TOKEN_OP_POWEQ,   // **=
		TOKEN_OP_MODEQ,   // %=
		TOKEN_OP_BANDEQ,  // &=
		TOKEN_OP_BOREQ,   // |=
		TOKEN_OP_BXOREQ,  // ^=
		TOKEN_OP_SHREQ,   // >>=
		TOKEN_OP_SHLEQ)   // <<=
}

func (t *Token) IsReturnOrBlockEnd() bool { return t.Type.IsReturnOrBlockEnd() }

func (tt TokenType) IsReturnOrBlockEnd() bool {
	return tt.Is(
		TOKEN_KW_RETURN,  // return
		TOKEN_EOF,        // EOF
		TOKEN_SEP_RCURLY) // }
}

func (t *Token) String() string {
	return fmt.Sprintf("Line: %d, Type: %s, Literal: %s", t.Line, t.Type, t.Literal)
}

func (t TokenType) String() string {
	switch t {
	case TOKEN_ILLEGAL:
		return "TOKEN_ILLEGAL"
	case TOKEN_EOF:
		return "TOKEN_EOF"
	case TOKEN_VARARG:
		return "TOKEN_VARARG"
	case TOKEN_SEP_EOLN:
		return "TOKEN_SEP_EOLN"
	case TOKEN_SEP_SEMI:
		return "TOKEN_SEP_SEMI"
	case TOKEN_SEP_COMMA:
		return "TOKEN_SEP_COMMA"
	case TOKEN_SEP_DOT:
		return "TOKEN_SEP_DOT"
	case TOKEN_SEP_COLON:
		return "TOKEN_SEP_COLON"
	case TOKEN_SEP_LPAREN:
		return "TOKEN_SEP_LPAREN"
	case TOKEN_SEP_RPAREN:
		return "TOKEN_SEP_RPAREN"
	case TOKEN_SEP_LBRACK:
		return "TOKEN_SEP_LBRACK"
	case TOKEN_SEP_RBRACK:
		return "TOKEN_SEP_RBRACK"
	case TOKEN_SEP_LCURLY:
		return "TOKEN_SEP_LCURLY"
	case TOKEN_SEP_RCURLY:
		return "TOKEN_SEP_RCURLY"
	case TOKEN_OP_ARROW:
		return "TOKEN_OP_ARROW"
	case TOKEN_OP_ASSIGN:
		return "TOKEN_OP_ASSIGN"
	case TOKEN_OP_LASSIGN:
		return "TOKEN_OP_LASSIGN"
	case TOKEN_OP_MINUS:
		return "TOKEN_OP_MINUS"
	case TOKEN_OP_SUBEQ:
		return "TOKEN_OP_SUBEQ"
	case TOKEN_OP_ADD:
		return "TOKEN_OP_ADD"
	case TOKEN_OP_ADDEQ:
		return "TOKEN_OP_ADDEQ"
	case TOKEN_OP_MUL:
		return "TOKEN_OP_MUL"
	case TOKEN_OP_MULEQ:
		return "TOKEN_OP_MULEQ"
	case TOKEN_OP_DIV:
		return "TOKEN_OP_DIV"
	case TOKEN_OP_DIVEQ:
		return "TOKEN_OP_DIVEQ"
	case TOKEN_OP_IDIV:
		return "TOKEN_OP_IDIV"
	case TOKEN_OP_IDIVEQ:
		return "TOKEN_OP_IDIVEQ"
	case TOKEN_OP_POW:
		return "TOKEN_OP_POW"
	case TOKEN_OP_POWEQ:
		return "TOKEN_OP_POWEQ"
	case TOKEN_OP_MOD:
		return "TOKEN_OP_MOD"
	case TOKEN_OP_MODEQ:
		return "TOKEN_OP_MODEQ"
	case TOKEN_OP_BAND:
		return "TOKEN_OP_BAND"
	case TOKEN_OP_BANDEQ:
		return "TOKEN_OP_BANDEQ"
	case TOKEN_OP_BNOT:
		return "TOKEN_OP_BNOT"
	case TOKEN_OP_BOR:
		return "TOKEN_OP_BOR"
	case TOKEN_OP_BOREQ:
		return "TOKEN_OP_BOREQ"
	case TOKEN_OP_BXOR:
		return "TOKEN_OP_BXOR"
	case TOKEN_OP_BXOREQ:
		return "TOKEN_OP_BXOREQ"
	case TOKEN_OP_SHR:
		return "TOKEN_OP_SHR"
	case TOKEN_OP_SHREQ:
		return "TOKEN_OP_SHREQ"
	case TOKEN_OP_SHL:
		return "TOKEN_OP_SHL"
	case TOKEN_OP_SHLEQ:
		return "TOKEN_OP_SHLEQ"
	case TOKEN_OP_LT:
		return "TOKEN_OP_LT"
	case TOKEN_OP_LE:
		return "TOKEN_OP_LE"
	case TOKEN_OP_GT:
		return "TOKEN_OP_GT"
	case TOKEN_OP_GE:
		return "TOKEN_OP_GE"
	case TOKEN_OP_EQ:
		return "TOKEN_OP_EQ"
	case TOKEN_OP_NE:
		return "TOKEN_OP_NE"
	case TOKEN_OP_QST:
		return "TOKEN_OP_QST"
	case TOKEN_OP_LEN:
		return "TOKEN_OP_LEN"
	case TOKEN_OP_AND:
		return "TOKEN_OP_AND"
	case TOKEN_OP_OR:
		return "TOKEN_OP_OR"
	case TOKEN_OP_NOT:
		return "TOKEN_OP_NOT"
	case TOKEN_KW_BREAK:
		return "TOKEN_KW_BREAK"
	case TOKEN_KW_CONTINUE:
		return "TOKEN_KW_CONTINUE"
	case TOKEN_KW_ELSE:
		return "TOKEN_KW_ELSE"
	case TOKEN_KW_FALSE:
		return "TOKEN_KW_FALSE"
	case TOKEN_KW_FOR:
		return "TOKEN_KW_FOR"
	case TOKEN_KW_FUNC:
		return "TOKEN_KW_FUNC"
	case TOKEN_KW_IF:
		return "TOKEN_KW_IF"
	case TOKEN_KW_IN:
		return "TOKEN_KW_IN"
	case TOKEN_KW_LOCAL:
		return "TOKEN_KW_LOCAL"
	case TOKEN_KW_NIL:
		return "TOKEN_KW_NIL"
	case TOKEN_KW_RETURN:
		return "TOKEN_KW_RETURN"
	case TOKEN_KW_TRUE:
		return "TOKEN_KW_TRUE"
	case TOKEN_KW_WHILE:
		return "TOKEN_KW_WHILE"
	case TOKEN_IDENTIFIER:
		return "TOKEN_IDENTIFIER"
	case TOKEN_NUMBER:
		return "TOKEN_NUMBER"
	case TOKEN_STRING:
		return "TOKEN_STRING"
	default:
		return "UNKNOWN"
	}
}
