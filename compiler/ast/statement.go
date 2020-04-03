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

// while {assignment ';'} exp '{' block '}' => LoopStat
// for assignment ';' exp ';' assignment '{' block '}' => LoopStat
type LoopStat struct {
	InitList []Statement
	Exp      Expression
	StepStat Statement
	Block    *Block
}

type BlockStat struct {
	Block *Block
}

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
	InitList []Statement
	Exp      Expression
	Block    *Block
}

type FuncCallStat = FuncCallExp // functioncall

// varlist ('+=' | '-=' | '*=' | '/=' | '~/=' | '%='
//		| '&=' | '^=' | '|=' | '**=' | '<<=' | '>>=' | '=') explist
// varlist ::= var {',' var}
// var ::=  Name | prefixexp '[' exp ']' | prefixexp '.' Name
type AssignmentStat struct {
	LastLine int
	VarList  []Expression
	ExpList  []Expression
}

// namelist ':=' explist | local namelist [':=' explist]
// namelist ::= Name {',' Name}
// explist ::= exp {',' exp}
type LocVarDeclStat struct {
	LastLine int
	NameList []string
	ExpList  []Expression
}
