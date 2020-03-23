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

	var assignList []Assignment
	if count := TOKEN_SEP_SEMI.CountIn(peeks...); count > 0 {
		for ; count > 0; count-- {
			assignList = append(assignList, p.parseAssignOrFuncCallAsn())
			p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
		}
	}
	exp := p.parseExp()                       // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &LoopStat{
		AsnList: assignList,
		Exp:     exp,
		StepAsn: nil,
		Block:   block,
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

	var assignList []Assignment
	if count := TOKEN_SEP_SEMI.CountIn(peeks...); count > 0 {
		for ; count > 0; count-- {
			assignList = append(assignList, p.parseAssignOrFuncCallAsn())
			p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
		}
	}
	exp := p.parseExp()                       // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &SubIfStat{
		AsnList: assignList,
		Exp:     exp,
		Block:   block,
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
	p.lexer.NextTokenOfType(TOKEN_KW_FOR)   // for
	initAsn := p.parseAssignOrFuncCallAsn() // assignment
	p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
	limitExp := p.parseExp()                // exp
	p.lexer.NextTokenOfType(TOKEN_SEP_SEMI) // ;
	stepAsn := p.parseAssignOrFuncCallAsn() // assignment

	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	block := p.parseBlock()                   // block
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }

	return &LoopStat{
		AsnList: []Assignment{initAsn},
		Exp:     limitExp,
		StepAsn: stepAsn,
		Block:   block,
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

// local varlist '=' explist
func (p *Parser) parseLocVarDeclStat() Statement {
	return &AssignmentStat{Asn: p.parseLocVarDeclAsn()}
}

// varlist '=' explist
// functioncall
func (p *Parser) parseAssignOrFuncCallStat() Statement {
	return &AssignmentStat{Asn: p.parseAssignOrFuncCallAsn()}
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

	decl := &LocVarDeclAsn{
		LastLine: name.Line,
		NameList: []string{name.Literal},
	}

	assign := &AssignAsn{
		LastLine: lastLine,
		VarList: []Expression{NameExp{
			Line: name.Line,
			Name: name.Literal,
		}},
		ExpList: []Expression{exp},
	}

	return &Statements{
		StatList: []Statement{
			&AssignmentStat{Asn: decl},
			&AssignmentStat{Asn: assign},
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
		Asn: &AssignAsn{
			LastLine: fdExp.Line,
			VarList:  []Expression{fnExp},
			ExpList:  []Expression{fdExp},
		},
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
