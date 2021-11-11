package runtime

type Env struct {
	store map[string]Object
}

func (env *Env) Get(name string) (Object, bool) {
	val, ok := env.store[name]
	return val, ok
}

func (env *Env) Set(name string, val Object) Object {
	env.store[name] = val
	return val
}

func NewEnv() *Env {
	return &Env{
		store: make(map[string]Object),
	}
}
