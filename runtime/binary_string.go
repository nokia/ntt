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

type BinaryString struct {
	Value *big.Int
	Unit  Unit
}

func (b *BinaryString) Type() ObjectType { return BINARY_STRING }
func (b *BinaryString) Inspect() string {
	switch b.Unit {
	case Bit:
		return fmt.Sprintf("'%b'B", b.Value)
	case Octett:
		return fmt.Sprintf("'%h'O", b.Value)
	default:
		return fmt.Sprintf("'%h'H", b.Value)
	}
}

func NewBinaryString(s string) (*BinaryString, error) {
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
		return &BinaryString{Value: i, Unit: unit}, nil
	}

	// TODO(5nord) parse BinaryString templates (e.g. '01*1'B)
	return nil, ErrSyntax
}
