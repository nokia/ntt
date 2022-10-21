// Package printer implements pretty printers for TTCN-3 source code.
package printer

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/nokia/ntt/ttcn3/v2/syntax"
)

type whiteSpace byte

const (
	ignore   = whiteSpace(0)
	blank    = whiteSpace(' ')
	newline  = whiteSpace('\n')
	cell     = whiteSpace('|')
	indent   = whiteSpace('>')
	unindent = whiteSpace('<')
)

// Fprint formats src in canonical TTCN-3 style and writes the result to w or
// returns an (I/O or syntax) error. src is expected to be syntactically
// correct TTCN-3 source text.
func Fprint(w io.Writer, src interface{}) error {
	return NewCanonicalPrinter(w).Fprint(src)
}

// CanonicalPrinter is a simple formatter that only fixes indentation and
// various whitespace issues.
type CanonicalPrinter struct {
	Indent           string // Indentation string; default is "\t"
	w                *tabwriter.Writer
	lastPos, currPos syntax.Position
	indent           int
	whiteBuf         string
	firstToken       bool
}

// NewCanonicalPrinter returns a new printer that formats source code.
func NewCanonicalPrinter(w io.Writer) *CanonicalPrinter {
	return &CanonicalPrinter{
		indent: 0,
		Indent: "\t",
		w: tabwriter.NewWriter(w, 0, 8, 1, ' ',
			tabwriter.TabIndent|tabwriter.StripEscape|tabwriter.DiscardEmptyColumns),
	}
}

func (p *CanonicalPrinter) Fprint(v interface{}) error {

	b, err := toBytes(v)
	if err != nil {
		return err
	}

	// The simple formatting rules do not need much context information.
	// This allows us to use the tokeniser and release initial
	// formatting experiments even before the parser is ready.
	n := syntax.Tokenize(b)

	// Only pretty print if there are no syntax errors.
	if n.Err() != nil {
		return n.Err()
	}

	return p.tree(n)
}

// tree prints the syntax tree by interspersing the token stream with spacing
// information.
func (p *CanonicalPrinter) tree(n syntax.Node) error {

	// Prime the position tracker with the first token to avoid printing
	// white-spaces before the first token.
	p.firstToken = true

	n.Inspect(func(n syntax.Node) bool {

		// We figure out spacing by looking at the token stream only.
		// Because the parser for creating a syntax tree is not yet
		// implemented.
		if !n.IsToken() {
			return true
		}

		// Incorporate user-defined line breaks and token separators
		// into output stream.
		currPos := n.Span()
		switch {
		case currPos.Begin.Line > p.lastPos.Line:
			p.print(newline)
			if currPos.Begin.Line-p.lastPos.Line > 1 {
				p.print(newline)
			}
		case currPos.Begin.Column > p.lastPos.Column:
			p.print(blank)
		}
		p.lastPos = currPos.End

		switch k, s := n.Kind(), n.Text(); {

		// Increment indentation after opening {, [, (.
		case k == syntax.LeftBrace, k == syntax.LeftBracket, k == syntax.LeftParen:
			p.print(s, indent)

		// Decrement indentation before closing }, ], ).
		case k == syntax.RightBrace, k == syntax.RightBracket, k == syntax.RightParen:
			p.print(unindent, s)

		// Add space after comma
		case k == syntax.Comma:
			p.print(",", blank)

		// Align assignments.
		case k == syntax.Assign:
			p.print(blank, ":=", blank)

		// Every line of a comment has to be indented individually.
		case k == syntax.Comment:
			p.print(cell)

			// Before we split a comment into its lines, we have to
			// remove the trailing newline of `//` comments.
			//
			// This makes the logic of p.comment easier, because
			// printing `//` comments is then identical to printing
			// single line `/*` comments.
			p.comment(strings.Split(strings.TrimSpace(s), "\n"))
			if strings.HasSuffix(s, "\n") {
				p.print(newline)
			}

		// Only literals may contain newlines and \t and need to be quoted.
		case n.Kind().IsLiteral():
			p.print(quote(s))

		// All other tokens are printed as is.
		default:
			p.print(s)
		}
		return true
	})

	// Terminate the last line with a newline.
	if !p.firstToken {
		fmt.Fprint(p.w, "\n")
	}

	return p.w.Flush()
}

// comment prints a comment line by line. It removes white-space prefixes from
// multi-line comments, so they can be properly indented.
func (p *CanonicalPrinter) comment(lines []string) {
	line, lines := lines[0], lines[1:]
	p.print(quote(line))

	if len(lines) == 0 {
		return
	}

	prefix := ""
	if l := lines[len(lines)-1]; strings.HasSuffix(l, "*/") {
		l = l[0 : len(l)-2]
		if strings.TrimSpace(l) == "" {
			prefix = l
		}
	}

	for _, line := range lines {
		p.print(newline)
		p.print(" ", quote(strings.TrimPrefix(line, prefix)))
	}
}

func (p *CanonicalPrinter) print(args ...interface{}) {
	for _, arg := range args {
		switch arg := arg.(type) {
		case whiteSpace:

			switch arg {
			case ignore:
				continue
			case indent:
				p.indent++
			case unindent:
				p.indent--
			case blank:
				if p.whiteBuf == "" {
					p.whiteBuf = " "
				}
			case cell:
				if p.whiteBuf == " " {
					p.whiteBuf = ""
				}
				if strings.Count(p.whiteBuf, "\t") == len(p.whiteBuf) {
					p.whiteBuf += "\t"
				}
			case newline:
				if strings.Count(p.whiteBuf, "\n") != len(p.whiteBuf) {
					p.whiteBuf = ""
				}
				p.whiteBuf += "\n"
			}
		default:
			if p.whiteBuf != "" {
				p.printSpace()
				p.whiteBuf = ""
			}
			fmt.Fprint(p.w, arg)
		}

	}
}

// printSpace prints the white-space buffer and indentation.
func (p *CanonicalPrinter) printSpace() {
	if p.firstToken {
		p.firstToken = false
		return
	}
	fmt.Fprint(p.w, p.whiteBuf)
	if strings.HasSuffix(p.whiteBuf, "\n") {
		for i := 0; i < p.indent; i++ {
			fmt.Fprint(p.w, p.Indent)
		}
	}
}

func toBytes(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		return nil, fmt.Errorf("printer: unsupported type %T", v)
	}
}

func quote(s string) string {
	return fmt.Sprintf("\xff%s\xff", s)
}
