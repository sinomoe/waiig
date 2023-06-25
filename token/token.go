package token

import "strings"

type TokenType byte

func (t TokenType) String() string {
	return tokenTypeStringMap[t]
}

var tokenTypeStringMap = map[TokenType]string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	IDENT:     "IDENT",
	INT:       "INT",
	FLOAT:     "FLOAT",
	STRING:    "STRING",
	ASSIGN:    "=",
	PLUS:      "+",
	MINUS:     "-",
	BANG:      "!",
	ASTERISK:  "*",
	SLASH:     "/",
	LT:        "<",
	LTE:       "<=",
	GT:        ">",
	GTE:       ">=",
	EQ:        "==",
	NOT_EQ:    "!=",
	DOT:       ".",
	COMMA:     ",",
	SEMICOLON: ";",
	COLON:     ":",
	LPAREN:    "(",
	RPAREN:    ")",
	LBRACE:    "{",
	RBRACE:    "}",
	LBRACKET:  "[",
	RBRACKET:  "]",
	FUNCTION:  "FUNCTION",
	LET:       "LET",
	TRUE:      "TRUE",
	FALSE:     "FALSE",
	IF:        "IF",
	ELSE:      "ELSE",
	RETURN:    "RETURN",
}

type Token struct {
	Type    TokenType // 词元类型
	Literal string    // 字面量
}

const (
	ILLEGAL TokenType = iota // 未知的词法单元或字符
	EOF                      // 文件结尾

	// 标识符+字面量
	IDENT // add, foobar, x, y, ...
	INT   // 1343456
	FLOAT
	STRING

	// 运算符
	ASSIGN
	PLUS
	MINUS
	BANG
	ASTERISK
	SLASH
	LT
	LTE
	GT
	GTE
	EQ
	NOT_EQ
	DOT

	// 分隔符
	COMMA
	SEMICOLON
	COLON
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET

	// 关键字
	FUNCTION
	LET
	TRUE
	FALSE
	IF
	ELSE
	RETURN
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func DetermineNumberType(number string) TokenType {
	if strings.Contains(number, ".") {
		return FLOAT
	}
	return INT
}
