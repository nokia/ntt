package runtime

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

type Unit int

func (u Unit) Base() int {
	switch u {
	case Bit:
		return 2
	case Hex, Octett:
		return 16
	default:
		return -1
	}
}

const (
	Bit    Unit = 1
	Hex         = 4
	Octett      = 8
)

type Bitstring struct {
	Value *big.Int
	Unit  Unit
}

func (b *Bitstring) Type() ObjectType { return BITSTRING }
func (b *Bitstring) Inspect() string {
	switch b.Unit {
	case Bit:
		return fmt.Sprintf("'%b'B", b.Value)
	case Octett:
		return fmt.Sprintf("'%h'O", b.Value)
	default:
		return fmt.Sprintf("'%h'H", b.Value)
	}
}

func NewBitstring(s string) (*Bitstring, error) {
	if len(s) < 3 || s[0] != '\'' || s[len(s)-2] != '\'' {
		return nil, ErrSyntax
	}

	var unit Unit
	switch strings.ToUpper(string(s[len(s)-1])) {
	case "B":
		unit = Bit
	case "H":
		unit = Hex
	case "O":
		unit = Octett
	default:
		return nil, ErrSyntax
	}

	removeWhitespaces := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	s = strings.Map(removeWhitespaces, s[1:len(s)-2])

	if i, ok := new(big.Int).SetString(s, unit.Base()); ok {
		return &Bitstring{Value: i, Unit: unit}, nil
	}

	// TODO(5nord) parse Bitstring templates (e.g. '01*1'B)
	return nil, ErrSyntax
}
