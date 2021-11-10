package runtime

import "fmt"

type Bool bool

func NewBool(b bool) Bool {
	return Bool(b)
}

func (b Bool) Type() ObjectType { return BOOL }
func (b Bool) Inspect() string  { return fmt.Sprintf("%t", b) }

func (b Bool) Bool() bool { return bool(b) }
