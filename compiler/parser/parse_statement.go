package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
)

/*
stat ::= ';'
	| assignment ';'
	| break
	| continue
	| while {assignment ';'} exp '{' block '}'
	| if {assignment ';'} exp '{' block '}' {else if {assignment ';'} exp '{' block '}'} [else '{' block '}']
	| for assignment ';' exp ';' assignment '{' block '}'
	| for namelist in explist '{' block '}'
	| func funcname funcbody
	| local func Name funcbody
*/

func (p *Parser) parseStatement() Statement {
	switch p.lexer.PeekToken().Type {
	case TOKEN_SEP_SEMI, TOKEN_SEP_EOLN: // EmptyStat
		return p.parseEmptyStat()
	case TOKEN_SEP_LCURLY: // { block }
		return p.parseBlockStat()
	case TOKEN_KW_BREAK: // BreakStat
		return p.parseBreakStat()
	case TOKEN_KW_CONTINUE: // ContinueStat
		return p.parseContinueStat()
	case TOKEN_KW_WHILE: // WhileStat
		return p.parseWhileStat()
	case TOKEN_KW_IF: // IfStat
		return p.parseIfStat()
	case TOKEN_KW_FOR: // ForStat
		return p.parseForStat()
	case TOKEN_KW_FUNC: // FuncDefStat
		return p.parseFuncDefStat()
	case TOKEN_KW_LOCAL: // LocalAssign or LocalFuncDefStat
		return p.parseLocAssignOrLocFuncDefStat()
	default: // Assignment or FuncCall
		return p.parseAssignOrFuncCallStat()
	}
}

var emptyStat = &EmptyStat{}

// ; \n
func (p *Parser) parseEmptyStat() *EmptyStat {
	p.lexer.NextTokenOfType(TOKEN_SEP_SEMI, TOKEN_SEP_EOLN)
	return emptyStat
}

func (p *Parser) parseBlockStat() *BlockStat {
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY)
	block := &BlockStat{Block: p.parseBlock()}
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY)
	return block
}

// break
func (p *Parser) parseBreakStat() *BreakStat {
	p.lexer.NextTokenOfType(TOKEN_KW_BREAK)
	return &BreakStat{
		Line: p.lexer.Line(),
	}
}

// continue
func (p *Parser) parseContinueStat() *ContinueStat {
	p.lexer.NextTokenOfType(TOKEN_KW_CONTINUE)
	return &ContinueStat{
		Line: p.lexer.Line(),
	}
}

// while {assignment ';'} exp '{' block '}'
func (p *Parser) parseWhileStat() *LoopStat {
	p.lexer.NextTokenOfType(TOKEN_KW_WHILE) // while
	peeks := p.lexer.PeekTokenOfType(TOKEN_SEP_LCURLY)

	var initList []Statement
	if count := TOKEN_SEP_SEMI.CountIn(peeks...); count > 0 {
		for ; count > 0; count-- {
			stat := p.parseAssignOrLocVarDeclOrFuncCallStat()
			if stat != nil {
				initList = append(initList, stat)
			}
			p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
		}
	}
	exp := p.parseExp()                       // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &LoopStat{
		InitList: initList,
		Exp:      exp,
		StepStat: nil,
		Block:    block,
	}
}

// if {assignment ';'} exp '{' block '}' {else if {assignment ';'} exp '{' block '}'} [else '{' block '}']
func (p *Parser) parseIfStat() *IfStat {
	subs := []*SubIfStat{p.parseSubIfStat()}

	for p.lexer.PeekToken().Is(TOKEN_KW_ELSE) {
		p.lexer.NextToken()
		switch p.lexer.PeekToken().Type {
		case TOKEN_KW_IF: // else if
			subs = append(subs, p.parseSubIfStat())
		case TOKEN_SEP_LCURLY: // else {
			break
		}
	}

	// else '{' block '}' => else if true '{' block '}'
	if p.lexer.PeekToken().Is(TOKEN_SEP_LCURLY) {
		exp := &TrueExp{Line: p.lexer.Line()}     // exps = true
		p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
		block := p.parseBlock()                   // block
		p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
		sub := &SubIfStat{
			Exp:   exp,
			Block: block,
		}
		subs = append(subs, sub)
	}

	return &IfStat{
		SubList: subs,
	}
}

func (p *Parser) parseSubIfStat() *SubIfStat {
	p.lexer.NextTokenOfType(TOKEN_KW_IF) // if
	peeks := p.lexer.PeekTokenOfType(TOKEN_SEP_LCURLY)

	var initList []Statement
	if count := TOKEN_SEP_SEMI.CountIn(peeks...); count > 0 {
		for ; count > 0; count-- {
			stat := p.parseAssignOrLocVarDeclOrFuncCallStat()
			if stat != nil {
				initList = append(initList, stat)
			}
			p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
		}
	}
	exp := p.parseExp()                       // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &SubIfStat{
		InitList: initList,
		Exp:      exp,
		Block:    block,
	}
}

// for for assignment ';' exp ';' assignment '{' block '}'
// for namelist in explist '{' block '}'
func (p *Parser) parseForStat() Statement {
	peeks := p.lexer.PeekTokenOfType(TOKEN_SEP_LCURLY) // {

	if TOKEN_KW_IN.In(peeks...) {
		return p.parseForInStat()
	} else {
		return p.parseForNumStat()
	}
}

// for for assignment ';' exp ';' assignment '{' block '}'
// => while assignment ';' exp {' block assignment '}'
func (p *Parser) parseForNumStat() *LoopStat {
	p.lexer.NextTokenOfType(TOKEN_KW_FOR)                 // for
	initStat := p.parseAssignOrLocVarDeclOrFuncCallStat() // assignment
	p.lexer.NextTokenOfType(TOKEN_SEP_SEMI)               // ;
	limitExp := p.parseExp()                              // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_SEMI)               // ;
	stepStat := p.parseAssignOrLocVarDeclOrFuncCallStat() // assignment

	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }

	return &LoopStat{
		InitList: []Statement{initStat},
		Exp:      limitExp,
		StepStat: stepStat,
		Block:    block,
	}
}

// for namelist in explist '{' block '}'
// namelist ::= Name {',' Name}
// explist ::= exp {',' exp}
func (p *Parser) parseForInStat() *ForInStat {
	p.lexer.NextTokenOfType(TOKEN_KW_FOR)              // for
	nameList := p.parseNameList()                      // namelist
	p.lexer.NextTokenOfType(TOKEN_KW_IN)               // in
	expList := p.parseExpList()                        // explist
	begin := p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                            // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY)          // }
	return &ForInStat{
		LineBlock: begin.Line,
		NameList:  nameList,
		ExpList:   expList,
		Block:     block,
	}
}

// namelist ::= Name {',' Name}
func (p *Parser) parseNameList() []string {
	nameList := []string{p.lexer.NextIdentifier().Literal}
	for p.lexer.PeekToken().Is(TOKEN_SEP_COMMA) {
		p.lexer.NextToken()              // ,
		name := p.lexer.NextIdentifier() // Name
		nameList = append(nameList, name.Literal)
	}
	return nameList
}

// local func Name funcbody
// local namelist ['=' explist]
func (p *Parser) parseLocAssignOrLocFuncDefStat() Statement {
	if p.lexer.PeekTokenN(2).Is(TOKEN_KW_FUNC) { // local func
		return p.parseLocalFuncDefStat()
	} else { // local identifier
		return p.parseLocVarDeclStat()
	}
}

// func f() {}          =>  f = func() {}
// func t.a.b.c.f() {}  =>  t.a.b.c.f = func() {}
// func t.a.b.c:f() {}  =>  t.a.b.c.f = func(self) {}
// local func f() {}    =>  local f; f = func() {}

// The statement `local func f () { body }`
// translates to `local f; f = func () { body }`
// not to `local f = func () { body }`
// (This only makes a difference when the body of the function
// contains references to f.)

// local func Name funcbody
func (p *Parser) parseLocalFuncDefStat() *Statements {
	p.lexer.NextTokenOfType(TOKEN_KW_LOCAL) // local
	p.lexer.NextTokenOfType(TOKEN_KW_FUNC)  // func
	name := p.lexer.NextIdentifier()        // name
	exp := p.parseFuncDefExp()              // funcbody
	lastLine := p.lexer.Line()

	decl := &LocVarDeclStat{
		LastLine: name.Line,
		NameList: []string{name.Literal},
	}

	assign := &AssignmentStat{
		LastLine: lastLine,
		VarList: []Expression{NameExp{
			Line: name.Line,
			Name: name.Literal,
		}},
		ExpList: []Expression{exp},
	}

	return &Statements{
		StatList: []Statement{
			decl,
			assign,
		},
	}
}

// func funcname funcbody
// funcname ::= Name {'.' Name} [':' Name]
// funcbody ::= '(' [parlist] ')' block end
// parlist ::= namelist [',' '...'] | '...'
// namelist ::= Name {',' Name}
func (p *Parser) parseFuncDefStat() *AssignmentStat {
	p.lexer.NextTokenOfType(TOKEN_KW_FUNC) // func
	fnExp, hasColon := p.parseFuncName()   // funcname
	fdExp := p.parseFuncDefExp()           // funcbody
	if hasColon {                          // insert self
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}

	return &AssignmentStat{
		LastLine: fdExp.Line,
		VarList:  []Expression{fnExp},
		ExpList:  []Expression{fdExp},
	}
}

// funcname ::= Name {'.' Name} [':' Name]
func (p *Parser) parseFuncName() (Expression, bool) {
	var exp Expression
	hasColon := false
	name := p.lexer.NextIdentifier()
	exp = &NameExp{
		Line: name.Line,
		Name: name.Literal,
	}

	for p.lexer.PeekToken().Is(TOKEN_SEP_DOT, TOKEN_SEP_COLON) {
		token := p.lexer.NextToken()
		name := p.lexer.NextIdentifier()
		idx := &StringExp{
			Line: name.Line,
			Str:  name.Literal,
		}
		exp = &TableAccessExp{
			LastLine:  name.Line,
			PrefixExp: exp,
			KeyExp:    idx,
		}

		if token.Is(TOKEN_SEP_COLON) {
			hasColon = true
			break
		}
	}

	return exp, hasColon
}

func (p *Parser) parseAssignOrLocVarDeclOrFuncCallStat() Statement {
	switch p.lexer.PeekToken().Type {
	case TOKEN_KW_LOCAL:
		return p.parseLocVarDeclStat()
	case TOKEN_IDENTIFIER, TOKEN_STRING, TOKEN_NUMBER, TOKEN_SEP_LPAREN:
		prefixExp := p.parsePrefixExp()
		if fc, ok := prefixExp.(*FuncCallExp); ok {
			return fc
		} else {
			return p.parseAssignmentStat(prefixExp)
		}
	default:
		return nil
	}

}

// varlist '=' explist
// functioncall
func (p *Parser) parseAssignOrFuncCallStat() Statement {
	prefixExp := p.parsePrefixExp()
	if fc, ok := prefixExp.(*FuncCallExp); ok {
		return fc
	} else {
		return p.parseAssignmentStat(prefixExp)
	}
}

// local namelist ['=' explist]
func (p *Parser) parseLocVarDeclStat() *LocVarDeclStat {
	p.lexer.NextTokenOfType(TOKEN_KW_LOCAL) // local
	names := p.parseNameList()              // namelist
	var exps []Expression
	if p.lexer.PeekToken().Is(TOKEN_OP_ASSIGN) {
		p.lexer.NextToken()     // =
		exps = p.parseExpList() // explist
	}
	lastLine := p.lexer.Line()
	return &LocVarDeclStat{
		LastLine: lastLine,
		NameList: names,
		ExpList:  exps,
	}
}

// varlist ('+=' | '-=' | '*=' | '/=' | '~/=' | '%='
//		| '&=' | '^=' | '|=' | '**=' | '<<=' | '>>=' | '=' | ':=') explist
func (p *Parser) parseAssignmentStat(var0 Expression) Statement {
	varList := p.parseVarList(var0) // varlist
	op := p.lexer.NextToken()       // operator
	var expList []Expression
	lastLine := p.lexer.Line()
	if len(varList) > 1 {
		if !op.Is(TOKEN_OP_ASSIGN, TOKEN_OP_LOCASSIGN) {
			p.Error("too many variables on left: %s", varList)
		} else {
			expList = p.parseExpList() // explist
		}
	} else if len(varList) == 1 {
		if op.Is(TOKEN_OP_ADDSELF, TOKEN_OP_SUBSELF) {
			newop, _ := op.Change()
			expList = []Expression{&BinopExp{
				Op:   newop,
				Exp1: varList[0],
				Exp2: &IntegerExp{
					Line: newop.Line,
					Val:  1,
				},
			}}
			op = &Token{op.Line, TOKEN_OP_ASSIGN, op.Literal}
		} else {
			expList = p.parseExpList() // explist
			if newop, ok := op.Change(); ok {
				if len(expList) > 1 {
					p.Error("too many expressions on right: %s", expList)
				} else {
					expList = []Expression{&BinopExp{
						Op:   newop,
						Exp1: varList[0],
						Exp2: expList[0],
					}}
					op = &Token{op.Line, TOKEN_OP_ASSIGN, op.Literal}
				}
			}
		}
	} else {
		p.Error("no variable on left")
	}

	if op.Is(TOKEN_OP_LOCASSIGN) {
		nameList := p.toName(varList...)
		return &LocVarDeclStat{
			LastLine: lastLine,
			NameList: nameList,
			ExpList:  expList,
		}
	} else if op.Is(TOKEN_OP_ASSIGN) {
		return &AssignmentStat{
			LastLine: lastLine,
			VarList:  varList,
			ExpList:  expList,
		}
	}
	p.Error("invalid operator: %s", op)
	panic("unreachable")
}

// varlist ::= var {',' var}
func (p *Parser) parseVarList(varList ...Expression) []Expression {
	var vars []Expression
	for _, v := range varList {
		vars = append(vars, p.checkVar(v))
	}
	if len(varList) == 0 {
		exp := p.parsePrefixExp() // var
		vars = append(vars, p.checkVar(exp))
	}

	for p.lexer.PeekToken().Is(TOKEN_SEP_COMMA) {
		p.lexer.NextToken()       // ,
		exp := p.parsePrefixExp() // var
		vars = append(vars, p.checkVar(exp))
	}
	return vars
}

// var ::=  Name | prefixexp '[' Expression ']' | prefixexp '.' Name
func (p *Parser) checkVar(exp Expression) Expression {
	switch exp.(type) {
	case *NameExp, *TableAccessExp:
		return exp
	}
	p.Error("not a variable: %s", exp)
	panic("unreachable")
}

func (p *Parser) toName(expList ...Expression) []string {
	var nameList []string
	for _, exp := range expList {
		if name, ok := exp.(*NameExp); ok {
			nameList = append(nameList, name.Name)
		} else {
			p.Error("not a identifer: %s", exp)
		}
	}
	return nameList
}
