package printer

import (
	"fmt"
	"io"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type whiteSpace byte

const (
	ignore   = whiteSpace(0)
	blank    = whiteSpace(' ')
	vtab     = whiteSpace('\v')
	newline  = whiteSpace('\n')
	formfeed = whiteSpace('\f')
	indent   = whiteSpace('>')
	unindent = whiteSpace('<')
)

func Print(w io.Writer, fset *loc.FileSet, n ast.Node) error {
	p := printer{w: w, fset: fset}
	p.print(n)
	return p.err
}

type printer struct {
	w      io.Writer
	fset   *loc.FileSet
	indent int
	err    error
}

func (p *printer) print(v interface{}) {
	switch n := v.(type) {
	case whiteSpace:
		switch n {
		case indent:
			p.indent++
		case unindent:
			p.indent--
		default:
			fmt.Fprint(p.w, n)
		}
	case ast.Node:
		ast.Apply(n, p.enter, p.exit)
	default:
		panic(fmt.Sprintf("type ast.Node expected, but %+v has type %T", v, v))
	}
}

func (p *printer) enter(c *ast.Cursor) bool {
	return true
}

func (p *printer) exit(c *ast.Cursor) bool {
	return true
}
