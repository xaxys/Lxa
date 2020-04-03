package generator

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
	. "lxa/vm"
)

const REG_SIZE = 255

var arithAndBitwiseBinops = map[TokenType]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

type funcInfo struct {
	parent   *funcInfo
	subFuncs []*funcInfo

	usedRegs int
	maxRegs  int

	scopeLv  int
	locVars  []*locVarInfo
	locNames map[string]*locVarInfo

	upvals    map[string]upvalInfo
	constants map[interface{}]int
	breaks    [][]int
	continues [][]int
	insts     []uint32
	lineNums  []uint32
	line      int
	lastLine  int
	numParams int
	isVararg  bool

	block *Block
}

type locVarInfo struct {
	prev     *locVarInfo
	name     string
	scopeLv  int
	slot     int
	startPC  int
	endPC    int
	captured bool
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvals:    map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		continues: make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		lineNums:  make([]uint32, 0, 8),
		line:      fd.Line,
		lastLine:  fd.LastLine,
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
		block:     fd.Block,
	}
}

// constants

func (fi *funcInfo) constantIndex(k interface{}) int {
	if idx, ok := fi.constants[k]; ok {
		return idx
	}

	idx := len(fi.constants)
	fi.constants[k] = idx
	return idx
}

// registers

func (fi *funcInfo) allocReg() int {
	fi.usedRegs++
	if fi.usedRegs >= REG_SIZE {
		panic("function or expression needs too many registers")
	}
	if fi.usedRegs > fi.maxRegs {
		fi.maxRegs = fi.usedRegs
	}
	return fi.usedRegs - 1
}

func (fi *funcInfo) preAllocReg() int {
	if fi.usedRegs+1 >= REG_SIZE {
		panic("function or expression needs too many registers")
	}
	if fi.usedRegs+1 > fi.maxRegs {
		fi.maxRegs = fi.usedRegs + 1
	}
	return fi.usedRegs
}

func (fi *funcInfo) checkAllocReg(a int) {
	if fi.usedRegs == a {
		fi.allocReg()
	} else if fi.usedRegs < a {
		panic("registers num error")
	}
}

func (fi *funcInfo) freeReg() {
	if fi.usedRegs <= 0 {
		panic("usedRegs <= 0 !")
	}
	fi.usedRegs--
}

func (fi *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("allocRegs n <= 0 !")
	}
	for i := 0; i < n; i++ {
		fi.allocReg()
	}
	return fi.usedRegs - n
}

func (fi *funcInfo) freeRegs(n int) {
	if n < 0 {
		panic("freeRegs n < 0 !")
	}
	for i := 0; i < n; i++ {
		fi.freeReg()
	}
}

// lexical scope

func (fi *funcInfo) enterScope(loopScope bool) {
	fi.scopeLv++
	if loopScope {
		fi.breaks = append(fi.breaks, []int{})
		fi.continues = append(fi.continues, []int{})
	} else {
		fi.breaks = append(fi.breaks, nil)
		fi.continues = append(fi.continues, nil)
	}
}

func (fi *funcInfo) exitScope(endPC int) { // Unchanged continue method
	pendingBreakJmps := fi.breaks[len(fi.breaks)-1]
	fi.breaks = fi.breaks[:len(fi.breaks)-1]

	a := fi.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := fi.pc() - pc
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP
		fi.insts[pc] = uint32(i)
	}

	fi.scopeLv--
	for _, locVar := range fi.locNames {
		if locVar.scopeLv > fi.scopeLv { // out of scope
			locVar.endPC = endPC
			fi.removeLocVar(locVar)
		}
	}
}

func (fi *funcInfo) removeLocVar(locVar *locVarInfo) {
	fi.freeReg()
	if locVar.prev == nil {
		delete(fi.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		fi.removeLocVar(locVar.prev)
	} else {
		fi.locNames[locVar.name] = locVar.prev
	}
}

func (fi *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    fi.locNames[name],
		scopeLv: fi.scopeLv,
		slot:    fi.allocReg(),
		startPC: startPC,
		endPC:   0,
	}

	fi.locVars = append(fi.locVars, newVar)
	fi.locNames[name] = newVar

	return newVar.slot
}

func (fi *funcInfo) slotOfLocVar(name string) int {
	if locVar, ok := fi.locNames[name]; ok {
		return locVar.slot
	}
	return -1
}

func (fi *funcInfo) addBreakJmp(pc int) {
	for i := fi.scopeLv; i >= 0; i-- {
		if fi.breaks[i] != nil { // breakable
			fi.breaks[i] = append(fi.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop!")
}

func (fi *funcInfo) addContinueJmp(pc int) {
	for i := fi.scopeLv; i >= 0; i-- {
		if fi.continues[i] != nil { // continueable
			fi.continues[i] = append(fi.continues[i], pc)
			return
		}
	}

	panic("<continue> at line ? not inside a loop!")
}

func (fi *funcInfo) setContinueJmp() {
	continueJmps := fi.continues[len(fi.continues)-1]
	fi.continues = fi.continues[:len(fi.continues)-1]

	for _, pc := range continueJmps {
		sBx := fi.pc() - pc
		fi.fixSbx(pc, sBx)
	}
}

// upvalues

func (fi *funcInfo) upvalIndex(name string) int {
	if upval, ok := fi.upvals[name]; ok {
		return upval.index
	}
	if fi.parent != nil {
		if locVar, ok := fi.parent.locNames[name]; ok {
			idx := len(fi.upvals)
			fi.upvals[name] = upvalInfo{locVar.slot, -1, idx}
			locVar.captured = true
			return idx
		}
		if upvalIdx := fi.parent.upvalIndex(name); upvalIdx >= 0 {
			idx := len(fi.upvals)
			fi.upvals[name] = upvalInfo{-1, upvalIdx, idx}
			return idx
		}
	}
	return -1
}

func (fi *funcInfo) closeOpenUpvals(line int) {
	a := fi.getJmpArgA()
	if a > 0 {
		fi.emitJmp(line, a, 0)
	}
}

func (fi *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := fi.maxRegs
	for _, locVar := range fi.locNames {
		if locVar.scopeLv == fi.scopeLv {
			for v := locVar; v != nil && v.scopeLv == fi.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}
