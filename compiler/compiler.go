package compiler

import (
	"lxa/binchunk"
	"lxa/compiler/generator"
	"lxa/compiler/parser"
)

func Compile(chunk, chunkName string) *binchunk.Prototype {
	p := parser.New(chunk, chunkName)
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
