package runtime

type ReturnValue struct {
	Value Object
}

func (r *ReturnValue) Inspect() string { return r.Value.Inspect() }
