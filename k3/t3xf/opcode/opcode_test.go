package opcode_test

import (
	"testing"

	"github.com/nokia/ntt/k3/t3xf/opcode"
)

func TestUnpack(t *testing.T) {
	tests := []struct {
		input uint32
		op    opcode.Opcode
		arg   any
	}{
		{input: 0b00_0000000000000000000000000000_00, op: opcode.REF, arg: 0},
		{input: 0b00_0000000000000000000000000001_00, op: opcode.REF, arg: 4},
		{input: 0b00_0000001000000000000000000011_00, op: opcode.REF, arg: 8388620},
		{input: 0b00_0111111111111111111111111111_00, op: opcode.REF, arg: 536870908},
		{input: 0b10_0000000000000000000000000000_00, op: opcode.FROZEN_REF, arg: 0},
		{input: 0b10_0000000000000000000000000001_00, op: opcode.FROZEN_REF, arg: 4},
		{input: 0b10_0000001000000000000000000011_00, op: opcode.FROZEN_REF, arg: 8388620},
		{input: 0b10_0111111111111111111111111111_00, op: opcode.FROZEN_REF, arg: 536870908},
		{input: 0b00_0000000000000000000000000000_01, op: opcode.GOTO, arg: 0},
		{input: 0b00_0000000000000000000000000001_01, op: opcode.GOTO, arg: 4},
		{input: 0b00_0000001000000000000000000011_01, op: opcode.GOTO, arg: 8388620},
		{input: 0b00_0111111111111111111111111111_01, op: opcode.GOTO, arg: 536870908},
		{input: 0b01_0000001000000000000000000011_01, op: opcode.CALL, arg: 8388620},
		{input: 0b10_0000001000000000000000000011_01, op: opcode.ISCAN, arg: 8388620},
		{input: 0b00_0000000000000000000000000000_10, op: opcode.LINE, arg: 0},
		{input: 0b00_0000000000000000000000000001_10, op: opcode.LINE, arg: 1},
		{input: 0b00_0111111111111111111111111111_10, op: opcode.LINE, arg: 134217727},
	}

	for _, tt := range tests {
		op, arg := opcode.Unpack(tt.input)
		if op != tt.op || arg != tt.arg {
			t.Errorf("opcode.Unpack(%b) = %s,%v; want=%s,%v ", tt.input, op, arg, tt.op, tt.arg)
		}
	}
}
