package runtime

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING }
func (s *String) Inspect() string  { return s.Value }
