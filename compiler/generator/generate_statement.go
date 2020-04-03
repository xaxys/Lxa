package generator

import (
	. "lxa/compiler/ast"
)

func (fi *funcInfo) generateStatement(node Statement) {
	switch stat := node.(type) {
	case *BlockStat:
		fi.generateBlockStat(stat)
	case *BreakStat:
		fi.generateBreakStat(stat)
	case *ContinueStat:
		fi.generateContinueStat(stat)
	case *LoopStat:
		fi.generateLoopStat(stat)
	case *IfStat:
		fi.generateIfStat(stat)
	case *ForInStat:
		fi.generateForInStat(stat)
	case *AssignmentStat:
		fi.generateAssignmentStat(stat)
	case *FuncCallStat:
		fi.generateFuncCallStat(stat)
	case *LocVarDeclStat:
		fi.generateLocVarDeclStat(stat)
	}
}

func (fi *funcInfo) generateBlockStat(node *BlockStat) {
	fi.enterScope(false)
	fi.generateBlock(node.Block)
	fi.exitScope(fi.pc() + 1)
}

func (fi *funcInfo) generateBreakStat(node *BreakStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addBreakJmp(pc)
}

func (fi *funcInfo) generateContinueStat(node *ContinueStat) {
	pc := fi.emitJmp(node.Line, 0, 0)
	fi.addContinueJmp(pc)
}

/*
         _____________
        /  false? jmp |
       /              |
while exp { block } <-'
      ^           \
      |___________/
           jmp
*/
func (fi *funcInfo) generateLoopStat(node *LoopStat) {
	fi.enterScope(true)

	for _, stat := range node.InitList {
		fi.generateStatement(stat)
	}

	pcBeforeExp := fi.pc()

	oldRegs := fi.usedRegs
	a, _ := fi.expToOpArg(node.Exp, ARG_REG)
	fi.usedRegs = oldRegs

	line := lastLineOf(node.Exp)
	fi.emitTest(line, a, 0)
	pcJmpToEnd := fi.emitJmp(line, 0, 0)

	fi.generateBlock(node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.setContinueJmp()
	if node.StepStat != nil {
		fi.generateStatement(node.StepStat)
	}
	fi.emitJmp(node.Block.LastLine, 0, pcBeforeExp-fi.pc()-1)
	fi.exitScope(fi.pc())

	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func (fi *funcInfo) generateIfStat(node *IfStat) {
	pcJmpToEnds := make([]int, len(node.SubList))
	pcJmpToNextExp := -1

	for i, sub := range node.SubList {
		if pcJmpToNextExp >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
		}

		fi.enterScope(false)

		for _, stat := range sub.InitList {
			fi.generateStatement(stat)
		}

		oldRegs := fi.usedRegs
		a, _ := fi.expToOpArg(sub.Exp, ARG_REG)
		fi.usedRegs = oldRegs

		line := lastLineOf(sub.Exp)
		fi.emitTest(line, a, 0)
		pcJmpToNextExp = fi.emitJmp(line, 0, 0)

		block := sub.Block
		fi.generateBlock(block)
		fi.closeOpenUpvals(block.LastLine)
		fi.exitScope(fi.pc() + 1)
		if i < len(node.SubList)-1 {
			pcJmpToEnds[i] = fi.emitJmp(block.LastLine, 0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

func (fi *funcInfo) generateForInStat(node *ForInStat) {
	forGeneratorVar := "(for generator)"
	forStateVar := "(for state)"
	forControlVar := "(for control)"

	fi.enterScope(true)

	fi.generateLocVarDeclStat(&LocVarDeclStat{
		//LastLine: 0,
		NameList: []string{forGeneratorVar, forStateVar, forControlVar},
		ExpList:  node.ExpList,
	})
	for _, name := range node.NameList {
		fi.addLocVar(name, fi.pc()+2)
	}

	pcJmpToTFC := fi.emitJmp(node.LineBlock, 0, 0)
	fi.generateBlock(node.Block)
	fi.closeOpenUpvals(node.Block.LastLine)
	fi.setContinueJmp()
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)

	line := lineOf(node.ExpList[0])
	rGenerator := fi.slotOfLocVar(forGeneratorVar)
	fi.emitTForCall(line, rGenerator, len(node.NameList))
	fi.emitTForLoop(line, rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope(fi.pc() - 1)
	fi.fixEndPC(forGeneratorVar, 2)
	fi.fixEndPC(forStateVar, 2)
	fi.fixEndPC(forControlVar, 2)
}

func (fi *funcInfo) generateFuncCallStat(node *FuncCallStat) {
	r := fi.preAllocReg()
	fi.generateFuncCallExp(node, r, 0)
	fi.checkAllocReg(r)
	fi.freeReg()
}

func (fi *funcInfo) generateLocVarDeclStat(node *LocVarDeclStat) {
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

func (fi *funcInfo) generateAssignmentStat(node *AssignmentStat) {
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
