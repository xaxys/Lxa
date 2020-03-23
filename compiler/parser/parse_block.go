package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
)

func (p *Parser) parseBlock() *Block {
	return &Block{
		Statements: p.parseStatements(),
		ReturnExps: p.parseReturnExps(),
		LastLine:   p.lexer.Line(),
	}
}

func (p *Parser) parseStatements() []Statement {
	stats := make([]Statement, 0, 8)
	for !p.lexer.PeekToken().IsReturnOrBlockEnd() {
		stat := p.parseStatement()
		if st, ok := stat.(*Statements); ok {
			stats = append(stats, st.StatList...)
		} else if _, ok := stat.(*EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

func (p *Parser) parseReturnExps() []Expression {
	if !p.lexer.PeekToken().Is(TOKEN_KW_RETURN) {
		return nil
	}

	p.lexer.NextToken()
	switch p.lexer.PeekToken().Type {
	case TOKEN_EOF, TOKEN_SEP_RCURLY:
		return []Expression{}
	case TOKEN_SEP_SEMI, TOKEN_SEP_EOLN:
		p.lexer.NextToken()
		return []Expression{}
	default:
		expList := p.parseExpList()
		if p.lexer.PeekToken().Is(TOKEN_SEP_SEMI, TOKEN_SEP_EOLN) {
			p.lexer.NextToken()
		}
		return expList
	}
}
