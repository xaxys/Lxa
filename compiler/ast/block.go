package ast

// chunk ::= block
// type Chunk *Block

// block ::= {stat} [retstat]
// retstat ::= return [explist] [';']
// explist ::= exp {',' exp}
type Block struct {
	LastLine   int
	Statements []Statement
	ReturnExps []Expression
}
