package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

// 以下对象全局都是一致的，无需在每次使用时都重复创建
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch v := node.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: v.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(v.Value)
	case *ast.StringLiteral:
		return &object.String{Value: v.Value}
	case *ast.Identifier:
		return evalIdentifier(v, env)
	case *ast.FunctionLiteral:
		return &object.Function{
			Parameters: v.Parameters,
			Body:       v.Body,
			Env:        env, // 解析函数字面量时保存申明的上下文 相当于创建了闭包
		}
	case *ast.Program:
		return evalProgram(v.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(v.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(v.Statements, env)
	case *ast.ReturnStatement:
		val := Eval(v.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		_, ok := env.GetLocal(v.Name.Value)
		if ok {
			return newError("identifier exist: " + v.Name.Value)
		}
		val := Eval(v.Value, env)
		if isError(val) {
			return val
		}
		env.Set(v.Name.Value, val)
	case *ast.AssignStatement:
		val := Eval(v.Value, env)
		if isError(val) {
			return val
		}
		env.Assign(v.Name.Value, val)
		return val
	case *ast.PrefixExpression:
		val := Eval(v.Right, env)
		if isError(val) {
			return val
		}
		return evalPrefixExpression(v.Operator, val)
	case *ast.InfixExpression:
		lVal := Eval(v.Left, env)
		if isError(lVal) {
			return lVal
		}
		rVal := Eval(v.Right, env)
		if isError(rVal) {
			return rVal
		}
		return evalInfixExpression(v.Operator, lVal, rVal)
	case *ast.IfExpression:
		val := Eval(v.Condition, env)
		if isError(val) {
			return val
		}
		return evalIfExpression(val, v.Consequence, v.Alternative, env)
	case *ast.CallExpression:
		val := Eval(v.Function, env) // val is function object
		if isError(val) {
			return val
		}
		args := evalExpressions(v.Arguments, env) // 首先对实参表达式求值
		if len(args) != 0 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(val, args)
	}
	return NULL
}

// evalBlockStatement 对多条语句求值,多条语句的求值结果为最后一条语句的求值结果
func evalBlockStatement(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(statement, env)
		// 求值的到解析错误则立刻返回
		if result.Type() == object.ERROR_OBJ {
			return result
		}
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
func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range stmts {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			// 如果遇到了返回值 终止继续向下解析语句
			// 由于这里需要能识别出 返回值 故必须要多出一个返回值类型的 object
			// 因为 program 本质就是最顶层的 block
			// 不存在更上层的 block 需要感知到 ReturnValue
			// 故在这里直接解包 ReturnValue 拿到解包的值
			return result.Value
		case *object.Error:
			// 求值的到解析错误则立刻返回
			return result
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
	}
	// 非 0 整数也视作 true
	if val, ok := right.(*object.Integer); ok {
		return nativeBoolToBooleanObject(val.Value == 0)
	}
	return newError("unknown operator: !%s", right.Type())
}

// evalMinusPrefixOperatorExpression 负数表达式
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch v := right.(type) {
	case *object.Integer:
		return &object.Integer{
			Value: -v.Value,
		}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

// evalInfixExpression 求值中缀表达式
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	var (
		leftVal  = left.(*object.String).Value
		rightVal = right.(*object.String).Value
	)
	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIfExpression(condition object.Object, consequence, alternative *ast.BlockStatement, env *object.Environment) object.Object {
	if isTruthy(condition) {
		return Eval(consequence, env)
	}
	if alternative == nil {
		return NULL
	}
	return Eval(alternative, env)
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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError 检查求值是否出错
// 所有调用 Eval 后都需要立刻检查 isError 并尽快返回 防止错误到处传播
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

// evalIdentifier　对标识符求值
// 实现方式是从作用域中寻找标识符的值 不存在值则抛出求值错误
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if ok {
		return val
	}
	builtin, ok := builtins[node.Value]
	if ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

// evalExpressions 对多条表达式求值
// 求值完成后返回对应顺序的值列表
// 求值过程一旦发生错误则只会返回错误
func evalExpressions(exprs []ast.Expression, env *object.Environment) []object.Object {
	var res []object.Object
	for _, expr := range exprs {
		val := Eval(expr, env)
		if isError(val) {
			return []object.Object{val}
		}
		res = append(res, val)
	}
	return res
}

// applyFunction 对函数调用求值
// 实现方法是首先对实参列表求值
// 然后创建新的包裹作用域 上层作用域指向函数申明时的作用域
// 接着将实参绑定到新的作用域中
// 最后使用新的作用域对函数的 body(block statement) 求值
func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch f := fn.(type) {
	case object.BuiltinFunction:
		return f(args...)
	case *object.Function:
		if len(args) != len(f.Parameters) {
			return newError("args number mismatch, expect lenght: %d, but got: %d", len(f.Parameters), len(args))
		}
		env := object.NewEnclosedEnviroment(f.Env)
		for i, param := range f.Parameters {
			env.Set(param.String(), args[i])
		}
		val := Eval(f.Body, env)
		// 重要：函数调用后应该返回一个解包后的值
		// 这里不进行解包会导致这个 ReturnValue 向上冒泡
		// 从而导致上层调用异常提前返回
		//
		// (因为设计 ReturnValue 这个类型的初衷是为了感知到多条 statment 执行时该何时返回
		// 以及避免 return 语句下的语句被执行 所以我们在 BlockStatement 和 Program 等
		// 涉及多条 statement 执行的地方都对 ReturnValue 进行了判断，并将 ReturnValue
		// 类型保留继续上抛)
		//
		// 直到遇到函数调用的边界就将 ReturnValue 解包得到内层值
		// 这是因为函数内的 return 不应该直接导致更上层函数的退出
		// 求值最顶层的语句 Program.Statements 时也解包 ReturnValue 的原因是
		// 整个程序的返回值应该是一个具体类型的值 而不是包装后的返回值
		// 假如不考虑程序的返回值 那么对 Program.Statements 求值时不解包 ReturnValue
		// 也是可以的 具体操作需要看对语言的行为怎么进行定义
		if retVal, ok := val.(*object.ReturnValue); ok {
			return retVal.Value
		}
		return val
	}
	return newError("not a function: %s", fn.Type())
}
