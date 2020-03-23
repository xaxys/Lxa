package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
)

// varlist '=' explist
// functioncall
func (p *Parser) parseAssignOrFuncCallAsn() Assignment {
	prefixExp := p.parsePrefixExp()
	if fc, ok := prefixExp.(*FuncCallExp); ok {
		return fc
	} else {
		return p.parseAssignAsn(prefixExp)
	}
}

// local namelist ['=' explist]
func (p *Parser) parseLocVarDeclAsn() *LocVarDeclAsn {
	p.lexer.NextTokenOfType(TOKEN_KW_LOCAL) // local
	names := p.parseNameList()              // namelist
	var exps []Expression
	if p.lexer.PeekToken().Is(TOKEN_OP_ASSIGN) {
		p.lexer.NextToken()     // =
		exps = p.parseExpList() // explist
	}
	lastLine := p.lexer.Line()
	return &LocVarDeclAsn{
		LastLine: lastLine,
		NameList: names,
		ExpList:  exps,
	}
}

// varlist ('+=' | '-=' | '*=' | '/=' | '~/=' | '%='
//		| '&=' | '^=' | '|=' | '**=' | '<<=' | '>>=' | '=' | ':=') explist
func (p *Parser) parseAssignAsn(var0 Expression) Assignment {
	varList := p.parseVarList(var0) // varlist
	op := p.lexer.NextToken()       // operator
	var expList []Expression
	lastLine := p.lexer.Line()
	if len(varList) > 1 {
		if !op.Is(TOKEN_OP_ASSIGN, TOKEN_OP_LASSIGN) {
			p.error("too many variables on left: %v", varList)
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
					p.error("too many expressions on right: %v", expList)
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
		p.error("no variable on left")
	}

	if op.Is(TOKEN_OP_LASSIGN) {
		nameList := p.toName(varList...)
		return &LocVarDeclAsn{
			LastLine: lastLine,
			NameList: nameList,
			ExpList:  expList,
		}
	} else if op.Is(TOKEN_OP_ASSIGN) {
		return &AssignAsn{
			LastLine: lastLine,
			VarList:  varList,
			ExpList:  expList,
		}
	}
	p.error("invalid operator: %v", op)
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
	p.error("not a variable: %v", exp)
	panic("unreachable")
}

func (p *Parser) toName(expList ...Expression) []string {
	var nameList []string
	for _, exp := range expList {
		if name, ok := exp.(*NameExp); ok {
			nameList = append(nameList, name.Name)
		} else {
			p.error("not a identifer: %v", exp)
		}
	}
	return nameList
}
