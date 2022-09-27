// Package printer implements pretty printers for TTCN-3 source code.
package printer

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/ttcn3/v2/syntax"
)

var std = &printer{}

// Bytes formats src in canonical TTCN-3 style and returns the result or an
// (I/O or syntax) error. src is expected to syntactically correct TTCN-3
// source text.
func Bytes(src []byte) ([]byte, error) {
	return std.Bytes(src)
}

// printer is a simple formatter that only fixes indentation and
// various whitespace issues.
type printer struct {
	buf     bytes.Buffer
	lastPos int
}

func (p *printer) Bytes(src []byte) ([]byte, error) {

	// The simple formatting rules do not need any contextional information
	// so far. This allows us to use the tokenzier and release initial
	// formatting experiments even before the parser is ready.
	tree := syntax.Tokenize(src)

	if tree.Err() != nil {
		return nil, tree.Err()
	}

	if tok := tree.FirstToken(); tok != syntax.Nil {
		p.lastPos = tok.End()
	}
	tree.Inspect(func(n syntax.Node) bool {
		if n == syntax.Nil || !n.IsTerminal() {
			return true
		}
		begin, end := n.Pos(), n.End()
		if begin > p.lastPos {
			p.print(" ")
		}
		p.lastPos = end
		p.print(string(src[begin:end]))
		return true
	})

	return p.buf.Bytes(), nil
}

func (p *printer) print(v interface{}) {
	fmt.Fprint(&p.buf, v)
}
