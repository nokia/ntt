// Package asn1 provides a pragmatic, hand-written front-end for ASN.1
// source files referenced from TTCN-3 test suites.
//
// 3GPP TTCN-3 suites commonly import ASN.1 modules for protocol message
// definitions (e.g. RRC, NGAP). Until now ntt had no way to understand
// those files at all - even producing a "module XYZ not found" error
// when the importing TTCN-3 file mentioned them. This package fills
// that gap with the same level of fidelity vanadium does in its initial
// ASN.1 layer: header parsing (module identifier, oid, tagging
// defaults, EXPORTS, IMPORTS) plus a coarse pass over the body to
// extract assignment names. That is enough to:
//
//   - Resolve `import from ASN1Module all` style references.
//   - Power "Go to definition" jumps from TTCN-3 into the corresponding
//     ASN.1 assignment.
//   - Surface helpful diagnostics when a referenced assignment doesn't
//     exist in the imported ASN.1 module.
//
// A full ASN.1 type checker (and round-trip transformation into the
// TTCN-3 type system) is tracked separately. This package is the
// minimum that unblocks the LSP for ASN.1-heavy suites.
package asn1

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Module is the in-memory representation of a single ASN.1 module.
type Module struct {
	// Name is the module identifier, e.g. "RRC-PDU-Definitions".
	Name string

	// OID is the optional object identifier following the module
	// name, including the surrounding braces (e.g. "{ itu-t (0) ... }").
	OID string

	// TaggingDefault is one of "EXPLICIT", "IMPLICIT", "AUTOMATIC"
	// or the empty string when unspecified.
	TaggingDefault string

	// Imports is a list of "module -> symbol names" mappings,
	// preserving source order.
	Imports []Import

	// Exports lists explicitly EXPORTed assignments. Empty means
	// "EXPORTS ALL" (or no EXPORTS clause at all - both are treated
	// as exporting everything).
	Exports []string

	// Assignments lists every type/value/object assignment found in
	// the module body, in source order. The Kind field reflects an
	// educated guess based on the assignment's first non-whitespace
	// token after `::=`.
	Assignments []Assignment

	// Filename is the path that produced this module, when known.
	Filename string

	// Diagnostics records issues encountered while parsing.
	Diagnostics []Diagnostic
}

// Import is a single "FROM Module" clause in an IMPORTS block.
type Import struct {
	From    string   // the source module name
	Symbols []string // symbols imported; empty means "IMPORTS ALL"
}

// AssignmentKind classifies an ASN.1 assignment.
type AssignmentKind int

const (
	UnknownKind AssignmentKind = iota
	TypeKind
	ValueKind
	ObjectClassKind
)

// Assignment is a single `Name ::= ...` entry in the module body.
type Assignment struct {
	Name string
	Kind AssignmentKind
}

// Diagnostic is a parse-time issue with a source line for context.
type Diagnostic struct {
	Line    int
	Column  int
	Message string
}

// ParseFile reads and parses the ASN.1 source at path.
func ParseFile(path string) (*Module, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := Parse(b)
	m.Filename = path
	return m, nil
}

// Parse parses src as an ASN.1 module and returns the result. The
// returned *Module is always non-nil; check Diagnostics for parse
// issues. Unrecognised constructs are tolerated and skipped, which
// matches what users expect from an LSP front-end.
func Parse(src []byte) *Module {
	p := newParser(string(src))
	mod := p.parseModule()
	mod.Diagnostics = append(mod.Diagnostics, p.diags...)
	return mod
}

// parser is intentionally simple: it operates on a string and a byte
// offset and uses Go's unicode helpers for character classification.
// ASN.1 is line-oriented enough that this gives the same fidelity as a
// hand-written scanner without the boilerplate.
type parser struct {
	src   string
	pos   int
	line  int
	col   int
	diags []Diagnostic
}

func newParser(src string) *parser {
	return &parser{src: src, line: 1, col: 1}
}

func (p *parser) eof() bool { return p.pos >= len(p.src) }

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) advance() byte {
	if p.eof() {
		return 0
	}
	b := p.src[p.pos]
	p.pos++
	if b == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return b
}

// skipWhitespaceAndComments eats spaces, tabs, newlines and ASN.1 line
// comments (`--`). Block comments are uncommon in protocol files but
// supported for completeness.
func (p *parser) skipWhitespaceAndComments() {
	for !p.eof() {
		c := p.peek()
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			p.advance()
		case c == '-' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '-':
			// Line comment.
			for !p.eof() {
				ch := p.advance()
				if ch == '\n' {
					break
				}
			}
		case c == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '*':
			p.advance()
			p.advance()
			for !p.eof() {
				ch := p.advance()
				if ch == '*' && p.peek() == '/' {
					p.advance()
					break
				}
			}
		default:
			return
		}
	}
}

func (p *parser) readWhile(pred func(byte) bool) string {
	start := p.pos
	for !p.eof() && pred(p.peek()) {
		p.advance()
	}
	return p.src[start:p.pos]
}

func isIdentChar(b byte) bool {
	r := rune(b)
	return r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (p *parser) readIdentifier() string {
	p.skipWhitespaceAndComments()
	if p.eof() {
		return ""
	}
	c := p.peek()
	if !unicode.IsLetter(rune(c)) {
		return ""
	}
	return p.readWhile(isIdentChar)
}

func (p *parser) expectKeyword(kw string) bool {
	saved := *p
	p.skipWhitespaceAndComments()
	if strings.HasPrefix(p.src[p.pos:], kw) {
		end := p.pos + len(kw)
		if end == len(p.src) || !isIdentChar(p.src[end]) {
			for i := 0; i < len(kw); i++ {
				p.advance()
			}
			return true
		}
	}
	*p = saved
	return false
}

func (p *parser) parseModule() *Module {
	m := &Module{}

	name := p.readIdentifier()
	if name == "" {
		p.error("expected module identifier")
		return m
	}
	m.Name = name

	// Optional object identifier.
	p.skipWhitespaceAndComments()
	if p.peek() == '{' {
		m.OID = p.readBalanced('{', '}')
	}

	if !p.expectKeyword("DEFINITIONS") {
		p.error("expected DEFINITIONS keyword")
		return m
	}

	// Tagging default.
	for _, kw := range []string{"EXPLICIT", "IMPLICIT", "AUTOMATIC"} {
		if p.expectKeyword(kw) {
			m.TaggingDefault = kw
			break
		}
	}
	p.expectKeyword("TAGS")
	p.expectKeyword("EXTENSIBILITY")
	p.expectKeyword("IMPLIED")

	p.skipWhitespaceAndComments()
	if p.peek() == ':' {
		// Expect "::= BEGIN"
		p.advance()
		p.advance()
		p.advance() // = sign
	}
	p.expectKeyword("BEGIN")

	m.Exports = p.parseExports()
	m.Imports = p.parseImports()
	m.Assignments = p.parseAssignments()
	return m
}

func (p *parser) parseExports() []string {
	if !p.expectKeyword("EXPORTS") {
		return nil
	}
	p.skipWhitespaceAndComments()
	// "EXPORTS ALL ;" exports everything; we return nil to signal that.
	if p.expectKeyword("ALL") {
		p.skipUntilSemicolon()
		return nil
	}
	var out []string
	for !p.eof() {
		p.skipWhitespaceAndComments()
		if p.peek() == ';' {
			p.advance()
			return out
		}
		id := p.readIdentifier()
		if id == "" {
			p.advance()
			continue
		}
		out = append(out, id)
		p.skipWhitespaceAndComments()
		if p.peek() == ',' {
			p.advance()
		}
	}
	return out
}

func (p *parser) parseImports() []Import {
	if !p.expectKeyword("IMPORTS") {
		return nil
	}
	var out []Import
	for !p.eof() {
		p.skipWhitespaceAndComments()
		if p.peek() == ';' {
			p.advance()
			return out
		}
		var symbols []string
		// Read comma-separated symbol list until FROM.
		for !p.eof() {
			p.skipWhitespaceAndComments()
			if p.expectKeyword("FROM") {
				break
			}
			id := p.readIdentifier()
			if id == "" {
				p.advance()
				continue
			}
			symbols = append(symbols, id)
			p.skipWhitespaceAndComments()
			if p.peek() == ',' {
				p.advance()
			}
		}
		from := p.readIdentifier()
		if from == "" {
			p.error("expected module name after FROM")
			return out
		}
		// Skip an optional OID after the module name.
		p.skipWhitespaceAndComments()
		if p.peek() == '{' {
			p.readBalanced('{', '}')
		}
		out = append(out, Import{From: from, Symbols: symbols})
	}
	return out
}

func (p *parser) parseAssignments() []Assignment {
	var out []Assignment
	for !p.eof() {
		p.skipWhitespaceAndComments()
		if p.expectKeyword("END") {
			return out
		}
		name := p.readIdentifier()
		if name == "" {
			// Skip unknown token defensively.
			p.advance()
			continue
		}
		// Walk forward past any (TypeRef | parameter list) tokens
		// to find the `::=`. This handles both type assignments
		// (Name ::=) and value assignments (name Type ::=).
		if !p.advanceTo("::=") {
			p.skipUntilLineStart()
			continue
		}
		// Consume "::="
		p.advance()
		p.advance()
		p.advance()
		p.skipWhitespaceAndComments()
		kind := classify(name, p.peek())
		out = append(out, Assignment{Name: name, Kind: kind})
		// Skip the assignment body. Heuristic: stop at the next
		// top-level identifier-followed-by-"::=" or END.
		p.skipAssignmentBody()
	}
	return out
}

// advanceTo consumes tokens (identifiers, balanced brackets and
// individual characters) until it finds the literal target at the
// current position. Returns false on EOF, on a newline encountered
// without an intervening "{...}" - which would mean the assignment is
// malformed - or after a reasonable token budget.
func (p *parser) advanceTo(target string) bool {
	const maxTokens = 16
	for i := 0; i < maxTokens && !p.eof(); i++ {
		p.skipWhitespaceAndComments()
		if strings.HasPrefix(p.src[p.pos:], target) {
			return true
		}
		switch p.peek() {
		case '{':
			p.readBalanced('{', '}')
		case '(':
			p.readBalanced('(', ')')
		case '[':
			p.readBalanced('[', ']')
		default:
			if id := p.readIdentifier(); id == "" {
				return false
			}
		}
	}
	return false
}

// isAssignmentStart looks ahead from p.pos and reports whether the next
// non-whitespace tokens form the start of an ASN.1 top-level
// assignment. It does not consume input.
func isAssignmentStart(p *parser) bool {
	probe := *p
	probe.skipWhitespaceAndComments()
	if probe.eof() {
		return false
	}
	if !unicode.IsLetter(rune(probe.peek())) {
		return false
	}
	for i := 0; i < 4 && !probe.eof(); i++ {
		probe.skipWhitespaceAndComments()
		if strings.HasPrefix(probe.src[probe.pos:], "::=") {
			return true
		}
		if probe.peek() == '\n' {
			return false
		}
		switch probe.peek() {
		case '{':
			probe.readBalanced('{', '}')
		case '(':
			probe.readBalanced('(', ')')
		default:
			if id := probe.readIdentifier(); id == "" {
				return false
			}
		}
	}
	return false
}

func classify(name string, lookahead byte) AssignmentKind {
	// ASN.1 convention: types start uppercase, values lowercase. The
	// lookahead helps disambiguate object class assignments which can
	// be uppercase but begin with a CLASS keyword.
	if name == "" {
		return UnknownKind
	}
	first := rune(name[0])
	switch {
	case unicode.IsUpper(first):
		if lookahead == 'C' {
			return ObjectClassKind
		}
		return TypeKind
	case unicode.IsLower(first):
		return ValueKind
	}
	return UnknownKind
}

func (p *parser) skipAssignmentBody() {
	// We walk until we either reach END or detect a new top-level
	// `Name ::=`. Track bracket depth so we don't terminate inside
	// nested structures.
	depth := 0
	for !p.eof() {
		c := p.peek()
		switch c {
		case '{', '(', '[':
			depth++
			p.advance()
		case '}', ')', ']':
			depth--
			p.advance()
		case '\n':
			p.advance()
			if depth == 0 {
				saved := *p
				p.skipWhitespaceAndComments()
				if p.expectKeyword("END") {
					*p = saved
					return
				}
				// A new top-level assignment looks like one of:
				//   Name ::=
				//   Name Type ::=
				//   Name { args } ::=
				// Walk forward up to a handful of identifiers
				// or a balanced brace until we find "::=" on
				// the same logical line.
				if isAssignmentStart(p) {
					*p = saved
					return
				}
				*p = saved
				p.advance()
			}
		case '-':
			if p.pos+1 < len(p.src) && p.src[p.pos+1] == '-' {
				p.skipWhitespaceAndComments()
				continue
			}
			p.advance()
		default:
			p.advance()
		}
	}
}

func (p *parser) skipUntilSemicolon() {
	for !p.eof() {
		if p.advance() == ';' {
			return
		}
	}
}

func (p *parser) skipUntilLineStart() {
	for !p.eof() {
		if p.advance() == '\n' {
			return
		}
	}
}

// readBalanced reads a balanced run of bytes starting at open and ending
// at the matching close. The returned string includes both delimiters.
func (p *parser) readBalanced(open, close byte) string {
	if p.peek() != open {
		return ""
	}
	start := p.pos
	depth := 0
	for !p.eof() {
		c := p.advance()
		switch c {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return p.src[start:p.pos]
			}
		}
	}
	return p.src[start:]
}

func (p *parser) error(msg string) {
	p.diags = append(p.diags, Diagnostic{
		Line:    p.line,
		Column:  p.col,
		Message: msg,
	})
}

// String renders a Module's exported summary for debugging.
func (m *Module) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "module %s", m.Name)
	if m.TaggingDefault != "" {
		fmt.Fprintf(&b, " %s TAGS", m.TaggingDefault)
	}
	names := make([]string, 0, len(m.Assignments))
	for _, a := range m.Assignments {
		names = append(names, a.Name)
	}
	sort.Strings(names)
	fmt.Fprintf(&b, " (%d defs: %s)", len(names), strings.Join(names, ", "))
	return b.String()
}

// HasAssignment reports whether m exports an assignment of the given
// name. When no explicit EXPORTS clause is present every assignment is
// considered exported.
func (m *Module) HasAssignment(name string) bool {
	if len(m.Exports) > 0 {
		for _, e := range m.Exports {
			if e == name {
				return true
			}
		}
		return false
	}
	for _, a := range m.Assignments {
		if a.Name == name {
			return true
		}
	}
	return false
}
