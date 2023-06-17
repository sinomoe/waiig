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
		return &object.Integer{Value: v.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(v.Value)
	case *ast.Program:
		return evalProgram(v.Statements)
	case *ast.ExpressionStatement:
		return Eval(v.Expression)
	case *ast.BlockStatement:
		return evalBlockStatement(v.Statements)
	case *ast.ReturnStatement:
		return &object.ReturnValue{Value: Eval(v.ReturnValue)}
	case *ast.PrefixExpression:
		return evalPrefixExpression(v.Operator, Eval(v.Right))
	case *ast.InfixExpression:
		return evalInfixExpression(v.Operator, Eval(v.Left), Eval(v.Right))
	case *ast.IfExpression:
		return evalIfExpression(Eval(v.Condition), v.Consequence, v.Alternative)
	default:
		return NULL
	}
}

// evalBlockStatement 对多条语句求值,多条语句的求值结果为最后一条语句的求值结果
func evalBlockStatement(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(statement)
		// 如果遇到了返回值 终止继续向下解析语句
		// 由于这里需要能识别出 返回值 故必须要多出一个返回值类型的 object
		if result.Type() == object.RETURN_VALUE_OBJ {
			// 这里不对返回值进行解包 是因为 block 是可能嵌套的 为了保证程序按照正确的顺序返回
			// 这里必须将返回值传递到上层
			// 递归结束后 在最上层的 block 即可正确感知到应该在第一个 ReturnValue 处返回
			return result
		}
	}
	return result
}

// evalProgram 对程序进行求值 并最后对返回值进行解包 遇到返回值则马上返回 不再向下解析
func evalProgram(stmts []ast.Statement) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(statement)
		// 如果遇到了返回值 终止继续向下解析语句
		// 由于这里需要能识别出 返回值 故必须要多出一个返回值类型的 object
		if retVal, ok := result.(*object.ReturnValue); ok {
			// 因为 program 本质就是最顶层的 block
			// 不存在更上层的 block 需要感知到 ReturnValue
			// 故在这里直接解包 ReturnValue 拿到解包的值
			return retVal.Value
		}
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

// evalInfixExpression 求值中缀表达式
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	default:
		return NULL
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	var (
		leftVal  = left.(*object.Integer).Value
		rightVal = right.(*object.Integer).Value
	)
	switch operator {
	case "+":
		return object.NewInteger(leftVal + rightVal)
	case "-":
		return object.NewInteger(leftVal - rightVal)
	case "*":
		return object.NewInteger(leftVal * rightVal)
	case "/":
		return object.NewInteger(leftVal / rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return NULL
	}
}

func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	var (
		leftVal  = left.(*object.Boolean).Value
		rightVal = right.(*object.Boolean).Value
	)
	switch operator {
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return NULL
	}
}

func evalIfExpression(condition object.Object, consequence, alternative *ast.BlockStatement) object.Object {
	if isTruthy(condition) {
		return Eval(consequence)
	}
	if alternative == nil {
		return NULL
	}
	return Eval(alternative)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	}
	if v, ok := obj.(*object.Integer); ok {
		return v.Value != 0
	}
	return false
}
