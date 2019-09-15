package t3xf

import (
	"testing"

	"github.com/nokia/ntt/internal/t3xf/opcode"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	// Test if an empty buffer and multi Scan calls don't cause any trouble.
	t.Run("buffer/empty", func(t *testing.T) {
		s := NewScanner([]byte{})
		assert.Nil(t, s.Err())
		assert.Equal(t, false, s.Scan())
		assert.Equal(t, false, s.Scan())
		assert.Equal(t, false, s.Scan())
		assert.Nil(t, s.Err())
	})

	// Test if NOPs are decoded correctly
	t.Run("NOP", func(t *testing.T) {
		s := NewScanner([]byte{
			0x03, 0x00, 0x00, 0x00, // NOP 0x0000
			0x03, 0x00, 0x34, 0x12, // NOP 0x1234
			0x0f, 0x00, 0x78, 0x56, // NOP 0x5678
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.NOP, s.Opcode())
		assert.Equal(t, 0, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.NOP, s.Opcode())
		assert.Equal(t, 0x1234, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.NOP, s.Opcode())
		assert.Equal(t, 0x5678, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, false, s.Scan())
		assert.Equal(t, opcode.NOP, s.Opcode())
		assert.Equal(t, 0x5678, s.Arg())
		assert.Nil(t, s.Err())
	})

	// Test line opcodes
	t.Run("LINE", func(t *testing.T) {
		s := NewScanner([]byte{
			0xfe, 0xff, 0xff, 0xff, // LINE 1073741823
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.LINE), s.Opcode())
		assert.Equal(t, 1073741823, s.Arg())
		assert.Nil(t, s.Err())
	})

	// Test goto opcodes
	t.Run("GOTO", func(t *testing.T) {
		s := NewScanner([]byte{
			0xfd, 0xff, 0xff, 0xff, // REF 0xfffffffc
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.GOTO), s.Opcode())
		assert.Equal(t, 0xfffffffc, s.Arg())
		assert.Nil(t, s.Err())
	})

	// Test reference opcodes
	t.Run("REF", func(t *testing.T) {
		s := NewScanner([]byte{
			0xfc, 0xff, 0xff, 0x7f, // REF 0x7ffffffc
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.REF), s.Opcode())
		assert.Equal(t, 0x7ffffffc, s.Arg())
		assert.Nil(t, s.Err())

	})

	// Test frozen reference opcodes
	t.Run("FROZEN_REF", func(t *testing.T) {
		s := NewScanner([]byte{
			0xfc, 0xff, 0xff, 0xff, // FROZEN_REF 0x7ffffffc
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.FROZEN_REF), s.Opcode())
		assert.Equal(t, 0x7ffffffc, s.Arg())
		assert.Nil(t, s.Err())
	})

	// Test empty hexstring literal + repetitive calls
	t.Run("NIBBLES", func(t *testing.T) {
		s := NewScanner([]byte{
			0x23, 0x01, 0x00, 0x00, // NIBBLES
			0x00, 0x00, 0x00, 0x00, // length 0
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, int(opcode.NIBBLES), int(s.Opcode()))
		assert.Equal(t, 0, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, false, s.Scan())
	})

	// Test odd hexstring literal + repetitive calls
	t.Run("NIBBLES", func(t *testing.T) {
		s := NewScanner([]byte{
			0x23, 0x01, 0x00, 0x00, // NIBBLES
			0x03, 0x00, 0x00, 0x00, // length 3
			0xab, 0xc0, 0x00, 0x00, // 0xabc + padding
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, int(opcode.NIBBLES), int(s.Opcode()))
		assert.Equal(t, 3*4, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, false, s.Scan())
	})

	// Test invalid NATLONG
	t.Run("NATLONG", func(t *testing.T) {
		s := NewScanner([]byte{
			0x43, 0x01, 0x00, 0x00, // NATLONG
		})

		assert.NotNil(t, s.Err())
	})

	// Test negative NATLONG
	t.Run("NATLONG", func(t *testing.T) {
		s := NewScanner([]byte{
			0x43, 0x01, 0x00, 0x00, // NATLONG
			0xfe, 0xff, 0xff, 0xff, // -2
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.NATLONG), s.Opcode())
		assert.Equal(t, -2, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, false, s.Scan())
	})

	// Test float
	t.Run("IEEE754DP", func(t *testing.T) {
		s := NewScanner([]byte{
			0x53, 0x01, 0x00, 0x00, // IEEE754DP
			0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x37, 0x40, // 23.5
		})

		assert.Nil(t, s.Err())
		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.IEEE754DP), s.Opcode())
		assert.Equal(t, 23.5, s.Float64())
		assert.Equal(t, false, s.Scan())
		assert.Nil(t, s.Err())
	})

	// Test block hierarchy
	t.Run("SCAN/BLOCK", func(t *testing.T) {
		s := NewScanner([]byte{
			0xd3, 0x00, 0x00, 0x00, // 0: SCAN
			0xd3, 0x00, 0x00, 0x00, // 1:   SCAN
			0x83, 0x00, 0x00, 0x00, // 2:   BLOCK
			0xd3, 0x00, 0x00, 0x00, // 3:   SCAN
			0xd3, 0x00, 0x00, 0x00, // 4:     SCAN
			0x83, 0x00, 0x00, 0x00, // 5:     BLOCK
			0x83, 0x00, 0x00, 0x00, // 6:   BLOCK
			0x83, 0x00, 0x00, 0x00, // 7: BLOCK
			0xd3, 0x00, 0x00, 0x00, // 8: SCAN
			0x83, 0x00, 0x00, 0x00, // 9: BLOCK
		})

		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.SCAN), s.Opcode())
		assert.Equal(t, 7, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.SCAN), s.Opcode())
		assert.Equal(t, 2, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.BLOCK), s.Opcode())
		assert.Equal(t, 1, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.SCAN), s.Opcode())
		assert.Equal(t, 6, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.SCAN), s.Opcode())
		assert.Equal(t, 5, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.BLOCK), s.Opcode())
		assert.Equal(t, 4, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.BLOCK), s.Opcode())
		assert.Equal(t, 3, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.BLOCK), s.Opcode())
		assert.Equal(t, 0, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.SCAN), s.Opcode())
		assert.Equal(t, 9, s.Arg())
		assert.Nil(t, s.Err())

		assert.Equal(t, true, s.Scan())
		assert.Equal(t, opcode.Opcode(opcode.BLOCK), s.Opcode())
		assert.Equal(t, 8, s.Arg())
		assert.Nil(t, s.Err())
	})

	// Test missing scan
	t.Run("SCAN", func(t *testing.T) {
		s := NewScanner([]byte{
			0x83, 0x00, 0x00, 0x00, // BLOCK
		})
		assert.Equal(t, ErrUnmatchedBLOCK, s.Err())
	})

	// Test missing block
	t.Run("BLOCK", func(t *testing.T) {
		s := NewScanner([]byte{
			0xd3, 0x00, 0x00, 0x00, // SCAN
		})
		assert.Equal(t, ErrUnmatchedSCAN, s.Err())
	})
}
