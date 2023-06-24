package ast

import (
	"bytes"
	"monkey/token"
	"strings"
)

// Node AST 中的每个节 点都必须实现 Node 接口，也就是说必须提供 TokenLiteral()方法，该方法返回与其 关联的词法单元的字面量
type Node interface {
	TokenLiteral() string
	String() string
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

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// LetStatement let 语句节点
type LetStatement struct {
	Token token.Token // token.LET 词法单元
	Name  *Identifier // 左侧标识符
	Value Expression  // 右侧表达式、字面量
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type FunctionDeclarationStatement struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionDeclarationStatement) statementNode()       {}
func (fs *FunctionDeclarationStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionDeclarationStatement) String() string {
	var out bytes.Buffer
	var params []string
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn ")
	out.WriteString(fs.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(" ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// AssignExpression 赋值表达式节点
type AssignExpression struct {
	Token token.Token // token.ASSIGN 词法单元
	Left  Expression  // 左侧标识符、索引表达式
	Value Expression  // 右侧表达式、字面量
}

func (ae *AssignExpression) expressionNode()      {}
func (ae *AssignExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignExpression) String() string {
	var out bytes.Buffer
	out.WriteByte('(')
	out.WriteString(ae.Left.String())
	out.WriteString("=")
	out.WriteString(ae.Value.String())
	out.WriteByte(')')
	return out.String()
}

// Identifier 标识符表达式
type Identifier struct {
	Token token.Token // token.IDENT 词法单元
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	return i.Value
}

// ReturnStatement 返回语句
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// BlockStatement 由 {} 包裹的多条语句组成的语句块
type BlockStatement struct {
	Token      token.Token // { 词法单元
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	var stmts []string
	for _, stmt := range bs.Statements {
		stmts = append(stmts, stmt.String())
	}
	out.WriteString(strings.Join(stmts, " "))
	return out.String()
}

// ExpressionStatement 表达式语句，用于表示单行表达式，例如 x+1;
type ExpressionStatement struct {
	Token      token.Token // 该表达式中的第一个词法单元
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IntegerLiteral 整数字面量表达式
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral 浮点数字面量表达式
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// BooleanLiteral 布尔字面量表达式节点
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// StringLiteral 字符串字面量节点
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// PrefixExpression  前缀表达式
type PrefixExpression struct {
	Token    token.Token // 前缀词法单元，如!
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression 中缀表达式
type InfixExpression struct {
	Token    token.Token // 运算符词法单元，如+
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IfExpression if 表达式
type IfExpression struct {
	Token       token.Token // if 词法单元
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// FunctionLiteral 函数字面量表达式节点
type FunctionLiteral struct {
	Token      token.Token     // fn 词法单元
	Parameters []*Identifier   // 形参列表
	Body       *BlockStatement // 语句块
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	var params []string
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(" ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression 函数调用表达式节点
type CallExpression struct {
	Token     token.Token  // ( 词法单元
	Function  Expression   // 标识符或者字面量
	Arguments []Expression // 实参
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	args := []string{}
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ArrayLiteral 数组字面量节点
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var buf bytes.Buffer
	var elms []string
	for _, elm := range al.Elements {
		elms = append(elms, elm.String())
	}
	buf.WriteByte('[')
	buf.WriteString(strings.Join(elms, ", "))
	buf.WriteByte(']')
	return buf.String()
}

// IndexExpression 数组索引表达值节点
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var buf bytes.Buffer
	buf.WriteByte('(')
	buf.WriteString(ie.Left.String())
	buf.WriteByte('[')
	buf.WriteString(ie.Index.String())
	buf.WriteByte(']')
	buf.WriteByte(')')
	return buf.String()
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
