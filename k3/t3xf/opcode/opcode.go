// Package opcode defines the opcodes used in T3XF.
package opcode

//go:generate go run ./internal/gen/main.go

import "fmt"

const (
	refClass   = 0
	gotoClass  = 1
	lineClass  = 2
	instrClass = 3
)

type Opcode int

func (op Opcode) String() string {
	if s, ok := opcodeStrings[op]; ok {
		return s
	}
	return fmt.Sprintf("unknown_opcode(0x%x)", int(op))
}

func Unpack(x uint32) (Opcode, int) {
	i := int(x)
	switch i & 0x3 {
	case refClass:
		if uint32(i)&(1<<31) != 0 {
			return FROZEN_REF, int(x &^ (1 << 31))
		}
		return REF, i
	case lineClass:
		return LINE, i >> 2
	case gotoClass:
		switch (uint32(i) & (3 << 30)) >> 30 {
		case 1:
			return APPLY, int(uint32(i) & 0x3ffffffc)
		case 2:
			return SCAN, int(uint32(i) & 0x3ffffffc)
		}
		return GOTO, i & ^(0x3)
	default:
		return Opcode((i & 0xffff)), i >> 16
	}
}

func Pack(op Opcode, x int) uint32 {
	switch op {
	case REF:
		return uint32(x)
	case FROZEN_REF:
		return uint32(x) | (1 << 31)
	case LINE:
		return uint32(x<<2) | lineClass
	case GOTO:
		return uint32(x) | gotoClass
	default:
		if x != 0 {
			switch op {
			case APPLY:
				return uint32(x) | gotoClass | (1 << 30)
			case SCAN:
				return uint32(x) | gotoClass | (2 << 30)
			}
		}
		return uint32(x)<<16 | (uint32(op) & 0xffff)
	}
}

func Parse(s string) (Opcode, error) {
	if op, ok := opcodeNames[s]; ok {
		return op, nil
	}
	return -1, fmt.Errorf("unknown opcode %q", s)
}
