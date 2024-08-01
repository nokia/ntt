// Package opcode defines the opcodes used in T3XF.
package opcode

//go:generate go run ./internal/gen/main.go

import "fmt"

type Opcode int

func (op Opcode) String() string {
	if s, ok := opcodeStrings[op]; ok {
		return s
	}
	return fmt.Sprintf("unknown_opcode(0x%x)", int(op))
}

func Unpack(x uint32) (Opcode, int) {
	cls := x & 0b11
	if cls == 0b11 {
		return Opcode(x & 0xffff), int(x >> 16)
	}

	op := Opcode((x >> 28 & 0b1100) | cls)
	arg := int(x & 0b00_1111111111111111111111111111_00)

	if op == LINE {
		arg >>= 2
	}

	return op, arg
}

func Pack(op Opcode, arg int) uint32 {
	if op&instrMask == instrMask {
		return uint32(arg)<<16 | (uint32(op) & 0xffff)
	}

	hi := uint32(op) & 0b1100
	lo := uint32(op) & 0b0011

	if op == LINE {
		arg <<= 2
	}

	return hi<<28 | lo | uint32(arg)&0b00_1111111111111111111111111111_00
}

func Parse(s string) (Opcode, error) {
	if op, ok := opcodeNames[s]; ok {
		return op, nil
	}
	return -1, fmt.Errorf("unknown opcode %q", s)
}
