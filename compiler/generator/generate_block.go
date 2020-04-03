package generator

import (
	. "lxa/compiler/ast"
)

func (fi *funcInfo) generateBlock(node *Block) {
	for _, stat := range node.Statements {
		fi.generateStatement(stat)
	}

	if node.ReturnExps != nil {
		fi.generateReturnStat(node.ReturnExps, node.LastLine)
	}
}

func (fi *funcInfo) generateReturnStat(exp []Expression, lastLine int) {
	nExps := len(exp)
	if nExps == 0 {
		fi.emitReturn(lastLine, 0, 0)
		return
	}

	if nExps == 1 {
		if nameExp, ok := exp[0].(*NameExp); ok {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				fi.emitReturn(lastLine, r, 1)
				return
			}
		}
		if fcExp, ok := exp[0].(*FuncCallExp); ok {
			r := fi.preAllocReg()
			fi.generateTailCallExp(fcExp, r)
			fi.emitReturn(lastLine, r, -1)
			return
		}
	}

	multRet := isVarargOrFuncCall(exp[nExps-1])
	for i, exp := range exp {
		r := fi.preAllocReg()
		if i == nExps-1 && multRet {
			fi.generateExpression(exp, r, -1)
		} else {
			fi.generateExpression(exp, r, 1)
		}
		fi.checkAllocReg(r)
	}
	fi.freeRegs(nExps)

	a := fi.usedRegs // correct?
	if multRet {
		fi.emitReturn(lastLine, a, -1)
	} else {
		fi.emitReturn(lastLine, a, nExps)
	}
}
