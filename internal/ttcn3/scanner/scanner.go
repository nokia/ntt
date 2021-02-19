// Package scanner provides a TTCN-3 scanner
package scanner

import (
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

// An ErrorHandler may be provided to Scanner.Init. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending
//
type ErrorHandler func(pos loc.Position, msg string)

// A Scanner holds the scanner's internal state while processing
// a given text. It can be allocated as part of another data
// structure but must be initialized via Init before use.
//
type Scanner struct {
	// immutable state
	file *loc.File    // source file handle
	dir  string       // directory portion of file.Name()
	src  []byte       // source
	Err  ErrorHandler // error reporting; or nil

	// scanning state
	ch         rune // current character
	offset     int  // character offset
	rdOffset   int  // reading offset (position after current character)
	lineOffset int  // current line offset

	// public state - ok to modify
	ErrorCount int // number of errors encountered
}

const bom = 0xFEFF // byte order mark, only permitted as very first character

// Read the next Unicode char into s.ch.
// s.ch < 0 means end-of-file.
//
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "illegal byte order mark")
			}
		}
		s.rdOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = -1 // eof
	}
}

// A SMode value is a set of flags (or 0).
// They control scanner behavior.
//
type SMode uint

const (
	ScanComments SMode = 1 << iota // return comments as COMMENT tokens
)

// Init prepares the scanner s to tokenize the text src by setting the
// scanner at the beginning of src. The scanner uses the file set file
// for position information and it adds line information for each line.
// It is ok to re-use the same file when re-scanning the same file as
// line information which is already present is ignored. Init causes a
// panic if the file size does not match the src size.
//
// Calls to Scan will invoke the error handler Err if they encounter a
// syntax error and Err is not nil. Also, for each error encountered,
// the Scanner field ErrorCount is incremented by one.
//
// Note that Init may call Err if there is an error in the first character
// of the file.
//
func (s *Scanner) Init(file *loc.File, src []byte, err ErrorHandler) {
	// Explicitly initialize all fields since a scanner may be reused.
	if file.Size() != len(src) {
		panic(fmt.Sprintf("file size (%d) does not match src len (%d)", file.Size(), len(src)))
	}
	s.file = file
	s.dir, _ = filepath.Split(file.Name())
	s.src = src
	s.Err = err

	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.lineOffset = 0
	s.ErrorCount = 0

	s.next()
	if s.ch == bom {
		s.next() // ignore BOM at file beginning
	}
}

func (s *Scanner) error(offs int, msg string) {
	if s.Err != nil {
		s.Err(s.file.Position(s.file.Pos(offs)), msg)
	}
	s.ErrorCount++
}

func (s *Scanner) scanComment() string {
	// initial '/' already consumed; s.ch == '/' || s.ch == '*'
	offs := s.offset - 1 // position of initial '/'
	hasCR := false

	if s.ch == '/' {
		//-style comment
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				hasCR = true
			}
			s.next()
		}
		s.next()
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			hasCR = true
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			goto exit
		}
	}

	s.error(offs, "comment not terminated")

exit:
	lit := s.src[offs:s.offset]
	if hasCR {
		lit = stripCR(lit)
	}

	return string(lit)
}

func (s *Scanner) scanPreproc() string {
	offs := s.offset - 1
	hasCR := false
	for s.ch != '\n' && s.ch >= 0 {
		if s.ch == '\r' {
			hasCR = true
		}
		s.next()
	}

	//if offs != s.lineOffset {
	//	s.error(s.offset, "preprocessor statement must start at line")
	//}

	lit := s.src[offs:s.offset]
	if hasCR {
		lit = stripCR(lit)
	}

	return string(lit)
}

func isAlpha(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanModifier() string {
	offs := s.offset - 1
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanTitanMacro() string {
	offs := s.offset - 1
	for isAlpha(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func (s *Scanner) scanDigits() {
	for isDigit(s.ch) {
		s.next()
	}
}

func (s *Scanner) scanNumber() (tok token.Kind, lit string) {
	offs := s.offset
	tok = token.INT

	if s.ch == '0' {
		s.next()
	} else {
		s.scanDigits()
	}

	// fraction
	if s.ch == '.' {
		// check for RANGE token (example: 0..23)
		if s.rdOffset < len(s.src) && s.src[s.rdOffset] == '.' {
			goto out
		}
		tok = token.FLOAT
		s.next()
		s.scanDigits()
	}

	// exponent
	if s.ch == 'e' || s.ch == 'E' {
		tok = token.FLOAT
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		s.scanDigits()
	}

	// FIXME: In standard TTCN-3:2014 some identifiers start with a number.
	//        For instance predefined function 291oolea.
	if isAlpha(s.ch) {
		tok = token.ILLEGAL
		for isAlpha(s.ch) || isDigit(s.ch) {
			s.next()
		}
		s.error(offs, "malformed number")
	}

out:
	return tok, string(s.src[offs:s.offset])
}

func (s *Scanner) scanBString() string {
	// opening ' already consumed
	offs := s.offset - 1

	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '\'' {
			if isAlpha(s.ch) {
				s.next()
				break
			}
			s.error(s.offset-1, "missing string specifier (O, H or B)")
			break
		}
		if ch == '\\' {
			s.next()
		}
	}

	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanString() string {
	// '"' opening already consumed
	offs := s.offset - 1

	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "string literal not terminated")
			break
		}
		s.next()
		if ch == '"' {
			if s.ch == '"' {
				s.next()
			} else {
				break
			}
		}
		if ch == '\\' {
			s.next()
		}
	}

	return string(s.src[offs:s.offset])
}

func stripCR(b []byte) []byte {
	c := make([]byte, len(b))
	i := 0
	for _, ch := range b {
		if ch != '\r' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) switch2(ch rune, tok1, tok2 token.Kind) token.Kind {
	if s.ch == ch {
		s.next()
		return tok1
	}
	return tok2
}

// Scan scans the next token and returns the token position, the token,
// and its literal string if applicable. The source end is indicated by
// EOF.
//
// If the returned token is a literal (IDENT, INT, FLOAT,
// IMAG, CHAR, STRING) or COMMENT, the literal string
// has the corresponding value.
//
// If the returned token is a keyword, the literal string is the keyword.
//
// If the returned token is ILLEGAL, the literal string is the
// offending character.
//
// In all other cases, Scan returns an empty literal string.
//
// For more tolerant parsing, Scan will return a valid token if
// possible even if a syntax error was encountered. Thus, even
// if the resulting token sequence contains no illegal tokens,
// a client may not assume that no error occurred. Instead it
// must check the scanner's ErrorCount or the number of calls
// of the error handler, if there was one installed.
//
// Scan adds line information to the file added to the file
// set with Init. Kind positions are relative to that file
// and thus relative to the file set.
//
func (s *Scanner) Scan() (pos loc.Pos, tok token.Kind, lit string) {
	s.skipWhitespace()

	// current token start
	pos = s.file.Pos(s.offset)

	// determine token value
	switch ch := s.ch; {
	case isAlpha(ch):
		lit = s.scanIdentifier()
		if len(lit) > 1 {
			// keywords are longer than one letter - avoid lookup otherwise
			tok = token.Lookup(lit)
		} else {
			tok = token.IDENT
		}
	case isDigit(ch):
		tok, lit = s.scanNumber()
	default:
		s.next() // always make progress
		switch ch {
		case -1:
			tok = token.EOF
		case '"':
			tok = token.STRING
			lit = s.scanString()
		case '\'':
			tok = token.BSTRING
			lit = s.scanBString()
		case '@':
			if s.ch == '>' {
				tok = token.ROR
				s.next()
			} else {
				tok = token.MODIF
				lit = s.scanModifier()
			}
		case '#':
			tok = token.PREPROC
			lit = s.scanPreproc()
		case '%':
			tok = token.IDENT
			lit = s.scanTitanMacro()
		case '.':
			tok = s.switch2('.', token.RANGE, token.DOT)
		case ',':
			tok = token.COMMA
		case ';':
			tok = token.SEMICOLON
		case '(':
			tok = token.LPAREN
		case ')':
			tok = token.RPAREN
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case '+':
			tok = token.ADD
		case '-':
			tok = s.switch2('>', token.REDIR, token.SUB)
		case '*':
			tok = token.MUL
		case '/':
			if s.ch == '/' || s.ch == '*' {
				comment := s.scanComment()
				tok = token.COMMENT
				lit = comment
			} else {
				tok = token.DIV
			}
		case ':':
			switch s.ch {
			case '=':
				tok = token.ASSIGN
				s.next()
			case ':':
				tok = token.COLONCOLON
				s.next()
			default:
				tok = token.COLON
			}
		case '>':
			switch s.ch {
			case '>':
				tok = token.SHR
				s.next()
			case '=':
				tok = token.GE
				s.next()
			default:
				tok = token.GT
			}

		case '<':
			switch s.ch {
			case '<':
				tok = token.SHL
				s.next()
			case '@':
				tok = token.ROL
				s.next()
			case '=':
				tok = token.LE
				s.next()
			default:
				tok = token.LT
			}

		case '=':
			switch s.ch {
			case '=':
				tok = token.EQ
				s.next()
			case '>':
				tok = token.DECODE
				s.next()
			default:
				tok = token.ILLEGAL
				lit = "="
				s.error(s.offset, fmt.Sprintf("stray %q", "="))

			}
		case '!':
			tok = s.switch2('=', token.NE, token.EXCL)
		case '&':
			tok = token.CONCAT
		case '?':
			tok = token.ANY

		default:
			// next reports unexpected BOMs - don't repeat
			if ch != bom {
				s.error(s.file.Offset(pos), fmt.Sprintf("illegal character %#U", ch))
			}
			tok = token.ILLEGAL
			lit = string(ch)
		}
	}

	return
}
