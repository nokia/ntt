package t3xf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/nokia/ntt/t3xf/opcode"
)

type instr struct {
	op  opcode.Opcode // T3XF Opcode
	arg int           // Optional instruction argument
	n   int           // Length of instruction in bytes
}

// Scanner provides a convenient interface for decoding T3XF byte code
// instructions. Successive calls to the Scan method will step through T3XF
// instructions.
type Scanner struct {
	b    []byte  // The t3xf buffer provided by the client.
	code []instr // Decoded instructions for Scan.

	pos     int   // Current Scan position
	nextPos int   // Position of next Scan
	done    bool  // Scan finished.
	err     error // Sticky error
}

// NewScanner decodes a byte slice containing t3xf byte code and returns a
// Scanner for convenient traversal. The underlaying array of b is not modified
// but referenced by Scanner methods.
//
// Decoding errors can be inspected using the Err method.
func NewScanner(b []byte) *Scanner {
	s := Scanner{
		b:    b,
		code: make([]instr, len(b)/4),
	}
	s.decode(b)
	s.nextPos = 0

	return &s
}

// Scan advances the Scanner to the next instruction, which will then be
// available through methods like Opcode, Arg, ... .
//
// It returns false when the scan stops, either by reaching the end of the input
// or an error. After Scan returns false, the Err method will return any error
// that occurred during scanning, except that if it was io.EOF, Err will return
// nil.
func (s *Scanner) Scan() bool {
	if s.done {
		return false
	}

	s.pos = s.nextPos
	if s.pos >= len(s.code) {
		s.pos = len(s.code) - 1
		s.done = true
		return false
	}
	s.nextPos = s.pos + (s.code[s.pos].n / 4)
	return true
}

// Seek implements the io.Seeker interface.
func (s *Scanner) Seek(offset int64, whence int) (int64, error) {
	// 4 byte align offset
	offset /= 4

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(s.pos) + offset
	case io.SeekEnd:
		abs = int64(len(s.code)) + offset
	default:
		return 0, errors.New("t3xf.Scanner.Seek: invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("t3xf.Scanner.Seek: negative position")
	}
	s.pos = int(abs)
	return abs * 4, nil

}

// Reset sets the scanner to the first t3xf instruction and allows re-scanning
// t3xf code.
func (s *Scanner) Reset() {
	if s.err == nil {
		s.nextPos = 0
		s.done = false
	}
}

// Offset returns the offset of the current instruction.
func (s *Scanner) Offset() int {
	return s.pos * 4
}

// Raw returns a byte slice of the most recent instruction. The underlying array
// points to data from NewScanner call.
func (s *Scanner) Raw() []byte {
	i := s.pos * 4
	return s.b[i : i+s.code[s.pos].n]
}

// Opcode returns the opcode of the most recent instruction.
func (s *Scanner) Opcode() opcode.Opcode {
	return s.code[s.pos].op
}

// Arg returns the argument of the most recent instruction.
//
// For NATLONG it's the integer value, for strings (NAME, UTF8, ...) it's the
// length in bits. For SCAN instructions it's the matching BLOCK instruction and
// vice versa.
func (s *Scanner) Arg() int {
	return int(s.code[s.pos].arg)
}

// Bytes returns a byte slice containing string data of the most recent
// instruction. The underlying array points to data from NewScanner call.
func (s *Scanner) Bytes() []byte {
	pos := (s.pos + 2) * 4
	nBits := s.Arg()
	nBytes := (nBits + 7) / 8
	return s.b[pos : pos+nBytes]
}

// Float64 returns the float64 value of the most-recent instruction.
func (s *Scanner) Float64() float64 {
	i := s.pos*4 + 4
	r := bytes.NewReader(s.b[i : i+8])
	var u uint64
	if err := binary.Read(r, binary.LittleEndian, &u); err != nil {
		return math.NaN()
	}
	return math.Float64frombits(u)
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) decode(b []byte) {

	r := bytes.NewReader(b)

	stack := make([]int, 0, 20)
	for {
		// Fetch and decode instruction
		var u uint32
		if err := binary.Read(r, binary.LittleEndian, &u); err != nil {

			if err != io.EOF {
				s.recordError(err)
			}

			if len(stack) > 0 {
				s.recordError(ErrUnmatchedSCAN)
			}

			return
		}

		op, arg := opcode.Unpack(u)
		s.code[s.pos].op = op
		s.code[s.pos].arg = arg
		s.code[s.pos].n = 4

		switch op {
		// Handle string instructions
		case opcode.ISTR:
			s.readBitstring(r, 8)
		case opcode.FSTR:
			s.readBitstring(r, 8)
		case opcode.OCTETS:
			s.readBitstring(r, 8)
		case opcode.UTF8:
			s.readBitstring(r, 8)
		case opcode.NAME:
			s.readBitstring(r, 8)
		case opcode.NIBBLES:
			s.readBitstring(r, 4)
		case opcode.BITS:
			s.readBitstring(r, 1)

		// Float
		case opcode.IEEE754DP:
			r.Seek(8, io.SeekCurrent)
			s.code[s.pos].n += 8
			s.pos += 2

		// Native integer
		case opcode.NATLONG:
			var i int32
			if err := binary.Read(r, binary.LittleEndian, &i); err != nil {
				s.recordError(err)
				return
			}
			s.code[s.pos].arg = int(i)
			s.code[s.pos].n += 4
			s.pos++

		// Record block-hierarchy
		case opcode.SCAN:
			stack = append(stack, s.pos)
		case opcode.BLOCK:
			if len(stack) == 0 {
				s.recordError(ErrUnmatchedBLOCK)
				return
			}
			scan := stack[len(stack)-1]
			s.code[scan].arg = s.pos
			s.code[s.pos].arg = scan
			stack = stack[:len(stack)-1]

		// Deprecated Instructions
		case opcode.ESWAP:
			s.recordError(ErrDeprecatedESWAP)
			return
		case opcode.WIDEN:
			s.recordError(ErrDeprecatedWIDEN)
			return
		}
		s.pos++
	}
}

// decodeBitstrings reads a binary string from a reader. Argument factor is required to
// convert t3xf the length field into bits. For example if the length field encodes
// bytes, pass Octett. if the field encodes Hexstring, pass Nibble.
func (s *Scanner) readBitstring(r io.ReadSeeker, factor int) {

	// Read length field.
	var n uint32
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		s.recordError(err)
	}

	bits := int(n) * factor
	words := (bits + 31) / 32
	r.Seek(int64(words)*4, io.SeekCurrent)

	s.code[s.pos].arg = bits
	s.code[s.pos].n += (1 + words) * 4
	s.pos += 1 + words
}

func (s *Scanner) recordError(err error) {
	if s.err == nil {
		s.err = err
	}
	s.done = true
}
