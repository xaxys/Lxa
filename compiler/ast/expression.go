package ast

import (
	"fmt"
	. "lxa/compiler/token"
)

/*
exp ::=  nil
	| false
	| true
	| Numeral
	| LiteralString
	| '...'
	| functiondef
	| functioncall
	| prefixexp
	| tableconstructor
	| exp binop exp
	| unop exp
	| varlist ['=' explist]
	| namelist ':=' explist

prefixexp ::= var | functioncall | '(' exp ')'
var ::=  Name | prefixexp '[' exp ']' | prefixexp '.' Name
varlist ::= var {',' var}
explist ::= exp {',' exp}
functioncall ::=  prefixexp args | prefixexp ':' Name args
*/
type Expression interface {
	IsTrue() bool
	IsFalse() bool
}

type NumeralExpression interface {
	CastToInt() int64
	CastToFloat() float64
}

type TrueExpression struct{}

func (exp TrueExpression) IsTrue() bool  { return true }
func (exp TrueExpression) IsFalse() bool { return false }

type FalseExpression struct{}

func (exp FalseExpression) IsTrue() bool  { return false }
func (exp FalseExpression) IsFalse() bool { return true }

type NoBoolExpression struct{}

func (exp NoBoolExpression) IsTrue() bool  { return false }
func (exp NoBoolExpression) IsFalse() bool { return false }

type BlankExp struct { // _
	FalseExpression
	Line int
}

type NilExp struct { // nil
	FalseExpression
	Line int
}

func (exp *NilExp) String() string {
	return fmt.Sprintf("Line: %d, nil", exp.Line)
}

type TrueExp struct { // true
	TrueExpression
	Line int
}

func (exp *TrueExp) String() string {
	return fmt.Sprintf("Line: %d, true", exp.Line)
}

type FalseExp struct { // false
	FalseExpression
	Line int
}

func (exp *FalseExp) String() string {
	return fmt.Sprintf("Line: %d, false", exp.Line)
}

type VarargExp struct { // ...
	NoBoolExpression
	Line int
}

func (exp *VarargExp) String() string {
	return fmt.Sprintf("Line: %d, vararg(...)", exp.Line)
}

// Numeral
type IntegerExp struct {
	TrueExpression
	Line int
	Val  int64
}

func (exp *IntegerExp) String() string {
	return fmt.Sprintf("Line: %d, Integer, Val: %d", exp.Line, exp.Val)
}

func (exp *IntegerExp) CastToInt() int64     { return exp.Val }
func (exp *IntegerExp) CastToFloat() float64 { return float64(exp.Val) }

type FloatExp struct {
	TrueExpression
	Line int
	Val  float64
}

func (exp *FloatExp) CastToInt() int64     { return int64(exp.Val) }
func (exp *FloatExp) CastToFloat() float64 { return exp.Val }

func (exp *FloatExp) String() string {
	return fmt.Sprintf("Line: %d, Float, Val: %f", exp.Line, exp.Val)
}

// LiteralString
type StringExp struct {
	Line int
	Str  string
}

func (exp *StringExp) IsTrue() bool  { return len(exp.Str) != 0 }
func (exp *StringExp) IsFalse() bool { return len(exp.Str) == 0 }

func (exp *StringExp) String() string {
	return fmt.Sprintf("Line: %d, Float, Str: %s", exp.Line, exp.Str)
}

// unop exp
type UnopExp struct {
	NoBoolExpression
	Op  *Token // operator
	Exp Expression
}

func (exp *UnopExp) String() string {
	return fmt.Sprintf("Line: %d, UnopExp, Op: %s, Exp: %s",
		exp.Op.Line, exp.Op, exp.Exp)
}

// (and | or) between expList
type LogicalExp struct {
	NoBoolExpression
	Op      *Token
	ExpList []Expression
}

func (exp *LogicalExp) String() string {
	return fmt.Sprintf("LogicalExp, ExpList: %s", exp.ExpList)
}

// exp1 op exp2
type BinopExp struct {
	NoBoolExpression
	Op   *Token // operator
	Exp1 Expression
	Exp2 Expression
}

func (exp *BinopExp) String() string {
	return fmt.Sprintf("Line: %d, BinopExp, Exp1: %s, Op: %s, Exp2: %s",
		exp.Op.Line, exp.Exp1, exp.Op, exp.Exp2)
}

type ConcatExp struct {
	NoBoolExpression
	Line    int // line of last ..
	ExpList []Expression
}

func (exp *ConcatExp) String() string {
	return fmt.Sprintf("Line: %d, ConcatExp, ExpList: %s", exp.Line, exp.ExpList)
}

// tableconstructor ::= '{' [fieldlist] '}'
// fieldlist ::= field {fieldsep field} [fieldsep]
// field ::= '[' exp ']' '=' exp | Name '=' exp | exp
// fieldsep ::= ',' | ';'
type TableConstructorExp struct {
	NoBoolExpression
	Line     int // line of `{` ?
	LastLine int // line of `}`
	KeyExps  []Expression
	ValExps  []Expression
}

func (exp *TableConstructorExp) String() string {
	s := fmt.Sprintf("Line: %d, LastLine: %d, TableConstructor, Entry:",
		exp.Line, exp.LastLine)
	for i, k := range exp.KeyExps {
		s += fmt.Sprintf("[%s: %s]", k, exp.ValExps[i])
	}
	return s
}

// functiondef ::= func funcbody
// funcbody ::= '(' [parlist] ')' '{' block '}'
// parlist ::= namelist [',' '...'] | '...'
// namelist ::= Name {',' Name}
type FuncDefExp struct {
	NoBoolExpression
	Line     int
	LastLine int // line of `}`
	ParList  []string
	IsVararg bool
	Block    *Block
}

func (exp *FuncDefExp) String() string {
	return fmt.Sprintf("Line: %d, LastLine: %d, FuncDef, IsVararg: %v, ParList: %s",
		exp.Line, exp.LastLine, exp.IsVararg, exp.ParList)
}

// prefixexp ::= Name
//		| '(' exp ')'
//		| prefixexp '[' exp ']'
//		| prefixexp '.' Name
//		| prefixexp ':' Name args
//		| prefixexp args

type NameExp struct {
	NoBoolExpression
	Line int
	Name string
}

func (exp *NameExp) String() string {
	return fmt.Sprintf("Line: %d, Name, Name: %s", exp.Line, exp.Name)
}

type ParensExp struct {
	NoBoolExpression
	Exp Expression
}

func (exp *ParensExp) String() string {
	return fmt.Sprintf("ParensExp, Exp: %s", exp.Exp)
}

type TableAccessExp struct {
	NoBoolExpression
	LastLine  int // line of ']' ?
	PrefixExp Expression
	KeyExp    Expression
}

func (exp *TableAccessExp) String() string {
	return fmt.Sprintf("LastLine: %d, TableAccessExp, PrefixExp: %s, KeyExp: %s",
		exp.LastLine, exp.PrefixExp, exp.KeyExp)
}

type FuncCallExp struct {
	NoBoolExpression
	Line      int // line of '(' ?
	LastLine  int // line of ')'
	PrefixExp Expression
	NameExp   *StringExp
	Args      []Expression
}

func (exp *FuncCallExp) String() string {
	return fmt.Sprintf("Line: %d, LastLine: %d, FuncCall, PrefixExp: %s, NameExp: %s, Args: %s",
		exp.Line, exp.LastLine, exp.PrefixExp, exp.NameExp, exp.Args)
}
