package generator

import (
	. "lxa/compiler/ast"
)

func (fi *funcInfo) generateStatement(node Statement) {
	switch stat := node.(type) {
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
		fi.generateAssignment(stat.Asn)
	}
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

	for _, asn := range node.AsnList {
		fi.generateAssignment(asn)
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
	if node.StepAsn != nil {
		fi.generateAssignment(node.StepAsn)
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

		for _, asn := range sub.AsnList {
			fi.generateAssignment(asn)
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

	fi.generateAssignment(&LocVarDeclAsn{
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
