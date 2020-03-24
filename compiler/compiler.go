package compiler

import (
	"lxa/binchunk"
	"lxa/compiler/generator"
	"lxa/compiler/lexer"
	"lxa/compiler/parser"
)

func Compile(chunk, chunkName string) *binchunk.Prototype {
	l := lexer.New(chunk, chunkName)
	p := parser.New(l)
	ast := p.Parse()
	proto := generator.GenerateProto(ast)
	setSource(proto, "@"+chunkName)
	return proto
}

func setSource(proto *binchunk.Prototype, chunkName string) {
	proto.Source = chunkName
	for _, f := range proto.Protos {
		setSource(f, chunkName)
	}
}
