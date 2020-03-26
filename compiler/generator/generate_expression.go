package generator

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
	. "lxa/vm"
)

// kind of operands
const (
	ARG_CONST = 1 // const index
	ARG_REG   = 2 // register index
	ARG_UPVAL = 4 // upvalue index
	ARG_BLANK = 8 // blank value
	ARG_RK    = ARG_REG | ARG_CONST
	ARG_RU    = ARG_REG | ARG_UPVAL
	ARG_RUK   = ARG_REG | ARG_UPVAL | ARG_CONST
)

func (fi *funcInfo) generateExpression(node Expression, a, n int) {
	switch exp := node.(type) {
	case *NilExp:
		fi.emitLoadNil(exp.Line, a, n)
	case *FalseExp:
		fi.emitLoadBool(exp.Line, a, 0, 0)
	case *TrueExp:
		fi.emitLoadBool(exp.Line, a, 1, 0)
	case *IntegerExp:
		fi.emitLoadK(exp.Line, a, exp.Val)
	case *FloatExp:
		fi.emitLoadK(exp.Line, a, exp.Val)
	case *StringExp:
		fi.emitLoadK(exp.Line, a, exp.Str)
	case *ParensExp:
		fi.generateExpression(exp.Exp, a, 1)
	case *VarargExp:
		fi.generateVarargExp(exp, a, n)
	case *FuncDefExp:
		fi.generateFuncDefExp(exp, a)
	case *ConcatExp:
		fi.generateConcatExp(exp, a)
	case *TableConstructorExp:
		fi.generateTableConstructorExp(exp, a)
	case *UnopExp:
		fi.generateUnopExp(exp, a)
	case *LogicalExp:
		fi.generateLogicalExp(exp, a)
	case *BinopExp:
		fi.generateBinopExp(exp, a)
	case *NameExp:
		fi.generateNameExp(exp, a)
	case *TableAccessExp:
		fi.generateTableAccessExp(exp, a)
	case *FuncCallExp:
		fi.generateFuncCallExp(exp, a, n)
	}
}

func (fi *funcInfo) generateVarargExp(node *VarargExp, a, n int) {
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	fi.emitVararg(node.Line, a, n)
}

// f[a] := function(args) body end
func (fi *funcInfo) generateFuncDefExp(node *FuncDefExp, a int) {
	subFI := newFuncInfo(fi, node)
	fi.subFuncs = append(fi.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocVar(param, 0)
	}

	subFI.generateBlock(node.Block)
	subFI.exitScope(subFI.pc() + 2)
	subFI.emitReturn(node.LastLine, 0, 0)

	bx := len(fi.subFuncs) - 1
	fi.emitClosure(node.LastLine, a, bx)
}

func (fi *funcInfo) generateTableConstructorExp(node *TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}
	nExps := len(node.KeyExps)
	multRet := nExps > 0 &&
		isVarargOrFuncCall(node.ValExps[nExps-1])

	fi.emitNewTable(node.Line, a, nArr, nExps-nArr)

	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		if keyExp == nil {
			arrIdx++
			tmp := fi.preAllocReg()
			if i == nExps-1 && multRet {
				fi.generateExpression(valExp, tmp, -1)
			} else {
				fi.generateExpression(valExp, tmp, 1)
			}
			fi.checkAllocReg(tmp)

			if arrIdx%50 == 0 || arrIdx == nArr { // LFIELDS_PER_FLUSH
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				fi.freeRegs(n)
				line := lastLineOf(valExp)
				c := (arrIdx-1)/50 + 1 // todo: c > 0xFF
				if i == nExps-1 && multRet {
					fi.emitSetList(line, a, 0, c)
				} else {
					fi.emitSetList(line, a, n, c)
				}
			}

			continue
		}

		b := fi.preAllocReg()
		fi.generateExpression(keyExp, b, 1)
		fi.checkAllocReg(b)
		c := fi.preAllocReg()
		fi.generateExpression(valExp, c, 1)
		fi.checkAllocReg(c)
		fi.freeRegs(2)

		line := lastLineOf(valExp)
		fi.emitSetTable(line, a, b, c)
	}
}

// r[a] := op exp
func (fi *funcInfo) generateUnopExp(node *UnopExp, a int) {
	oldRegs := fi.usedRegs
	b, _ := fi.expToOpArg(node.Exp, ARG_REG)
	fi.emitUnaryOp(node.Op.Line, node.Op.Type, a, b)
	fi.usedRegs = oldRegs
}

// r[a] := (and | or) between expList
func (fi *funcInfo) generateLogicalExp(node *LogicalExp, a int) {
	var Jmps []int
	oldRegs := fi.usedRegs

	b, _ := fi.expToOpArg(node.ExpList[0], ARG_REG)
	fi.usedRegs = oldRegs
	for _, exp := range node.ExpList[1:] {
		if node.Op.Is(TOKEN_OP_AND) {
			fi.emitTestSet(node.Op.Line, a, b, 0)
		} else {
			fi.emitTestSet(node.Op.Line, a, b, 1)
		}
		pcOfJmp := fi.emitJmp(node.Op.Line, 0, 0)
		Jmps = append(Jmps, pcOfJmp)

		b, _ = fi.expToOpArg(exp, ARG_REG)
		fi.usedRegs = oldRegs
	}
	for _, pcOfJmp := range Jmps {
		fi.fixSbx(pcOfJmp, fi.pc()-pcOfJmp)
	}
	if b != a {
		fi.emitMove(node.Op.Line, a, b)
	}
}

// r[a] := exp1 op exp2
func (fi *funcInfo) generateBinopExp(node *BinopExp, a int) {
	oldRegs := fi.usedRegs
	b, _ := fi.expToOpArg(node.Exp1, ARG_RK)
	c, _ := fi.expToOpArg(node.Exp2, ARG_RK)
	fi.emitBinaryOp(node.Op.Line, node.Op.Type, a, b, c)
	fi.usedRegs = oldRegs
}

// r[a] := name
func (fi *funcInfo) generateNameExp(node *NameExp, a int) {
	if r := fi.slotOfLocVar(node.Name); r >= 0 {
		fi.emitMove(node.Line, a, r)
	} else if idx := fi.upvalIndex(node.Name); idx >= 0 {
		fi.emitGetUpval(node.Line, a, idx)
	} else { // x => _ENV['x']
		taExp := &TableAccessExp{
			LastLine: node.Line,
			PrefixExp: &NameExp{
				Line: node.Line,
				Name: "_ENV",
			},
			KeyExp: &StringExp{
				Line: node.Line,
				Str:  node.Name,
			},
		}
		fi.generateTableAccessExp(taExp, a)
	}
}

// r[a] := exp1 .. exp2
func (fi *funcInfo) generateConcatExp(node *ConcatExp, a int) {
	for _, subExp := range node.ExpList {
		a := fi.preAllocReg()
		fi.generateExpression(subExp, a, 1)
		fi.checkAllocReg(a)
	}

	c := fi.usedRegs - 1
	b := c - len(node.ExpList) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(node.Line, OP_CONCAT, a, b, c)
}

// r[a] := prefix[key]
func (fi *funcInfo) generateTableAccessExp(node *TableAccessExp, a int) {
	oldRegs := fi.usedRegs
	b, kindB := fi.expToOpArg(node.PrefixExp, ARG_RU)
	c, _ := fi.expToOpArg(node.KeyExp, ARG_RK)
	fi.usedRegs = oldRegs

	if kindB == ARG_UPVAL {
		fi.emitGetTabUp(node.LastLine, a, b, c)
	} else {
		fi.emitGetTable(node.LastLine, a, b, c)
	}
}

// r[a] := f(args)
func (fi *funcInfo) generateFuncCallExp(node *FuncCallExp, a, n int) {
	nArgs := fi.prepFuncCall(node, a)
	fi.emitCall(node.Line, a, nArgs, n)
}

// return f(args)
func (fi *funcInfo) generateTailCallExp(node *FuncCallExp, a int) {
	nArgs := fi.prepFuncCall(node, a)
	fi.emitTailCall(node.Line, a, nArgs)
}

func (fi *funcInfo) prepFuncCall(node *FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false

	fi.generateExpression(node.PrefixExp, a, 1)
	if node.NameExp != nil {
		fi.allocReg()
		c, k := fi.expToOpArg(node.NameExp, ARG_RK)
		fi.emitSelf(node.Line, a, a, c)
		if k == ARG_REG {
			fi.freeRegs(1)
		}
	}
	fi.checkAllocReg(a)
	for i, arg := range node.Args {
		tmp := fi.preAllocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			fi.generateExpression(arg, tmp, -1)
		} else {
			fi.generateExpression(arg, tmp, 1)
		}
		fi.checkAllocReg(tmp)
	}
	fi.freeRegs(nArgs)

	if node.NameExp != nil {
		fi.freeReg()
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	return nArgs
}

func (fi *funcInfo) expToOpArg(node Expression, argKinds int) (arg, argKind int) {
	if argKinds&ARG_CONST > 0 {
		idx := -1
		switch x := node.(type) {
		case *NilExp:
			idx = fi.constantIndex(nil)
		case *FalseExp:
			idx = fi.constantIndex(false)
		case *TrueExp:
			idx = fi.constantIndex(true)
		case *IntegerExp:
			idx = fi.constantIndex(x.Val)
		case *FloatExp:
			idx = fi.constantIndex(x.Val)
		case *StringExp:
			idx = fi.constantIndex(x.Str)
		}
		if idx >= 0 && idx <= 0xFF {
			return 0x100 + idx, ARG_CONST
		}
	}

	if nameExp, ok := node.(*NameExp); ok {
		if nameExp.Name == "_" {
			return 0, ARG_BLANK
		}
		if argKinds&ARG_REG > 0 {
			if r := fi.slotOfLocVar(nameExp.Name); r >= 0 {
				return r, ARG_REG
			}
		}
		if argKinds&ARG_UPVAL > 0 {
			if idx := fi.upvalIndex(nameExp.Name); idx >= 0 {
				return idx, ARG_UPVAL
			}
		}
	}

	a := fi.preAllocReg()
	fi.generateExpression(node, a, 1)
	fi.checkAllocReg(a)
	return a, ARG_REG
}
