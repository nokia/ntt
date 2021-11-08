package runtime

import "math/big"

type Int struct{ *big.Int }

func (i Int) Inspect() string { return i.String() }

func NewInt(s string) Int {
	return Int{parseInt(s, 10)}
}

func parseInt(s string, base int) *big.Int {
	i := &big.Int{}
	i.SetString(s, base)
	return i
}
