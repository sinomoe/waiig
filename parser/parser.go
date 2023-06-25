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
	// 运算符的左结合力一定等于如下优先级
	// 右结合力在中缀解析函数中制定，一般也等于如下优先级
	// 但是也有特殊情况，比如赋值运算符的右结合力不等于左结合力
	_           int = iota
	LOWEST          // 最低优先级定义为 1 也是有用意的, 遇到其他未定义优先级的 token, 则优先级都为 0
	_               // 赋值表达式左右结合力不同，这里空一个保证赋值运算符优先级总高于 	LOWEST          // 最低优先级定义为 1 也是有用意的, 遇到其他未定义优先级的 token, 则优先级都为 0
	ASSIGN          // =
	EQUALS          // ==
	LESSGREATER     // > or <
	SUM             // +
	PRODUCT         // *
	PREFIX          // -X or !X
	CALL            // myFunction(X)
	INDEX           // a[i]
)

// precedences 中缀表达式优先级表
var precedences = map[token.TokenType]int{
	token.ASSIGN:   ASSIGN,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
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
	p.registerPrefix(token.IDENT, p.parseIdentifier) // 这六个解析函数相当于递归的 base case，因为他们的解析还书里不包含对 parseExpression 的递归调用
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression) // 解析括号表达式
	p.registerPrefix(token.IF, p.parseIfExpression)          // 解析 if 表达式
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral) // 解析数组字面量
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)    // 解析哈希表字面量

	// 初始化中缀表达式解释函数
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression) // 解析赋值表达式
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)    // 解析函数调用,  把函数调用当作中缀表达式
	p.registerInfix(token.LBRACKET, p.parseIndexExpression) // 解析数组索引

	// 读取两个词法单元，以设置curToken和peekToken
	p.nextToken() // curToken=nil peekToken=第一个 token
	p.nextToken() // curToken=第一个词法单元 peekToken=第二个词法单元
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
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

// peekPrecedence 下一个词法单元的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence 当前词法单元的优先级
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
		// parseStatement 后当前词元为当前语句的最后一个词元
		// 调用 nextToken 将当前词元指向下调语句的第一个词元
		p.nextToken()
	}
	return program
}

// parseStatement 解析语句
// 解析语句时当前词元指向语句的第一个词元 返回时当前词元为语句的最后一个词元
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FUNCTION:
		// 可能是函数申明
		if p.peekTokenIs(token.IDENT) {
			return p.parseFunctionDeclarationStatement()
		}
		// 也可能是匿名函数字面量表达式
		if p.peekTokenIs(token.LPAREN) {
			return p.parseExpressionStatement()
		}
	}
	return p.parseExpressionStatement()
}

// parseLetStatement 解析 let 语句
// let <identifier> = <expression>;
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
	if p.peekTokenIs(token.SEMICOLON) { // 允许 let 语句后不带分号
		p.nextToken()
	}
	return stmt
}

// parseReturnStatement 解析 return 语句
// return <expression>;
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	// 指向 return 的下一个 token
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) { // 允许 return 语句后不带分号
		p.nextToken()
	}
	return stmt
}

// parseFunctionDeclarationStatement 解析 function 申明语句
// fn <identifier>(<identifier>,...) <blockstatement>
func (p *Parser) parseFunctionDeclarationStatement() *ast.FunctionDeclarationStatement {
	stmt := &ast.FunctionDeclarationStatement{
		Token:      p.curToken,
		Name:       &ast.Identifier{},
		Parameters: []*ast.Identifier{},
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = p.parseIdentifier().(*ast.Identifier)
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	stmt.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

// parseExpressionStatement 表达式语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) { // 允许表达式语句后不带分号
		p.nextToken()
	}
	return stmt
}

// parseExpression 表达式解析函数
// 调用时 curToken 是表达式第一个 token, 返回后 curToken 是表达式最后一个 token
// 处理表达式时不要吞掉句末的 ; 词元, 吞掉 ; 词元统一交给语句解析式处理, 这里对应的是 parseExpressionStatement
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	// 以下循环结构要求
	// 前缀表达式解析函数: 调用时 curToken=表达式第一个 token, 返回时 curToken=表达式最后一个 token
	// 中缀表达式解析函数: 调用时 curToken=中缀运算符, 返回时 curToken=表达式最后一个 token
	// precedence 为左边运算符的结合力
	// peekPrecedence 为右边运算符的结合力
	// 只有右边的结合力强于左边时 将左表达式做为中缀表达式左节点
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() { // 由于优先级可以一直变大所以需要向右循环
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()            // curToken 转移到中缀运算符
		leftExp = infix(leftExp) // left 变成后一个运算符的左节点
	}
	return leftExp // left 变成前一个预算符的右节点
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

// parseFloatLiteral 浮点数字面量表达式解析函数
func (p *Parser) parseFloatLiteral() ast.Expression {
	fl := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	fl.Value = value
	return fl
}

// parseBooleanLiteral 布尔字面量解析函数
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{
		Token: p.curToken,
		Value: p.curTokenIs(token.TRUE),
	}
}

// parseStringLiteral 字符串字面量解析函数
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parsePrefixExpression 前缀表达式解析函数
// -<expression>
// !<expression>
func (p *Parser) parsePrefixExpression() ast.Expression {
	pe := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	pe.Right = p.parseExpression(PREFIX)
	return pe
}

// parseInfixExpression 中缀表达式解析函数
// <expression> + <expression>
// <expression> - <expression>
// <expression> * <expression>
// <expression> / <expression>
// <expression> == <expression>
// <expression> > <expression>
// <expression> < <expression>
// <expression> >= <expression>
// <expression> <= <expression>
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
// (<expression>)
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// parseIfExpression 解析 if 表达式
// if (<expression>) <blockstatement> else <blockstatement>
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
// { <statement>;... }
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
// fn(<identifier>,...) <blockstatement>
// fn() <blockstatement>
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
// <functionLiteral>(<expression>,...)
// <identifier>(<expression>,...)
func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	return &ast.CallExpression{
		Token:     p.curToken, // ( 词元
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

// parseArrayLiteral 解析数组字面量
func (p *Parser) parseArrayLiteral() ast.Expression {
	al := &ast.ArrayLiteral{
		Token:    p.curToken,
		Elements: []ast.Expression{},
	}
	p.nextToken()
	if p.curTokenIs(token.RBRACKET) {
		return al
	}
	al.Elements = append(al.Elements, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // curToken=COMMA
		p.nextToken() // curToken=表达式第一个 token
		al.Elements = append(al.Elements, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return al
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	ie := &ast.IndexExpression{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken()
	ie.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return ie
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hl := &ast.HashLiteral{
		Token: p.curToken,
		Pairs: map[ast.Expression]ast.Expression{},
	}
	p.nextToken()
	if p.curTokenIs(token.RBRACE) {
		return hl
	}

	key, val, ok := p.parseKeyValPair()
	if !ok {
		return nil
	}
	hl.Pairs[key] = val

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		key, val, ok = p.parseKeyValPair()
		if !ok {
			return nil
		}
		hl.Pairs[key] = val
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hl
}

func (p *Parser) parseKeyValPair() (key, val ast.Expression, ok bool) {
	key = p.parseExpression(LOWEST)
	if !p.expectPeek(token.COLON) {
		return
	}
	p.nextToken()
	val = p.parseExpression(LOWEST)
	ok = true
	return
}

// parseAssignExpression 赋值表达式解析函数
// <identifier> = <expression>
// <expression> = <identifier> = <expression>
func (p *Parser) parseAssignExpression(left ast.Expression) ast.Expression {
	ae := &ast.AssignExpression{
		Token: p.curToken,
		Left:  left,
	}
	p.nextToken()
	// 降低 = 的右结合力 保证连等赋值时 从右往左赋值
	ae.Value = p.parseExpression(ASSIGN - 1)
	return ae
}
