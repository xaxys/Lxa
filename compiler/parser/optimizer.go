package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
	"lxa/number"
	"math"
)

func OptimizeLogicalQst(exp *UnopExp) Expression {
	switch x := exp.Exp.(type) {
	case *IntegerExp:
		if x.Val == 0 {
			return &FalseExp{Line: x.Line}
		} else {
			return &TrueExp{Line: x.Line}
		}
	case *FloatExp:
		if x.Val == 0 {
			return &FalseExp{Line: x.Line}
		} else {
			return &TrueExp{Line: x.Line}
		}
	case *StringExp:
		if len(x.Str) == 0 {
			return &FalseExp{Line: x.Line}
		} else {
			return &TrueExp{Line: x.Line}
		}
	}
	return exp
}

func OptimizeLogicalOr(exp *BinopExp) Expression {
	if exp.Exp1.IsTrue() {
		return exp.Exp1 // true or x => true
	} else if exp.Exp1.IsFalse() && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // false or x => x
	}
	return exp
}

func OptimizeLogicalAnd(exp *BinopExp) Expression {
	if exp.Exp1.IsFalse() {
		return exp.Exp1 // false and x => false
	} else if exp.Exp1.IsTrue() && !isVarargOrFuncCall(exp.Exp2) {
		return exp.Exp2 // true and x => x
	}
	return exp
}

func OptimizeBitwiseBinaryOp(exp *BinopExp) Expression {
	if exp1, ok := exp.Exp1.(NumeralExpression); ok {
		i := exp1.CastToInt()
		if exp2, ok := exp.Exp2.(NumeralExpression); ok {
			j := exp2.CastToInt()
			switch exp.Op.Type {
			case TOKEN_OP_BAND:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  i & j,
				}
			case TOKEN_OP_BOR:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  i | j,
				}
			case TOKEN_OP_BXOR:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  i ^ j,
				}
			case TOKEN_OP_SHL:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  number.ShiftLeft(i, j),
				}
			case TOKEN_OP_SHR:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  number.ShiftRight(i, j),
				}
			}
		}
	}
	return exp
}

func OptimizeArithBinaryOp(exp *BinopExp) Expression {
	if x, ok := exp.Exp1.(*IntegerExp); ok {
		if y, ok := exp.Exp2.(*IntegerExp); ok {
			switch exp.Op.Type {
			case TOKEN_OP_ADD:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  x.Val + y.Val,
				}
			case TOKEN_OP_SUB:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  x.Val - y.Val,
				}
			case TOKEN_OP_MUL:
				return &IntegerExp{
					Line: exp.Op.Line,
					Val:  x.Val * y.Val,
				}
			case TOKEN_OP_IDIV:
				if y.Val != 0 {
					return &IntegerExp{
						Line: exp.Op.Line,
						Val:  number.IFloorDiv(x.Val, y.Val),
					}
				}
			case TOKEN_OP_MOD:
				if y.Val != 0 {
					return &IntegerExp{
						Line: exp.Op.Line,
						Val:  number.IMod(x.Val, y.Val),
					}
				}
			}
		}
	}
	if exp1, ok := exp.Exp1.(NumeralExpression); ok {
		f := exp1.CastToFloat()
		if exp2, ok := exp.Exp2.(NumeralExpression); ok {
			g := exp2.CastToFloat()
			switch exp.Op.Type {
			case TOKEN_OP_ADD:
				return &FloatExp{
					Line: exp.Op.Line,
					Val:  f + g,
				}
			case TOKEN_OP_SUB:
				return &FloatExp{
					Line: exp.Op.Line,
					Val:  f - g,
				}
			case TOKEN_OP_MUL:
				return &FloatExp{
					Line: exp.Op.Line,
					Val:  f * g,
				}
			case TOKEN_OP_DIV:
				if g != 0 {
					return &FloatExp{
						Line: exp.Op.Line,
						Val:  f / g,
					}
				}
			case TOKEN_OP_IDIV:
				if g != 0 {
					return &FloatExp{
						Line: exp.Op.Line,
						Val:  number.FFloorDiv(f, g),
					}
				}
			case TOKEN_OP_MOD:
				if g != 0 {
					return &FloatExp{
						Line: exp.Op.Line,
						Val:  number.FMod(f, g),
					}
				}
			case TOKEN_OP_POW:
				return &FloatExp{
					Line: exp.Op.Line,
					Val:  math.Pow(f, g),
				}
			}
		}
	}
	return exp
}

func OptimizePow(exp Expression) Expression {
	if binop, ok := exp.(*BinopExp); ok {
		if binop.Op.Is(TOKEN_OP_POW) {
			binop.Exp2 = OptimizePow(binop.Exp2)
		}
		return OptimizeArithBinaryOp(binop)
	}
	return exp
}

func OptimizeUnaryOp(exp *UnopExp) Expression {
	switch exp.Op.Type {
	case TOKEN_OP_UNM:
		return OptimizeUnm(exp)
	case TOKEN_OP_NOT:
		return OptimizeNot(exp)
	case TOKEN_OP_BNOT:
		return OptimizeBnot(exp)
	}
	return exp
}

func OptimizeUnm(exp *UnopExp) Expression {
	switch x := exp.Exp.(type) { // number?
	case *IntegerExp:
		x.Val = -x.Val
		return x
	case *FloatExp:
		if x.Val != 0 {
			x.Val = -x.Val
			return x
		}
	}
	return exp
}

func OptimizeNot(exp *UnopExp) Expression {
	if exp.Exp.IsTrue() {
		return &FalseExp{
			Line: exp.Op.Line,
		}
	}
	if exp.Exp.IsFalse() {
		return &TrueExp{
			Line: exp.Op.Line,
		}
	}
	return exp
}

func OptimizeBnot(exp *UnopExp) Expression {
	switch x := exp.Exp.(type) { // number?
	case *IntegerExp:
		x.Val = ^x.Val
		return x
	case *FloatExp:
		if i, ok := number.FloatToInteger(x.Val); ok {
			return &IntegerExp{
				Line: x.Line,
				Val:  ^i,
			}
		}
	}
	return exp
}

func isVarargOrFuncCall(exp Expression) bool {
	switch exp.(type) {
	case *VarargExp, *FuncCallExp:
		return true
	default:
		return false
	}
}
