package parser

import (
	"fmt"
	. "lxa/compiler/ast"
	"lxa/compiler/lexer"
	. "lxa/compiler/token"
	"os"
)

// Parser represents lexical analyzer struct
type Parser struct {
	lexer *lexer.Lexer

	// curToken  token.Token
	// peekToken token.Token

	// // Determine if call expression should accept block argument,
	// // currently only used when parsing while statement.
	// // However, this is not a very good practice should change it in the future.
	// acceptBlock bool
	// fsm         *fsm.FSM
	// Mode        ParserMode
}

func New(chunk, chunkName string) *Parser {
	p := &Parser{
		lexer: lexer.New(chunk, chunkName),
	}
	return p
}

func (p *Parser) Parse() *Block {
	block := p.parseBlock()
	p.lexer.NextTokenOfType(TOKEN_EOF)
	return block
}

func (p *Parser) Error(f string, a ...interface{}) {
	p.lexer.Error(f, a...)
}

func (p *Parser) Warn(f string, a ...interface{}) {
	err := fmt.Sprintf(f, a...)
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Warning @ %s:%d: %s", p.lexer.ChunkName(), p.lexer.Line(), err))
	os.Exit(0)
}
