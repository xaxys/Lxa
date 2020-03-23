package ast

type Assignment interface{}

type FuncCallAsn = FuncCallExp // functioncall

// varlist ('+=' | '-=' | '*=' | '/=' | '~/=' | '%='
//		| '&=' | '^=' | '|=' | '**=' | '<<=' | '>>=' | '=') explist
// varlist ::= var {',' var}
// var ::=  Name | prefixexp '[' exp ']' | prefixexp '.' Name
type AssignAsn struct {
	LastLine int
	VarList  []Expression
	ExpList  []Expression
}

// namelist ':=' explist | local namelist [':=' explist]
// namelist ::= Name {',' Name}
// explist ::= exp {',' exp}
type LocVarDeclAsn struct {
	LastLine int
	NameList []string
	ExpList  []Expression
}
