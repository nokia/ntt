package lint

import (
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type errNaming struct {
	fset *loc.FileSet
	node ast.Node
	msg  string
}

func (e errNaming) Error() string {
	return fmt.Sprintf("%s: error: %s", e.fset.Position(e.node.Pos()), e.msg)
}

type errLines struct {
	fset  *loc.FileSet
	node  ast.Node
	lines int
}

func (e errLines) Error() string {
	return fmt.Sprintf("%s: error: %q must not have more than %d lines (%d)",
		e.fset.Position(e.node.Pos()), identName(e.node), style.MaxLines, e.lines)
}

type errBraces struct {
	fset        *loc.FileSet
	left, right ast.Node
}

func (e errBraces) Error() string {
	return fmt.Sprintf("%s: error: braces must be in the same line or same column",
		e.fset.Position(e.right.Pos()))
}

type errComplexity struct {
	fset       *loc.FileSet
	node       ast.Node
	complexity int
}

func (e errComplexity) Error() string {
	return fmt.Sprintf("%s: error: cyclomatic complexity of %q (%d) must not be higher than %d",
		e.fset.Position(e.node.Pos()), identName(e.node), e.complexity, style.Complexity.Max)
}
