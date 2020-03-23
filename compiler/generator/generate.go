package generator

import (
	. "lxa/binchunk"
	. "lxa/compiler/ast"
)

func GenerateProto(chunk *Block) *Prototype {
	fd := &FuncDefExp{
		LastLine: chunk.LastLine,
		IsVararg: true,
		Block:    chunk,
	}

	fi := newFuncInfo(nil, fd)
	fi.addLocVar("_ENV", 0)
	fi.generateFuncDefExp(fd, 0)
	return fi.subFuncs[0].toProto()
}
