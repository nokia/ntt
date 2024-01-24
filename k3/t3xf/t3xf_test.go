package t3xf_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/stretchr/testify/assert"
)

func TestEncoder(t *testing.T) {
	e := t3xf.NewEncoder()

	t.Run("empty", func(t *testing.T) {
		e.Reset()
		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("resolving", func(t *testing.T) {
		e.Reset()

		// Unknown references are address 0xfffffffc
		assert.NoError(t, e.Encode(opcode.REF, 1))
		b, err := e.Assemble()
		assert.ErrorIs(t, err, t3xf.ErrUnknownReference)
		assert.Equal(t, []byte{0xfc, 0xff, 0xff, 0xff}, b)

		// Known references are resolved
		assert.NoError(t, e.Encode(opcode.REF, 0))
		b, err = e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x04, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		}, b)
	})

	t.Run("self-reference", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.REF, 0))
		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, b)
	})

	t.Run("references", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.REF, 0))
		assert.NoError(t, e.Encode(opcode.FROZEN_REF, 0))
		assert.NoError(t, e.Encode(opcode.GOTO, 0))

		// Note: SCAN cannot use 0 for reference, because of the
		//       opcode.Pack function. This is not a problem, because
		//       SCAN references have to be greater than 0.
		assert.NoError(t, e.Encode(opcode.SCAN, 1))

		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x80,
			0x01, 0x00, 0x00, 0x00,
			0x05, 0x00, 0x00, 0x80,
		}, b)
	})

	t.Run("SCAN", func(t *testing.T) {
		e.Reset()

		// SCAN without argument is a regular SCAN instruction
		assert.NoError(t, e.Encode(opcode.SCAN, nil))

		b, err := e.Assemble()
		assert.NoError(t, err, t3xf.ErrUnknownReference)

		assert.Equal(t, []byte{
			0xd3, 0x00, 0x00, 0x00,
		}, b)
	})

	t.Run("LINE", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.LINE, 1))
		assert.NoError(t, e.Encode(opcode.LINE, 0))
		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x06, 0x00, 0x00, 0x0,
			0x02, 0x00, 0x00, 0x0,
		}, b)
	})

	t.Run("NATLONG", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.NATLONG, 42))
		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x43, 0x01, 0x00, 0x00,
			0x2a, 0x00, 0x00, 0x00,
		}, b)
		e.Reset()
	})

	t.Run("padding", func(t *testing.T) {
		e.Reset()

		assert.NoError(t, e.Encode(opcode.NAME, ""))
		assert.NoError(t, e.Encode(opcode.NAME, "a"))
		assert.NoError(t, e.Encode(opcode.NAME, "ab"))
		assert.NoError(t, e.Encode(opcode.NAME, "abc"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcd"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcde"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcdef"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcdefg"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcdefgh"))
		assert.NoError(t, e.Encode(opcode.NAME, "abcdefghi"))

		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x83, 0x01, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x61, 0x00, 0x00, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x02, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x00, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x03, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x04, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,

			0x83, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,
			0x65, 0x00, 0x00, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x06, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,
			0x65, 0x66, 0x00, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x07, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,
			0x65, 0x66, 0x67, 0x00,

			0x83, 0x01, 0x00, 0x00,
			0x08, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,
			0x65, 0x66, 0x67, 0x68,

			0x83, 0x01, 0x00, 0x00,
			0x09, 0x00, 0x00, 0x00,
			0x61, 0x62, 0x63, 0x64,
			0x65, 0x66, 0x67, 0x68,
			0x69, 0x00, 0x00, 0x00,
		}, b)

	})

	t.Run("BITS", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.BITS, t3xf.NewString(0, nil)))
		assert.NoError(t, e.Encode(opcode.BITS, t3xf.NewString(1, []byte{0x01})))
		assert.NoError(t, e.Encode(opcode.BITS, t3xf.NewString(9, []byte{0xff, 0x01})))

		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x13, 0x01, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,

			0x13, 0x01, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,

			0x13, 0x01, 0x00, 0x00,
			0x09, 0x00, 0x00, 0x00,
			0xff, 0x01, 0x00, 0x00,
		}, b)
	})

	t.Run("NIBBLES", func(t *testing.T) {
		e.Reset()
		assert.NoError(t, e.Encode(opcode.NIBBLES, t3xf.NewString(3, []byte{0x21, 0x03})))

		b, err := e.Assemble()
		assert.NoError(t, err)
		assert.Equal(t, []byte{
			0x23, 0x01, 0x00, 0x00,
			0x03, 0x00, 0x00, 0x00,
			0x21, 0x03, 0x00, 0x00,
		}, b)
	})
}

func TestDecode(t *testing.T) {
	input := []struct {
		name  string
		input []byte
		op    opcode.Opcode
		arg   string
		n     int
	}{
		{name: "REF", input: []byte{0x00, 0x00, 0x00, 0x00},
			n: 4, op: opcode.REF, arg: "t3xf.Reference:0"},

		{name: "FROZEN_REF", input: []byte{0x00, 0x00, 0x00, 0x80},
			n: 4, op: opcode.FROZEN_REF, arg: "t3xf.Reference:0"},

		{name: "LINE", input: []byte{0x02, 0x00, 0x00, 0x00},
			n: 4, op: opcode.LINE, arg: "int:0"},

		{name: "GOTO", input: []byte{0x01, 0x00, 0x00, 0x00},
			n: 4, op: opcode.GOTO, arg: "t3xf.Reference:0"},

		{name: "SCAN", input: []byte{0x01, 0x00, 0x00, 0x80},
			n: 4, op: opcode.SCAN, arg: "t3xf.Reference:0"},

		{name: "SCAN", input: []byte{0xd3, 0x00, 0x00, 0x00},
			n: 4, op: opcode.SCAN, arg: "<nil>"},

		{name: "SCAN", input: []byte{0xd3, 0x00, 0x04, 0x00},
			n: 4, op: opcode.SCAN, arg: "t3xf.Reference:4"},

		{name: "IFIELD", input: []byte{0xd3, 0x05, 0x06, 0x00},
			n: 4, op: opcode.IFIELD, arg: "int:6"},

		{name: "IDEF", input: []byte{0xc3, 0x05, 0x06, 0x00},
			n: 4, op: opcode.IDEF, arg: "int:6"},

		{name: "IGET", input: []byte{0xa3, 0x05, 0x06, 0x00},
			n: 4, op: opcode.IGET, arg: "int:6"},

		{name: "NOP", input: []byte{0x03, 0x00, 0x06, 0x00},
			n: 4, op: opcode.NOP, arg: "int:6"},

		{name: "NATLONG", n: 8, op: opcode.NATLONG, arg: "int:9",
			input: []byte{
				0x43, 0x01, 0x00, 0x00,
				0x09, 0x00, 0x00, 0x00,
			}},

		{name: "BITS", n: 12, op: opcode.BITS, arg: "*t3xf.String:&{l:4 b:[1]}",
			input: []byte{
				0x13, 0x01, 0x00, 0x00,
				0x04, 0x00, 0x00, 0x00,
				0x01, 0x02, 0x03, 0x04,
			}},
	}
	for _, tt := range input {
		t.Run(tt.name, func(t *testing.T) {
			b := []byte{
				0xaa, 0xaa, 0xaa, 0xaa,
				0xaa, 0xaa, 0xaa, 0xaa,
				0xaa, 0xaa, 0xaa, 0xaa,
				0xaa, 0xaa, 0xaa, 0xaa,
				0xaa, 0xaa, 0xaa, 0xaa,
				0xaa, 0xaa, 0xaa, 0xaa,
			}
			copy(b, tt.input)

			n, op, arg := t3xf.Decode(b)
			if op != tt.op {
				t.Errorf("Wrong opcode: want=%s, got=%s", tt.op.String(), op.String())
			}
			if n != tt.n {
				t.Errorf("Wrong number of decoded bytes: want=%d, got=%d", tt.n, n)
			}
			if got := strings.TrimPrefix(fmt.Sprintf("%T:%+v", arg, arg), "<nil>:"); got != tt.arg {
				t.Errorf("Wrong argument: want=%s, got=%s", tt.arg, got)
			}
		})
	}
}

type Instruction struct {
	Op  opcode.Opcode
	Arg interface{}
}
