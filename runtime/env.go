package runtime

type Scope interface {
	Get(name string) (Object, bool)
	Set(name string, val Object) Object
}

type Env struct {
	outer Scope
	store map[string]Object
}

func (env *Env) Get(name string) (Object, bool) {
	if builtin, ok := builtins[name]; ok {
		return builtin, true
	}
	if val, ok := env.store[name]; ok {
		return val, true
	}
	if env.outer != nil {
		return env.outer.Get(name)
	}
	return nil, false
}

func (env *Env) Set(name string, val Object) Object {
	env.store[name] = val
	return val
}

func NewEnv(outer Scope) *Env {
	return &Env{
		outer: outer,
		store: make(map[string]Object),
	}
}
