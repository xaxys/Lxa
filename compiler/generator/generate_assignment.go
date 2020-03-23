package generator

import (
	. "lxa/compiler/ast"
)

func (fi *funcInfo) generateAssignment(node Assignment) {
	switch asn := node.(type) {
	case *AssignAsn:
		fi.generateAssignAsn(asn)
	case *LocVarDeclAsn:
		fi.generateLocVarDeclAsn(asn)
	case *FuncCallAsn:
		fi.generateFuncCallAsn(asn)
	}
}

func (fi *funcInfo) generateFuncCallAsn(node *FuncCallAsn) {
	r := fi.preAllocReg()
	fi.generateFuncCallExp(node, r, 0)
	fi.checkAllocReg(r)
	fi.freeReg()
}

func (fi *funcInfo) generateLocVarDeclAsn(node *LocVarDeclAsn) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nNames := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nNames {
		for _, exp := range exps {
			a := fi.preAllocReg()
			fi.generateExpression(exp, a, 1)
			fi.checkAllocReg(a)
		}
	} else if nExps > nNames {
		for i, exp := range exps {
			a := fi.preAllocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				fi.generateExpression(exp, a, 0)
			} else {
				fi.generateExpression(exp, a, 1)
			}
			fi.checkAllocReg(a)
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.preAllocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				fi.generateExpression(exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				fi.generateExpression(exp, a, 1)
			}
			fi.checkAllocReg(a)
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	fi.usedRegs = oldRegs
	startPC := fi.pc() + 1
	for _, name := range node.NameList {
		fi.addLocVar(name, startPC)
	}
}

func (fi *funcInfo) generateAssignAsn(node *AssignAsn) {
	exps := removeTailNils(node.ExpList)
	nExps := len(exps)
	nVars := len(node.VarList)

	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)
	oldRegs := fi.usedRegs

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*TableAccessExp); ok {
			tRegs[i] = fi.preAllocReg()
			fi.generateExpression(taExp.PrefixExp, tRegs[i], 1)
			fi.checkAllocReg(tRegs[i])
			kRegs[i] = fi.preAllocReg()
			fi.generateExpression(taExp.KeyExp, kRegs[i], 1)
			fi.checkAllocReg(kRegs[i])
		} else {
			name := exp.(*NameExp).Name
			if fi.slotOfLocVar(name) < 0 && fi.upvalIndex(name) < 0 {
				// global var
				kRegs[i] = -1
				if fi.constantIndex(name) > 0xFF {
					kRegs[i] = fi.allocReg()
				}
			}
		}
	}
	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range exps {
			a := fi.preAllocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				fi.generateExpression(exp, a, 0)
			} else {
				fi.generateExpression(exp, a, 1)
			}
			fi.checkAllocReg(a)
		}
	} else { // nVars > nExps
		multRet := false
		for i, exp := range exps {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				fi.generateExpression(exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				fi.generateExpression(exp, a, 1)
			}
			fi.checkAllocReg(a)
		}
		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(node.LastLine, a, n)
		}
	}

	lastLine := node.LastLine
	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(lastLine, a, vRegs[i])
			} else if b := fi.upvalIndex(varName); b >= 0 {
				fi.emitSetUpval(lastLine, vRegs[i], b)
			} else if a := fi.slotOfLocVar("_ENV"); a >= 0 {
				if kRegs[i] < 0 {
					b := 0x100 + fi.constantIndex(varName)
					fi.emitSetTable(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTable(lastLine, a, kRegs[i], vRegs[i])
				}
			} else { // global var
				a := fi.upvalIndex("_ENV")
				if kRegs[i] < 0 {
					b := 0x100 + fi.constantIndex(varName)
					fi.emitSetTabUp(lastLine, a, b, vRegs[i])
				} else {
					fi.emitSetTabUp(lastLine, a, kRegs[i], vRegs[i])
				}
			}
		} else {
			fi.emitSetTable(lastLine, tRegs[i], kRegs[i], vRegs[i])
		}
	}

	// todo
	fi.usedRegs = oldRegs
}
