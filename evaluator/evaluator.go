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
	case *ast.PrefixExpression:
		return evalPrefixExpression(v.Operator, Eval(v.Right))
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

// evalPrefixExpression 前缀表达式求值
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	}
	return nil
}

// evalBangOperatorExpression 求反表达式
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE, NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression 负数表达式
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch v := right.(type) {
	case *object.Integer:
		return &object.Integer{
			Value: -v.Value,
		}
	default:
		return NULL
	}
}
