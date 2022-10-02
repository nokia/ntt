// Package printer implements pretty printers for TTCN-3 source code.
package printer

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/ttcn3/v2/syntax"
)

// Bytes formats src in canonical TTCN-3 style and returns the result or an
// (I/O or syntax) error. src is expected to syntactically correct TTCN-3
// source text.
func Bytes(src []byte) ([]byte, error) {
	p := &printer{Indent: "\t"}
	return p.Bytes(src)
}

// printer is a simple formatter that only fixes indentation and
// various whitespace issues.
type printer struct {
	Indent     string // Indentation string; default is "\t"
	buf        bytes.Buffer
	lastPos    syntax.Position
	indent     int
	needIndent bool
}

func (p *printer) Bytes(src []byte) ([]byte, error) {

	// The simple formatting rules do not need any contextional information
	// so far. This allows us to use the tokenzier and release initial
	// formatting experiments even before the parser is ready.
	tree := syntax.Tokenize(src)

	// Only pretty print if there are no syntax errors.
	if tree.Err() != nil {
		return nil, tree.Err()
	}

	// Prime the position tracker with the first token.
	if tok := tree.FirstToken(); tok != syntax.Nil {
		s := tok.Span()
		p.lastPos = s.End
	}
	tree.Inspect(func(n syntax.Node) bool {
		if n == syntax.Nil || !n.IsToken() {
			return true
		}

		// Handle user defined whitespace
		currPos := n.Span()
		switch {
		case currPos.Begin.Line > p.lastPos.Line:
			p.print("\n")
			if currPos.Begin.Line-p.lastPos.Line > 1 {
				p.print("\n")
			}
			p.needIndent = true
		case currPos.Begin.Column > p.lastPos.Column:
			p.print(" ")
		}
		p.lastPos = currPos.End

		// Handle indent.
		//
		// Indentation is usually handled by non-terminal nodes. But the pretty
		// printer rules are very simple and we can handle it here.
		//
		// The rule is: Increment indentation after every opening brace and
		// decrement bevor every closing brace.
		switch s := n.Text(); s {
		case "{", "[", "(":
			p.print(s)
			p.indent++
		case "}", "]", ")":
			p.indent--
			p.print(s)
		default:
			p.print(s)
		}

		return true
	})

	return p.buf.Bytes(), nil
}

func (p *printer) print(v interface{}) {
	if p.needIndent {
		for i := 0; i < p.indent; i++ {
			fmt.Fprint(&p.buf, p.Indent)
		}
		p.needIndent = false
	}
	fmt.Fprint(&p.buf, v)
}
