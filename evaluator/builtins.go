package evaluator

import "monkey/object"

var builtins = map[string]object.BuiltinFunction{
	"len": func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}
		switch val := args[0].(type) {
		case *object.String:
			return object.NewInteger(int64(len(val.Value)))
		}
		return newError("argument to `len` not supported, got %s", args[0].Type())
	},
}
