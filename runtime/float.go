package runtime

import "math/big"

type Float struct{ *big.Float }

func (f Float) Type() ObjectType { return FLOAT }
func (f Float) Inspect() string  { return f.String() }

func NewFloat(s string) Float {
	f, _ := new(big.Float).SetString(s)
	return Float{Float: f}
}
