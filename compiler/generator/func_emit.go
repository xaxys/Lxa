package generator

import (
	. "lxa/compiler/token"
	. "lxa/vm"
)

// code

func (fi *funcInfo) pc() int {
	return len(fi.insts) - 1
}

func (fi *funcInfo) fixSbx(pc, sBx int) {
	i := fi.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+MAXARG_sBx)<<14 // reset sBx
	fi.insts[pc] = i
}

// todo: rename?
func (fi *funcInfo) fixEndPC(name string, delta int) {
	for i := len(fi.locVars) - 1; i >= 0; i-- {
		locVar := fi.locVars[i]
		if locVar.name == name {
			locVar.endPC += delta
			return
		}
	}
}

func (fi *funcInfo) emitABC(line, opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitABx(line, opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitAsBx(line, opcode, a, b int) {
	i := (b+MAXARG_sBx)<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

func (fi *funcInfo) emitAx(line, opcode, ax int) {
	i := ax<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
	fi.lineNums = append(fi.lineNums, uint32(line))
}

// r[a] = r[b]
func (fi *funcInfo) emitMove(line, a, b int) {
	fi.emitABC(line, OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (fi *funcInfo) emitLoadNil(line, a, n int) {
	fi.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (fi *funcInfo) emitLoadBool(line, a, b, c int) {
	fi.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (fi *funcInfo) emitLoadK(line, a int, k interface{}) {
	idx := fi.constantIndex(k)
	if idx < (1 << 18) {
		fi.emitABx(line, OP_LOADK, a, idx)
	} else {
		fi.emitABx(line, OP_LOADKX, a, 0)
		fi.emitAx(line, OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (fi *funcInfo) emitVararg(line, a, n int) {
	fi.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (fi *funcInfo) emitClosure(line, a, bx int) {
	fi.emitABx(line, OP_CLOSURE, a, bx)
}

// r[a] = {}
func (fi *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	fi.emitABC(line, OP_NEWTABLE,
		a, Int2fb(nArr), Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (fi *funcInfo) emitSetList(line, a, b, c int) {
	fi.emitABC(line, OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (fi *funcInfo) emitGetTable(line, a, b, c int) {
	fi.emitABC(line, OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (fi *funcInfo) emitSetTable(line, a, b, c int) {
	fi.emitABC(line, OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (fi *funcInfo) emitGetUpval(line, a, b int) {
	fi.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (fi *funcInfo) emitSetUpval(line, a, b int) {
	fi.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (fi *funcInfo) emitGetTabUp(line, a, b, c int) {
	fi.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (fi *funcInfo) emitSetTabUp(line, a, b, c int) {
	fi.emitABC(line, OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (fi *funcInfo) emitCall(line, a, nArgs, nRet int) {
	fi.emitABC(line, OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (fi *funcInfo) emitTailCall(line, a, nArgs int) {
	fi.emitABC(line, OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (fi *funcInfo) emitReturn(line, a, n int) {
	fi.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (fi *funcInfo) emitSelf(line, a, b, c int) {
	fi.emitABC(line, OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (fi *funcInfo) emitJmp(line, a, sBx int) int {
	fi.emitAsBx(line, OP_JMP, a, sBx)
	return len(fi.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (fi *funcInfo) emitTest(line, a, c int) {
	fi.emitABC(line, OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (fi *funcInfo) emitTestSet(line, a, b, c int) {
	fi.emitABC(line, OP_TESTSET, a, b, c)
}

func (fi *funcInfo) emitForPrep(line, a, sBx int) int {
	fi.emitAsBx(line, OP_FORPREP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitForLoop(line, a, sBx int) int {
	fi.emitAsBx(line, OP_FORLOOP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitTForCall(line, a, c int) {
	fi.emitABC(line, OP_TFORCALL, a, 0, c)
}

func (fi *funcInfo) emitTForLoop(line, a, sBx int) {
	fi.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (fi *funcInfo) emitUnaryOp(line int, op TokenType, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		fi.emitABC(line, OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		fi.emitABC(line, OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		fi.emitABC(line, OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		fi.emitABC(line, OP_UNM, a, b, 0)
	case TOKEN_OP_QST:
		fi.emitABC(line, OP_EQ, 0, b, 0)
		fi.emitJmp(line, 0, 1)
		fi.emitLoadBool(line, a, 0, 1)
		fi.emitLoadBool(line, a, 1, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (fi *funcInfo) emitBinaryOp(line int, op TokenType, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		fi.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case TOKEN_OP_EQ:
			fi.emitABC(line, OP_EQ, 1, b, c)
		case TOKEN_OP_NE:
			fi.emitABC(line, OP_EQ, 0, b, c)
		case TOKEN_OP_LT:
			fi.emitABC(line, OP_LT, 1, b, c)
		case TOKEN_OP_GT:
			fi.emitABC(line, OP_LT, 1, c, b)
		case TOKEN_OP_LE:
			fi.emitABC(line, OP_LE, 1, b, c)
		case TOKEN_OP_GE:
			fi.emitABC(line, OP_LE, 1, c, b)
		}
		fi.emitJmp(line, 0, 1)
		fi.emitLoadBool(line, a, 0, 1)
		fi.emitLoadBool(line, a, 1, 0)
	}
}
