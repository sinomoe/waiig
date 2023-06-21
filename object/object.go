package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	BULTIN_OBJ       = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
)

// Object 用来表示解释器中的值
type Object interface {
	Type() ObjectType
	Inspect() string // 用于 REPL 返回展示
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

func NewInteger(val int64) *Integer {
	return &Integer{
		Value: val,
	}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_OBJ
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_OBJ
}

func (s *String) Inspect() string {
	return fmt.Sprintf("%s", s.Value)
}

type Null struct{}

func (n *Null) Type() ObjectType {
	return NULL_OBJ
}

func (n *Null) Inspect() string {
	return "null"
}

// ReturnValue 是对返回值的包装 用于标识返回值
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_VALUE_OBJ
}

func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}

// Error 表示求值错误
type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR_OBJ
}

func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

// Function 函数的值表示 一等公民
type Function struct {
	Parameters []*ast.Identifier   //继承自 AST 节点
	Body       *ast.BlockStatement // 继承自 AST 节点
	Env        *Environment        // 函数内部变量 可以实现闭包
}

func (f *Function) Type() ObjectType {
	return FUNCTION_OBJ
}

func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type BuiltinFunction func(args ...Object) Object

func (bf BuiltinFunction) Type() ObjectType {
	return BULTIN_OBJ
}

func (bf BuiltinFunction) Inspect() string {
	return "builtin function"
}

type Array []Object

func (a Array) Type() ObjectType {
	return ARRAY_OBJ
}

func (a Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range a {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
