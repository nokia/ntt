package runtime

type Env struct {
	outer *Env
	store map[string]Object
}

func (env *Env) Get(name string) (Object, bool) {
	val, ok := env.store[name]
	if !ok && env.outer != nil {
		val, ok = env.outer.Get(name)
	}
	return val, ok
}

func (env *Env) Set(name string, val Object) Object {
	env.store[name] = val
	return val
}

func NewEnv(outer *Env) *Env {
	return &Env{
		outer: outer,
		store: make(map[string]Object),
	}
}
