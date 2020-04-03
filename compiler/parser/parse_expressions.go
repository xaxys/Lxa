package parser

import (
	. "lxa/compiler/ast"
	. "lxa/compiler/token"
	"lxa/number"
)

// exp ::=  nil | false | true | Numeral | LiteralString | '...' | functiondef |
// 	 prefixexp | tableconstructor | exp binop exp | unop exp

// exp   ::= exp12
// exp12 ::= exp11 { '?' exp11}
// exp11 ::= exp10 {('||' | or) exp10}
// exp10 ::= exp9 {('&&' | and) exp9}
// exp9  ::= exp8 {('<' | '>' | '<=' | '>=' | '!=' | '==') exp8}
// exp8  ::= exp7 {'|' exp7}
// exp7  ::= exp6 {'^' exp6}
// exp6  ::= exp5 {'&' exp5}
// exp5  ::= exp4 {('<<' | '>>') exp4}
// exp4  ::= exp3 {('+' | '-') exp3}
// exp3  ::= exp2 {('*' | '/' | '//' | '%') exp2}
// exp2  ::= {('!' | not' | '#' | '-' | '~')} exp1
// exp1  ::= exp0 {'**' exp2}
// exp0  ::= nil | false | true | Numeral | LiteralString
//		| '...' | functiondef | prefixexp | tableconstructor

// explist ::= exp {',' exp}
func (p *Parser) parseExpList() []Expression {
	expList := []Expression{p.parseExp()}
	for p.lexer.PeekToken().Is(TOKEN_SEP_COMMA) { // ,
		p.lexer.NextToken()
		expList = append(expList, p.parseExp())
	}
	return expList
}

func (p *Parser) parseExp() Expression {
	return p.parseExp13()
}

// x?
func (p *Parser) parseExp13() Expression {
	exp := p.parseExp12()
	for p.lexer.PeekToken().Is(TOKEN_OP_QST) {
		op := p.lexer.NextToken()
		lqst := &UnopExp{
			Op:  op,
			Exp: exp,
		}
		exp = OptimizeLogicalQst(lqst)
	}
	return exp
}

// x ('||' | or) y
func (p *Parser) parseExp12() Expression {
	exp := p.parseExp11()
	if !p.lexer.PeekToken().Is(TOKEN_OP_OR) {
		return exp
	}

	expList := []Expression{exp}
	var op *Token
	for p.lexer.PeekToken().Is(TOKEN_OP_OR) {
		op = p.lexer.NextToken()
		expList = append(expList, p.parseExp11())
	}
	lor := &LogicalExp{
		Op:      op,
		ExpList: expList,
	}

	return OptimizeLogicalOr(lor)
}

// x ('&&' | and) y
func (p *Parser) parseExp11() Expression {
	exp := p.parseExp10()
	if !p.lexer.PeekToken().Is(TOKEN_OP_AND) {
		return exp
	}

	expList := []Expression{exp}
	var op *Token
	for p.lexer.PeekToken().Is(TOKEN_OP_AND) {
		op = p.lexer.NextToken()
		expList = append(expList, p.parseExp10())
	}
	land := &LogicalExp{
		Op:      op,
		ExpList: expList,
	}

	return OptimizeLogicalAnd(land)
}

// compare
func (p *Parser) parseExp10() Expression {
	exp := p.parseExp9()
	for p.lexer.PeekToken().Is(
		TOKEN_OP_LT,   // <
		TOKEN_OP_GT,   // >
		TOKEN_OP_NE,   // !=
		TOKEN_OP_LE,   // <=
		TOKEN_OP_GE,   // >=
		TOKEN_OP_EQ) { // ==
		op := p.lexer.NextToken()
		exp = &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp9(),
		}
	}
	return exp
}

// x .. y
func (p *Parser) parseExp9() Expression {
	exp := p.parseExp8()
	if !p.lexer.PeekToken().Is(TOKEN_OP_CONCAT) {
		return exp
	}

	var token *Token
	expList := []Expression{exp}
	for p.lexer.PeekToken().Is(TOKEN_OP_CONCAT) {
		token = p.lexer.NextToken()
		expList = append(expList, p.parseExp8())
	}
	return &ConcatExp{
		Line:    token.Line,
		ExpList: expList,
	}
}

// x | y
func (p *Parser) parseExp8() Expression {
	exp := p.parseExp7()
	for p.lexer.PeekToken().Is(TOKEN_OP_BOR) {
		op := p.lexer.NextToken()
		bor := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp7(),
		}
		exp = OptimizeBitwiseBinaryOp(bor)
	}
	return exp
}

// x ^ y
func (p *Parser) parseExp7() Expression {
	exp := p.parseExp6()
	for p.lexer.PeekToken().Is(TOKEN_OP_BXOR) {
		op := p.lexer.NextToken()
		bxor := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp6(),
		}
		exp = OptimizeBitwiseBinaryOp(bxor)
	}
	return exp
}

// x & y
func (p *Parser) parseExp6() Expression {
	exp := p.parseExp5()
	for p.lexer.PeekToken().Is(TOKEN_OP_BAND) {
		op := p.lexer.NextToken()
		band := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp5(),
		}
		exp = OptimizeBitwiseBinaryOp(band)
	}
	return exp
}

// shift
func (p *Parser) parseExp5() Expression {
	exp := p.parseExp4()
	for p.lexer.PeekToken().Is(TOKEN_OP_SHL, TOKEN_OP_SHR) {
		op := p.lexer.NextToken()
		shx := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp4(),
		}
		exp = OptimizeBitwiseBinaryOp(shx)
	}
	return exp
}

// x +/- y
func (p *Parser) parseExp4() Expression {
	exp := p.parseExp3()
	for p.lexer.PeekToken().Is(TOKEN_OP_ADD, TOKEN_OP_SUB) {
		op := p.lexer.NextToken()
		arith := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp3(),
		}
		exp = OptimizeArithBinaryOp(arith)
	}
	return exp
}

// *, %, /, //
func (p *Parser) parseExp3() Expression {
	exp := p.parseExp2()
	for p.lexer.PeekToken().Is(
		TOKEN_OP_MUL,
		TOKEN_OP_MOD,
		TOKEN_OP_DIV,
		TOKEN_OP_IDIV) {
		op := p.lexer.NextToken()
		arith := &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp2(),
		}
		exp = OptimizeArithBinaryOp(arith)
	}
	return exp
}

// unary
func (p *Parser) parseExp2() Expression {
	if p.lexer.PeekToken().Is(
		TOKEN_OP_UNM,
		TOKEN_OP_BNOT,
		TOKEN_OP_LEN,
		TOKEN_OP_NOT) {
		op := p.lexer.NextToken()
		exp := &UnopExp{
			Op:  op,
			Exp: p.parseExp2(),
		}
		return OptimizeUnaryOp(exp)
	}
	return p.parseExp1()
}

// x ** y
func (p *Parser) parseExp1() Expression { // pow is right associative
	exp := p.parseExp0()
	if p.lexer.PeekToken().Is(TOKEN_OP_POW) {
		op := p.lexer.NextToken()
		exp = &BinopExp{
			Op:   op,
			Exp1: exp,
			Exp2: p.parseExp2(),
		}
	}
	return OptimizePow(exp)
}

func (p *Parser) parseExp0() Expression {
	switch p.lexer.PeekToken().Type {
	case TOKEN_VARARG: // ...
		op := p.lexer.NextToken()
		return &VarargExp{Line: op.Line}
	case TOKEN_KW_NIL: // nil
		op := p.lexer.NextToken()
		return &NilExp{Line: op.Line}
	case TOKEN_KW_TRUE: // true
		op := p.lexer.NextToken()
		return &TrueExp{Line: op.Line}
	case TOKEN_KW_FALSE: // false
		op := p.lexer.NextToken()
		return &FalseExp{Line: op.Line}
	case TOKEN_STRING: // LiteralString
		token := p.lexer.NextToken()
		return &StringExp{
			Line: token.Line,
			Str:  token.Literal,
		}
	case TOKEN_NUMBER: // Numeral
		return p.parseNumberExp()
	case TOKEN_SEP_LCURLY: // tableconstructor
		return p.parseTableConstructorExp()
	case TOKEN_KW_FUNC: // functiondef
		p.lexer.NextToken()
		return p.parseFuncDefExp()
	default: // prefixexp
		return p.parsePrefixExp()
	}
}

func (p *Parser) parseNumberExp() Expression {
	token := p.lexer.NextToken()
	if i, ok := number.ParseInteger(token.Literal); ok {
		return &IntegerExp{
			Line: token.Line,
			Val:  i,
		}
	} else if f, ok := number.ParseFloat(token.Literal); ok {
		return &FloatExp{
			Line: token.Line,
			Val:  f,
		}
	} else {
		p.Error("not a number: %s", token)
	}
	panic("unreachable")
}

// lambda ::= '(' [parlist] ')' => '{' block '}'
func (p *Parser) parseLambda() *FuncDefExp {
	line := p.lexer.Line()
	p.lexer.NextTokenOfType(TOKEN_SEP_LPAREN)        // (
	parList, isVararg := p.parseParList()            // [parlist]
	p.lexer.NextTokenOfType(TOKEN_SEP_RPAREN)        // )
	p.lexer.NextTokenOfType(TOKEN_OP_ARROW)          // =>
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY)        // {
	block := p.parseBlock()                          // block
	end := p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &FuncDefExp{
		Line:     line,
		LastLine: end.Line,
		ParList:  parList,
		IsVararg: isVararg,
		Block:    block,
	}
}

// functiondef ::= func funcbody | lambda
// funcbody ::= '(' [parlist] ')' '{' block '}'
func (p *Parser) parseFuncDefExp() *FuncDefExp {
	line := p.lexer.Line()                           // func
	p.lexer.NextTokenOfType(TOKEN_SEP_LPAREN)        // (
	parList, isVararg := p.parseParList()            // [parlist]
	p.lexer.NextTokenOfType(TOKEN_SEP_RPAREN)        // )
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY)        // {
	block := p.parseBlock()                          // block
	end := p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	return &FuncDefExp{
		Line:     line,
		LastLine: end.Line,
		ParList:  parList,
		IsVararg: isVararg,
		Block:    block,
	}
}

// [parlist]
// parlist ::= namelist [',' '...'] | '...'
func (p *Parser) parseParList() ([]string, bool) {
	switch p.lexer.PeekToken().Type {
	case TOKEN_SEP_RPAREN:
		return nil, false
	case TOKEN_VARARG:
		p.lexer.NextToken()
		return nil, true
	}

	isVararg := false
	nameList := []string{p.lexer.NextIdentifier().Literal}
	for p.lexer.PeekToken().Is(TOKEN_SEP_COMMA) {
		p.lexer.NextToken()
		if p.lexer.PeekToken().Is(TOKEN_IDENTIFIER) {
			nameList = append(nameList, p.lexer.NextIdentifier().Literal)
		} else {
			p.lexer.NextTokenOfType(TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return nameList, isVararg
}

// tableconstructor ::= '{' [fieldlist] '}'
func (p *Parser) parseTableConstructorExp() *TableConstructorExp {
	line := p.lexer.Line()
	p.lexer.NextTokenOfType(TOKEN_SEP_LCURLY) // {
	keyExps, valExps := p.parseFieldList()    // [fieldlist]
	p.lexer.NextTokenOfType(TOKEN_SEP_RCURLY) // }
	lastLine := p.lexer.Line()
	return &TableConstructorExp{
		Line:     line,
		LastLine: lastLine,
		KeyExps:  keyExps,
		ValExps:  valExps,
	}
}

// fieldlist ::= field {fieldsep field} [fieldsep]
func (p *Parser) parseFieldList() ([]Expression, []Expression) {
	var ks, vs []Expression
	if !p.lexer.PeekToken().Is(TOKEN_SEP_RCURLY) {
		k, v := p.parseField()
		ks = append(ks, k)
		vs = append(vs, v)

		for p.lexer.PeekToken().Is(TOKEN_SEP_COMMA, TOKEN_SEP_SEMI) {
			p.lexer.NextToken()
			if !p.lexer.PeekToken().Is(TOKEN_SEP_RCURLY) {
				k, v := p.parseField()
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return ks, vs
}

// field ::= '[' exp ']' '=' exp | Name '=' exp | exp
func (p *Parser) parseField() (Expression, Expression) {
	var k, v Expression
	if p.lexer.PeekToken().Is(TOKEN_SEP_LBRACK) {
		p.lexer.NextToken()                       // [
		k = p.parseExp()                          // exp
		p.lexer.NextTokenOfType(TOKEN_SEP_RBRACK) // ]
		p.lexer.NextTokenOfType(TOKEN_OP_ASSIGN)  // =
		v = p.parseExp()                          // exp
		return k, v
	}

	exp := p.parseExp()
	if nameExp, ok := exp.(*NameExp); ok {
		if p.lexer.PeekToken().Is(TOKEN_OP_ASSIGN) {
			// Name '=' exp => '[' LiteralString ']' = exp
			p.lexer.NextToken()
			k = &StringExp{
				Line: nameExp.Line,
				Str:  nameExp.Name,
			}
			v = p.parseExp()
			return k, v
		}
	}

	return nil, exp
}
