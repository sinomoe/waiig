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
	obj, ok := e.store[name]
	if ok {
		return obj, true
	}
	if e.outer != nil {
		return e.outer.Get(name) // 递归向上层作用域寻找
	}
	return nil, false
}
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
