package syntax

// Tokenize given source code and return a root node with all the tokens.
func Tokenize(src []byte) Node {
	var b Builder
	b.content = src
	root := b.Push(Root)
	s := NewScanner(src)
	for {
		kind, begin, end := s.Scan()
		if kind == EOF {
			break
		}
		b.PushToken(kind, begin, end)
	}
	b.Pop()
	root.tree.lines = s.Lines()
	return root
}

// NewScanner returns a new TTCN-3 scanner for src.
func NewScanner(src []byte) *Scanner {
	return &Scanner{
		src:   src,
		lines: []int{0},
	}
}

// Scanner scans a TTCN-3 source.
type Scanner struct {
	lines []int
	src   []byte
	pos   int
}

// Lines returns the line offsets of the source.
func (s *Scanner) Lines() []int {
	return s.lines
}

// Scan returns the next token and its range.
func (s *Scanner) Scan() (Kind, int, int) {
	s.scanWhitespace()

	if s.pos >= len(s.src) {
		return EOF, s.pos, s.pos + 1
	}

	pos := s.pos
	s.pos++
	ch := s.src[pos]

	var (
		typ  Kind
		next byte
	)

	if s.pos < len(s.src) {
		next = s.src[s.pos]
	}

	switch {
	case isAlpha(ch):
		typ = Identifier
		s.scanAlnum()
	case isDigit(ch):
		typ = s.scanNumber()
	case ch == ',':
		typ = Comma
	case ch == '+':
		typ = Add
	case ch == '*':
		typ = Mul
	case ch == '&':
		typ = Concat
	case ch == '?':
		typ = Any
	case ch == '(':
		typ = LeftParen
	case ch == '[':
		typ = LeftBracket
	case ch == '{':
		typ = LeftBrace
	case ch == ')':
		typ = RightParen
	case ch == ']':
		typ = RightBracket
	case ch == '}':
		typ = RightBrace
	case ch == ';':
		typ = Semicolon
	case ch == '/':
		switch next {
		case '/':
			s.scanLine()
			typ = Comment
		case '*':
			typ = s.scanMultiLineComment()
		default:
			typ = Div
		}
	case ch == '@':
		switch {
		case isAlpha(next):
			s.scanAlnum()
			typ = Modifier
		case next == '>':
			s.pos++
			typ = RotateRight
		default:
			typ = Unknown
		}
	case ch == '%':
		if isAlpha(next) {
			s.scanAlnum()
			typ = Identifier
		}
	case ch == '!':
		typ = Exclude
		if next == '=' {
			s.pos++
			typ = NotEqual
		}
	case ch == '-':
		typ = Sub
		if next == '>' {
			s.pos++
			typ = Arrow
		}
	case ch == '.':
		typ = Dot
		if next == '.' {
			s.pos++
			typ = DotDot
		}
	case ch == ':':
		switch next {
		case ':':
			s.pos++
			typ = ColonColon
		case '=':
			s.pos++
			typ = Assign
		default:
			typ = Colon
		}
	case ch == '<':
		switch next {
		case '<':
			s.pos++
			typ = ShiftLeft
		case '=':
			s.pos++
			typ = LessEqual
		case '@':
			s.pos++
			typ = RotateLeft
		default:
			typ = Less
		}
	case ch == '=':
		switch next {
		case '=':
			s.pos++
			typ = Equal
		case '>':
			s.pos++
			typ = DoubleArrow
		}
	case ch == '>':
		switch next {
		case '>':
			s.pos++
			typ = ShiftRight
		case '=':
			s.pos++
			typ = GreaterEqual
		default:
			typ = Greater
		}
	case ch == '\'':
		typ = s.scanBitstring()
	case ch == '#':
		s.scanLine()
		typ = Preproc
	case ch == '"':
		typ = s.scanString()
	default:
		typ = Unknown
	}

	return typ, pos, s.pos
}

func (s *Scanner) scanWhitespace() {
	for s.pos < len(s.src) {
		switch ch := s.src[s.pos]; ch {
		case ' ', '\t', '\r':
		case '\n', '\v', '\f':
			s.lines = append(s.lines, s.pos+1)
		default:
			return
		}
		s.pos++
	}
}

func (s *Scanner) scanLine() {
	for s.pos < len(s.src) && s.src[s.pos] != '\n' {
		s.pos++
	}

	//  new lines belong to the comment token. EOF does not.
	if s.pos < len(s.src) {
		s.lines = append(s.lines, s.pos+1)
		s.pos++
	}
}

func (s *Scanner) scanAlnum() {
	for s.pos < len(s.src) && isAlnum(s.src[s.pos]) {
		s.pos++
	}
}

func (s *Scanner) scanMultiLineComment() Kind {
	s.pos++ // skip the first '*'
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == '\n' || ch == '\v' || ch == '\f' {
			s.lines = append(s.lines, s.pos+1)
		}
		s.pos++
		if ch == '*' && s.pos < len(s.src) && s.src[s.pos] == '/' {
			s.pos++
			return Comment
		}
	}

	return Unterminated
}

func (s *Scanner) scanString() Kind {
	s.pos-- // backup for proper quoting ("")
	for {
		s.pos++
		if s.pos >= len(s.src) {
			return Unterminated
		}

		switch ch := s.src[s.pos]; ch {
		case '\n', '\v', '\f':
			s.lines = append(s.lines, s.pos+1)
			s.pos++
		case '\\':
			s.pos++
		case '"':
			s.pos++
			if s.pos >= len(s.src) || s.src[s.pos] != '"' {
				return String
			}
		}
	}
}

func (s *Scanner) scanBitstring() Kind {
L:
	for {
		if s.pos >= len(s.src) {
			return Unterminated
		}
		switch ch := s.src[s.pos]; ch {
		case '\n', '\v', '\f':
			s.lines = append(s.lines, s.pos+1)
		case '\'':
			s.pos++
			break L
		}
		s.pos++
	}

	typ := Bitstring
	if s.pos >= len(s.src) || !isAlpha(s.src[s.pos]) {
		typ = Malformed
	}

	s.scanAlnum()
	return typ
}

func (s *Scanner) scanNumber() Kind {
	tok := Integer

	// We already increatemented s.pos in caller, but we must check for a single zero first.
	s.pos--

	if s.src[s.pos] == '0' {
		s.pos++
	} else {
		s.scanDigits()
	}

	// scan fractional part
	if s.pos < len(s.src) && s.src[s.pos] == '.' {
		tok = Float
		s.pos++
		s.scanDigits()
	}

	// scan exponent
	if s.pos < len(s.src) && (s.src[s.pos] == 'e' || s.src[s.pos] == 'E') {
		tok = Float
		s.pos++
		if s.pos < len(s.src) && (s.src[s.pos] == '+' || s.src[s.pos] == '-') {
			s.pos++
		}
		if s.pos >= len(s.src) || !isDigit(s.src[s.pos]) {
			tok = Malformed
		} else {
			s.scanDigits()
		}
	}

	// handle trailing carbage
	if s.pos < len(s.src) && isAlnum(s.src[s.pos]) {
		tok = Malformed
		s.scanAlnum()
	}

	return tok
}

func (s *Scanner) scanDigits() {
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
