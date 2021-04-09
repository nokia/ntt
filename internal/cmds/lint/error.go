package lint

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

var (
	nolintRegex = regexp.MustCompile(`^[/*\s]*NOLINT\(([^\)]+)\)[/*\r\n\s]*$`)
)

type errPattern struct {
	fset *loc.FileSet
	node ast.Node
	msg  string
}

func (e errPattern) Error() string {
	return fmt.Sprintf("%s: error: %s", e.fset.Position(e.node.Pos()), e.msg)
}

func (e errPattern) IsSilent() bool { return isSilent(e.node, "TemplateDef") }

type errLines struct {
	fset  *loc.FileSet
	node  ast.Node
	lines int
}

func (e errLines) Error() string {
	return fmt.Sprintf("%s: error: %q must not have more than %d lines (%d)",
		e.fset.Position(e.node.Pos()), identName(e.node), style.MaxLines, e.lines)
}

func (e errLines) IsSilent() bool { return isSilent(e.node, "CodeStatistics.TooLong") }

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

func (e errComplexity) IsSilent() bool { return isSilent(e.node, "CodeStatistics.TooComplex") }

type errMissingCaseElse struct {
	fset *loc.FileSet
	node ast.Node
}

func (e errMissingCaseElse) Error() string {
	return fmt.Sprintf("%s: error: missing case else in select statement", e.fset.Position(e.node.Pos()))
}

type errUsageExceedsLimit struct {
	fset  *loc.FileSet
	node  ast.Node
	usage int
	limit int
	text  string
}

func (e errUsageExceedsLimit) Error() string {
	return fmt.Sprintf("%s: error: %q must not be used more than %d times. %s",
		e.fset.Position(e.node.Pos()), identName(e.node), e.limit, e.text)
}

type errUnusedModule struct {
	file string
}

func (e errUnusedModule) Error() string {
	return fmt.Sprintf("%s: error: unused module", e.file)
}

func isSilent(n ast.Node, checks ...string) bool {
	scanner := bufio.NewScanner(strings.NewReader(ast.FirstToken(n).Comments()))
	for scanner.Scan() {
		if s := nolintRegex.FindStringSubmatch(scanner.Text()); len(s) == 2 {
			for _, s := range strings.Split(s[1], ",") {
				if searchString(checks, s) {
					return true
				}
			}
		}
	}
	return false
}
func searchString(slice []string, s string) bool {
	for _, s2 := range slice {
		if strings.TrimSpace(s) == strings.TrimSpace(s2) {
			return true
		}
	}
	return false
}
