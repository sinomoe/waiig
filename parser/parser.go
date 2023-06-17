package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type (
	// prefixParseFn 前缀解析函数
	prefixParseFn func() ast.Expression
	// infixParseFn 中缀解析函数，接受的参数为中缀运算符左边的表达式，由于前缀运算符左边没有表达式，故无参数
	infixParseFn func(ast.Expression) ast.Expression
)

const (
	// 定义运算符优先级
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// precedences 中缀表达式优先级表
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

// Parser 是语法解析器，负责将词法单元解析为 AST
type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token // 输入中的当前词法单元
	peekToken token.Token // 下一个词法单元

	prefixParseFns map[token.TokenType]prefixParseFn // 存放处理前缀词法单元的解析函数
	infixParseFns  map[token.TokenType]infixParseFn  // 存放处理中缀词法单元的解析函数
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// 初始化前缀解析函数，标识符和字面量部署运算符，属于特殊的前缀解析函数
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier) // 这四个解析函数相当于递归的 base case，因为他们的解析还书里不包含对 parseExpression 的递归调用
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression) // 解析括号表达式
	p.registerPrefix(token.IF, p.parseIfExpression)          // 解析 if 表达式
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	// 初始化中缀表达式解释函数
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression) // 解析函数调用,  把函数调用当作中缀表达式
	// 读取两个词法单元，以设置curToken和peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// peekError 向 l.errors 中追加错误信息
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{} // 构建 AST 根节点
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF { // 循环将 Token 读完
		stmt := p.parseStatement() // Program 由语句组成，循环解析语句
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement 解析语句
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement 解析 let 语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken} // 初始化 let 语句节点
	// let 语句前两个 token 一定是 IDENT 和 ASSIGN
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = p.parseIdentifier().(*ast.Identifier)
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken() // curToken=表达式第一个 token
	stmt.Value = p.parseExpression(LOWEST)
	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}
	return stmt
}

// parseReturnStatement 解析 return 语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	// 指向 return 的下一个 token
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}
	return stmt
}

// parseExpressionStatement 表达式语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseExpression 表达式解析函数, 调用时 curToken 是表达式第一个 token, 返回后 curToken 是表达式最后一个 token
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// parseIdentifier 标识符表达式解析函数
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 整数字面量表达式解析函数
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseBooleanLiteral 布尔字面量解析函数
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(token.TRUE),
	}
}

// parsePrefixExpression 前缀表达式解析函数
func (p *Parser) parsePrefixExpression() ast.Expression {
	pe := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	pe.Right = p.parseExpression(PREFIX)
	return pe
}

// parseInfixExpression 前缀表达式解析函数
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	ie := &ast.InfixExpression{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
		Right:    nil,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	ie.Right = p.parseExpression(precedence)
	return ie
}

// parseGroupedExpression 解析分组(括号)表达式 (a+b)
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// parseIfExpression 解析 if 表达式
func (p *Parser) parseIfExpression() ast.Expression {
	ie := &ast.IfExpression{
		Token: p.curToken,
	}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	ie.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	ie.Consequence = p.parseBlockStatement()
	if !p.peekTokenIs(token.ELSE) {
		return ie
	}
	p.nextToken()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	ie.Alternative = p.parseBlockStatement()
	return ie
}

// parseBlockStatement 解析语句块 {}
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	bs := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}
	p.nextToken() // 指向 { 的下一个 token
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		bs.Statements = append(bs.Statements, p.parseStatement())
		p.nextToken()
	}
	return bs
}

// parseFunctionLiteral 函数字面量解析函数
func (p *Parser) parseFunctionLiteral() ast.Expression {
	fl := &ast.FunctionLiteral{
		Token:      p.curToken,
		Parameters: []*ast.Identifier{},
	}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fl.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	fl.Body = p.parseBlockStatement()
	return fl
}

// parseFunctionParameters 解析函数形参列表, (a,b,c) () (a)
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	ids := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return ids
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	ids = append(ids, p.parseIdentifier().(*ast.Identifier))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		ids = append(ids, p.parseIdentifier().(*ast.Identifier))
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return ids
}

// parseCallExpression 函数调用解析函数, 函数调用是一种中缀表达式, 左边的表达式是函数
func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	return &ast.CallExpression{
		Token:     p.curToken,
		Function:  left,
		Arguments: p.parseCallExpressionArguments(),
	}
}

// parseCallExpressionArguments 解析函数实参列表, (a,b,c) () (a) (a+1, b, 3)
func (p *Parser) parseCallExpressionArguments() []ast.Expression {
	args := []ast.Expression{}
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return args
	}
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // curToken=COMMA
		p.nextToken() // curToken=表达式第一个 token
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}
