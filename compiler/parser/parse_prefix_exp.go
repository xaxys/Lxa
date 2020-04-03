package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
)

// prefixexp ::= var | functioncall | '(' exp ')'
// var ::=  Name | prefixexp '[' exp ']' | prefixexp '.' Name
// functioncall ::=  prefixexp args | prefixexp ':' Name args

/*
prefixexp ::= Name
	| '(' exp ')'
	| prefixexp '[' exp ']'
	| prefixexp '.' Name
	| prefixexp [':' Name] args
*/
func (p *Parser) parsePrefixExp() Expression {
	var exp Expression
	if p.lexer.PeekToken().Is(TOKEN_IDENTIFIER) {
		name := p.lexer.NextIdentifier() // Name
		exp = &NameExp{
			Line: name.Line,
			Name: name.Literal,
		}
	} else { // '(' exp ')'
		exp = p.parseParensExpOrLambda()
		if _, ok := exp.(*FuncDefExp); ok {
			return exp
		}
	}

	for {
		switch p.lexer.PeekToken().Type {
		case TOKEN_SEP_LBRACK: // prefixexp '[' exp ']'
			p.lexer.NextToken()                       // '['
			keyExp := p.parseExp()                    // exp
			p.lexer.NextTokenOfType(TOKEN_SEP_RBRACK) // ']'
			lastLine := p.lexer.Line()
			exp = &TableAccessExp{
				LastLine:  lastLine,
				PrefixExp: exp,
				KeyExp:    keyExp,
			}
		case TOKEN_SEP_DOT: // prefixexp '.' Name
			p.lexer.NextToken()              // '.'
			name := p.lexer.NextIdentifier() // Name
			keyExp := &StringExp{
				Line: name.Line,
				Str:  name.Literal,
			}
			exp = &TableAccessExp{
				LastLine:  name.Line,
				PrefixExp: exp,
				KeyExp:    keyExp,
			}
		case TOKEN_SEP_COLON, // prefixexp ':' Name args
			TOKEN_SEP_LPAREN, // (
			// TOKEN_SEP_LCURLY, // { Unknown why should parse { which would cause a conflict with '{' block '}'
			TOKEN_STRING: // prefixexp args
			nameExp := p.parseNameExp()
			line := p.lexer.Line() // todo
			args := p.parseArgs()
			lastLine := p.lexer.Line()
			return &FuncCallExp{
				Line:      line,
				LastLine:  lastLine,
				PrefixExp: exp,
				NameExp:   nameExp,
				Args:      args,
			}
		default:
			return exp
		}
	}
}

func (p *Parser) parseParensExpOrLambda() Expression {
	peeks := p.lexer.PeekTokenOfType(TOKEN_SEP_RPAREN) // )
	idx := len(peeks) + 1
	if TOKEN_SEP_COMMA.In(peeks...) || p.lexer.PeekTokenN(idx).Is(TOKEN_OP_ARROW) { // (name, ...) | (name) =>
		return p.parseLambda()
	} else {
		return p.parseParensExp()
	}
}

func (p *Parser) parseParensExp() Expression {
	p.lexer.NextTokenOfType(TOKEN_SEP_LPAREN) // (
	exp := p.parseExp()                       // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_RPAREN) // )

	switch exp.(type) {
	case *VarargExp, *FuncCallExp, *NameExp, *TableAccessExp:
		return &ParensExp{Exp: exp}
	}

	// no need to keep parens
	return exp
}

func (p *Parser) parseNameExp() *StringExp {
	if p.lexer.PeekToken().Is(TOKEN_SEP_COLON) {
		p.lexer.NextToken() // :
		name := p.lexer.NextIdentifier()
		return &StringExp{
			Line: name.Line,
			Str:  name.Literal,
		}
	}
	return nil
}

// args ::=  '(' [explist] ')' | tableconstructor | LiteralString
func (p *Parser) parseArgs() []Expression {
	var args []Expression
	switch p.lexer.PeekToken().Type {
	case TOKEN_SEP_LPAREN: // '(' [explist] ')'
		p.lexer.NextToken() // (
		if !p.lexer.PeekToken().Is(TOKEN_SEP_RPAREN) {
			args = p.parseExpList()
		}
		p.lexer.NextToken() // )
	case TOKEN_SEP_LCURLY: // '{' [fieldlist] '}'
		args = []Expression{p.parseTableConstructorExp()}
	default: // LiteralString
		str := p.lexer.NextTokenOfType(TOKEN_STRING)
		args = []Expression{&StringExp{
			Line: str.Line,
			Str:  str.Literal,
		}}
	}
	return args
}
