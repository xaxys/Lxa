package generator

import (
	. "lxa/compiler/ast"
)

func isVarargOrFuncCall(exp Expression) bool {
	switch exp.(type) {
	case *VarargExp, *FuncCallExp:
		return true
	}
	return false
}

func removeTailNils(exps []Expression) []Expression {
	for n := len(exps) - 1; n >= 0; n-- {
		if _, ok := exps[n].(*NilExp); !ok {
			return exps[0 : n+1]
		}
	}
	return nil
}

func lineOf(exp Expression) int {
	switch x := exp.(type) {
	case *NilExp:
		return x.Line
	case *TrueExp:
		return x.Line
	case *FalseExp:
		return x.Line
	case *IntegerExp:
		return x.Line
	case *FloatExp:
		return x.Line
	case *StringExp:
		return x.Line
	case *VarargExp:
		return x.Line
	case *NameExp:
		return x.Line
	case *FuncDefExp:
		return x.Line
	case *FuncCallExp:
		return x.Line
	case *TableConstructorExp:
		return x.Line
	case *UnopExp:
		return x.Op.Line
	case *TableAccessExp:
		return lineOf(x.PrefixExp)
	case *BinopExp:
		return lineOf(x.Exp1)
	default:
		panic("unreachable!")
	}
}

func lastLineOf(exp Expression) int {
	switch x := exp.(type) {
	case *NilExp:
		return x.Line
	case *TrueExp:
		return x.Line
	case *FalseExp:
		return x.Line
	case *IntegerExp:
		return x.Line
	case *FloatExp:
		return x.Line
	case *StringExp:
		return x.Line
	case *VarargExp:
		return x.Line
	case *NameExp:
		return x.Line
	case *FuncDefExp:
		return x.LastLine
	case *FuncCallExp:
		return x.LastLine
	case *TableConstructorExp:
		return x.LastLine
	case *TableAccessExp:
		return x.LastLine
	case *BinopExp:
		return lastLineOf(x.Exp2)
	case *UnopExp:
		return lastLineOf(x.Exp)
	default:
		panic("unreachable!")
	}
}
