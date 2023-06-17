package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

// 以下对象全局都是一致的，无需在每次使用时都重复创建
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch v := node.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{
			Value: v.Value,
		}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(v.Value)
	case *ast.Program:
		return evalStatements(v.Statements)
	case *ast.ExpressionStatement:
		return Eval(v.Expression)
	}
	return nil
}

func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(statement)
	}
	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}
