package ast

// stat ::= ';'
// 		| assignment ';'
// 		| break
// 		| continue
// 		| while {assignment ';'} exp '{' block '}'
// 		| if {assignment ';'} exp '{' block '}' {else if {assignment ';'} exp '{' block '}'} [else '{' block '}']
// 		| for assignment ';' exp ';' assignment '{' block '}'
// 		| for namelist in explist '{' block '}'
// 		| func funcname funcbody
// 		| local func Name funcbody

type Statement interface{}

type Statements struct {
	StatList []Statement
}

type EmptyStat struct{}              // ';'
type BreakStat struct{ Line int }    // break
type ContinueStat struct{ Line int } // continue
type AssignmentStat struct {         // assignment
	Asn Assignment
}

type LoopStat struct {
	AsnList []Assignment
	Exp     Expression
	StepAsn Assignment
	Block   *Block
}

// while {assignment ';'} exp '{' block '}' => LoopStat
/*
type WhileStat struct {
	AsnList []Assignment
	Exp     Expression
	Block   *Block
}
*/

// for assignment ';' exp ';' assignment '{' block '}' => LoopStat
/*
type ForNumStat struct {
	LineFor   int
	LineBlock int
	InitAsn   Assignment
	LimitExp  Expression
	StepAsn   Assignment
	Block     *Block
}
*/

// for namelist in explist '{' block '}'
// namelist ::= Name {',' Name}
// explist ::= exp {',' exp}
type ForInStat struct {
	LineBlock int
	NameList  []string
	ExpList   []Expression
	Block     *Block
}

// if {assignment ';'} exp '{' block '}' {else if {assignment ';'} exp '{' block '}'}
type IfStat struct {
	SubList []*SubIfStat
}

// {stat ';'} exp '{' block '}'
type SubIfStat struct {
	AsnList []Assignment
	Exp     Expression
	Block   *Block
}
