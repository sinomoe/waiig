package object

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func NewEnclosedEnviroment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]Object),
		outer: outer,
	}
}

type Environment struct {
	store map[string]Object // 当前作用域内的值
	outer *Environment      // 指向上层作用域
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, _, ok := e.getWithEnv(name)
	return obj, ok
}

// GetLocal 仅从当前作用域尝试获取变量
func (e *Environment) GetLocal(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

// getWithEnv 从作用域链中获取参数 并返回参数所在作用域
func (e *Environment) getWithEnv(name string) (Object, *Environment, bool) {
	obj, ok := e.store[name]
	if ok {
		return obj, e, true
	}
	if e.outer != nil {
		return e.outer.getWithEnv(name) // 递归向上层作用域寻找
	}
	return nil, nil, false
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

func (e *Environment) Assign(name string, val Object) Object {
	_, env, ok := e.getWithEnv(name)
	if !ok {
		return &Error{Message: "illegal assign, symbol not exist: " + name}
	}
	env.Set(name, val)
	return val
}
