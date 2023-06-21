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
		case object.Array:
			return object.NewInteger(int64(len(val)))
		}
		return newError("argument to `len` not supported, got %s", args[0].Type())
	},
	"first": func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		if args[0].Type() != object.ARRAY_OBJ {
			return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
		}
		arr := args[0].(object.Array)
		if len(arr) > 0 {
			return arr[0]
		}
		return NULL
	},
	"last": func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		if args[0].Type() != object.ARRAY_OBJ {
			return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
		}
		arr := args[0].(object.Array)
		if len(arr) > 0 {
			return arr[len(arr)-1]
		}
		return NULL
	},
	"rest": func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}
		if args[0].Type() != object.ARRAY_OBJ {
			return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
		}
		arr := args[0].(object.Array)
		length := len(arr)
		if length > 0 {
			newElements := make([]object.Object, length-1)
			copy(newElements, arr[1:length])
			return object.Array(newElements)
		}
		return NULL
	},
	"push": func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		if args[0].Type() != object.ARRAY_OBJ {
			return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
		}
		arr := args[0].(object.Array)
		length := len(arr)
		newElements := make([]object.Object, length+1)
		copy(newElements, arr)
		newElements[length] = args[1]
		return object.Array(newElements)
	},
}
