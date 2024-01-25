package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/spf13/cobra"
)

var (
	T3xfasmCommand = &cobra.Command{
		Use:   "t3xfasm <file>",
		Short: "Assemble text file with t3xf instructions to T3XF binary file",
		Long: `Assemble text file with t3xf instructions to T3XF binary file.

This commands implements a simple assembler to generate T3XF binary files.
Every line in the input file represents a single instruction. Empty lines or
lines with just a comment are generated as NOP instructions. Comments start
with a semicolon (';') and continue to the end of the line. The assembler is
case-insensitive. Availabl instructions can be found in the opcodes.yml file in
the github.com/nokia/ntt/k3/t3xf/opcode package.

Instructions optionally take an argument. References as start with '@'.
When referencing an TTCN-3 entity use the line-number in decimal.
Line instructions start with '='.


Example:

	nop 	; the next three instructions server as a header and BOM.
	natlong	2
	version

	; var integer x := 2 + 3 (Note: above and this line will become a nop)
	integer
	name x
	var
	natlong 2
	natlong 3
	add
	@8
	assign

	; log(x)
	@8
	log
	goto @9

`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}

			e := t3xf.NewEncoder()
			s := bufio.NewScanner(f)
			line := 0
			for s.Scan() {
				line++
				ls := scanner{src: s.Bytes()}
				op, err := ls.parseOpcode()
				if err != nil {
					return fmt.Errorf("%s:%d: %w", args[0], line, err)
				}

				arg, err := ls.parseArgument()
				if err != nil {
					return fmt.Errorf("%s:%d: %w", args[0], line, err)
				}

				ls.scanWhitespace()
				if ls.pos < len(ls.src) && ls.src[ls.pos] != ';' {
					return fmt.Errorf("%s:%d: %w: unexpected trailing characters", args[0], line, ErrSyntax)
				}
				if err := e.Encode(op, arg); err != nil {
					return fmt.Errorf("%s:%d: %w", args[0], line, err)
				}

			}
			if err := s.Err(); err != nil {
				return err
			}

			b, err := e.Assemble()
			if err != nil {
				return err
			}
			out := args[0][:len(args[0])-len(filepath.Ext(args[0]))] + ".t3xf"
			return os.WriteFile(out, b, 0644)
		},
	}

	ErrSyntax = fmt.Errorf("syntax error")
)

type scanner struct {
	src []byte
	pos int
}

func (s *scanner) parseOpcode() (opcode.Opcode, error) {
	s.scanWhitespace()
	if s.pos >= len(s.src) {
		return opcode.NOP, nil
	}

	pos := s.pos
	s.pos++
	ch := s.src[pos]

	switch {
	case isAlpha(ch):
		s.scanAlnum()
		word := strings.ToLower(string(s.src[pos:s.pos]))
		return opcode.Parse(word)

	case ch == ';':
		s.pos = len(s.src)
		return opcode.NOP, nil

	case ch == '=':
		return opcode.LINE, nil

	case ch == '@':
		// We need to backup the position to allow the argument parser
		// to convert lines to zero-based references.
		s.pos--
		return opcode.REF, nil

	default:
		return -1, fmt.Errorf("%w: unexpected character %q", ErrSyntax, ch)
	}
}

func (s *scanner) parseArgument() (interface{}, error) {
	s.scanWhitespace()
	if s.pos >= len(s.src) {
		return nil, nil
	}

	pos := s.pos
	s.pos++
	ch := s.src[pos]

	switch {
	case isAlpha(ch):
		s.scanAlnum()
		return string(s.src[pos:s.pos]), nil

	case ch == '"':
		s.scanString()
		v, err := syntax.Unquote(string(s.src[pos:s.pos]))
		if err != nil {
			return nil, err
		}
		return v, nil

	case isDigit(ch) || ch == '-' || ch == '+':
		s.scanFloat()
		f, err := strconv.ParseFloat(string(s.src[pos:s.pos]), 64)
		if err != nil {
			return nil, err
		}
		if f == math.Trunc(f) {
			return int(f), nil
		}
		return f, nil

	case ch == '@':
		s.scanFloat()
		f, err := strconv.ParseFloat(string(s.src[pos+1:s.pos]), 64)
		if err != nil {
			return nil, err
		}
		if f != math.Trunc(f) {
			return nil, fmt.Errorf("%w: integer expected", ErrSyntax)
		}
		return int(f) - 1, nil

	case ch == '\'':
		s.scanBitstring()
		return t3xf.NewString(0, nil), fmt.Errorf("not implemented")

	case ch == ';':
		s.pos = len(s.src)
		return nil, nil

	default:
		return -1, fmt.Errorf("%w: unexpected character %q", ErrSyntax, ch)
	}
}

func (s *scanner) scanWhitespace() {
	for s.pos < len(s.src) {
		switch ch := s.src[s.pos]; ch {
		case ' ', '\t', '\r':
		default:
			return
		}
		s.pos++
	}
}

func (s *scanner) scanAlnum() {
	for s.pos < len(s.src) && isAlnum(s.src[s.pos]) {
		s.pos++
	}
}

func (s *scanner) scanString() {
	s.pos-- // backup for proper quoting ("")
	for {
		s.pos++
		if s.pos >= len(s.src) {
			return
		}

		switch ch := s.src[s.pos]; ch {
		case '\\':
			s.pos++
		case '"':
			s.pos++
			if s.pos >= len(s.src) || s.src[s.pos] != '"' {
				return
			}
		}
	}
}

func (s *scanner) scanBitstring() {
L:
	for {
		if s.pos >= len(s.src) {
			return
		}
		switch ch := s.src[s.pos]; ch {
		case '\'':
			s.pos++
			break L
		}
		s.pos++
	}
	s.scanAlnum()
}

func (s *scanner) scanFloat() {
	if s.src[s.pos-1] != '0' {
		s.scanDigits()
	}

	// scan fractional part
	if s.pos < len(s.src) && s.src[s.pos] == '.' {
		// check '..' token
		if s.pos+1 < len(s.src) && s.src[s.pos+1] == '.' {
			return
		}
		s.pos++
		s.scanDigits()
	}

	// scan exponent
	if s.pos < len(s.src) && (s.src[s.pos] == 'e' || s.src[s.pos] == 'E') {
		s.pos++
		if s.pos < len(s.src) && (s.src[s.pos] == '+' || s.src[s.pos] == '-') {
			s.pos++
		}
		if s.pos < len(s.src) && isDigit(s.src[s.pos]) {
			s.scanDigits()
		}
	}
}

func (s *scanner) scanDigits() {
	for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.pos++
	}
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
