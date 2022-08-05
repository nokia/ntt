package runtime

var builtins = map[string]*Builtin{}

func AddBuiltin(name string, f func(args ...Object) Object) {
	builtins[name] = &Builtin{Fn: f}
}
