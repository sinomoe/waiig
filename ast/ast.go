package ast

import "monkey/token"

// Node AST 中的每个节 点都必须实现 Node 接口，也就是说必须提供 TokenLiteral()方法，该方法返回与其 关联的词法单元的字面量
type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode() // 仅占位，防止 Statement 和 Expression 混淆
}

type Expression interface {
	Node
	expressionNode() // 仅占位，防止 Statement 和 Expression 混淆
}

// Program 节点是语法分析器生成的每个 AST 的根节点
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// LetStatement let 语句节点
type LetStatement struct {
	Token token.Token // token.LET 词法单元
	Name  *Identifier // 左侧标识符
	Value Expression  // 右侧表达式、字面量
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token token.Token // token.IDENT 词法单元
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type ReturnStatement struct {
	Token token.Token
	Value Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
