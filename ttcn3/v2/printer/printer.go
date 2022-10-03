// Package printer implements pretty printers for TTCN-3 source code.
package printer

import (
	"fmt"
	"io"

	"github.com/nokia/ntt/ttcn3/v2/syntax"
)

// Fprint formats src in canonical TTCN-3 style and writes the result to w or
// returns an (I/O or syntax) error. src is expected to be syntactically
// correct TTCN-3 source text.
func Fprint(w io.Writer, src interface{}) error {
	p := &CanonicalPrinter{
		Indent: "\t",
		w:      w,
	}
	return p.Fprint(src)
}

// CanonicalPrinter is a simple formatter that only fixes indentation and
// various whitespace issues.
type CanonicalPrinter struct {
	Indent     string // Indentation string; default is "\t"
	w          io.Writer
	lastPos    syntax.Position
	indent     int
	needIndent bool
}

func (p *CanonicalPrinter) Fprint(v interface{}) error {

	var (
		n syntax.Node

		// The simple formatting rules do not need any context information
		// so far. This allows us to use the tokeniser and release initial
		// formatting experiments even before the parser is ready.
		parse = syntax.Tokenize
	)

	switch v := v.(type) {
	case []byte:
		n = parse(v)
	case string:
		n = parse([]byte(v))
	case io.Reader:
		b, err := io.ReadAll(v)
		if err != nil {
			return err
		}
		n = parse(b)
	default:
		return fmt.Errorf("printer: unsupported type %T", v)
	}

	// Only pretty print if there are no syntax errors.
	if n.Err() != nil {
		return n.Err()
	}

	// Prime the position tracker with the first token.
	if tok := n.FirstToken(); tok != syntax.Nil {
		s := tok.Span()
		p.lastPos = s.End
	}

	return p.node(n)
}

func (p *CanonicalPrinter) node(n syntax.Node) error {
	n.Inspect(func(n syntax.Node) bool {
		if n == syntax.Nil || !n.IsToken() {
			return true
		}

		// Handle user defined white space
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
		// decrement before every closing brace.
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
	return nil
}

func (p *CanonicalPrinter) print(v interface{}) {
	if p.needIndent {
		for i := 0; i < p.indent; i++ {
			fmt.Fprint(p.w, p.Indent)
		}
		p.needIndent = false
	}
	fmt.Fprint(p.w, v)
}
