package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/nokia/ntt/ttcn3/syntax"
)

var (
	nolintRegex = regexp.MustCompile(`^[/*\s]*NOLINT\(([^\)]+)\)[/*\r\n\s]*$`)
)

type errPattern struct {
	node syntax.Node
	msg  string
}

func (e errPattern) Error() string {
	return fmt.Sprintf("%s:%s: error: %s", syntax.Filename(e.node), syntax.Begin(e.node), e.msg)
}

func (e errPattern) IsSilent() bool { return isSilent(e.node, "TemplateDef") }

type errLines struct {
	node  syntax.Node
	lines int
}

func (e errLines) Error() string {
	return fmt.Sprintf("%s:%s: error: %q must not have more than %d lines (%d)",
		syntax.Filename(e.node), syntax.Begin(e.node), syntax.Name(e.node), style.MaxLines, e.lines)
}

func (e errLines) IsSilent() bool { return isSilent(e.node, "CodeStatistics.TooLong") }

type errBraces struct {
	left, right syntax.Node
}

func (e errBraces) Error() string {
	return fmt.Sprintf("%s:%s: error: braces must be in the same line or same column",
		syntax.Filename(e.right), syntax.Begin(e.right))
}

type errComplexity struct {
	node       syntax.Node
	complexity int
}

func (e errComplexity) Error() string {
	return fmt.Sprintf("%s:%s: error: cyclomatic complexity of %q (%d) must not be higher than %d",
		syntax.Filename(e.node), syntax.Begin(e.node), syntax.Name(e.node), e.complexity, style.Complexity.Max)
}

func (e errComplexity) IsSilent() bool { return isSilent(e.node, "CodeStatistics.TooComplex") }

type errMissingCaseElse struct {
	node syntax.Node
}

func (e errMissingCaseElse) Error() string {
	return fmt.Sprintf("%s:%s: error: missing case else in select statement", syntax.Filename(e.node), syntax.Begin(e.node))
}

type errUsageExceedsLimit struct {
	node  syntax.Node
	usage int
	limit int
	text  string
}

func (e errUsageExceedsLimit) Error() string {
	return fmt.Sprintf("%s:%s: error: %q must not be used more than %d times. %s",
		syntax.Filename(e.node), syntax.Begin(e.node), syntax.Name(e.node), e.limit, e.text)
}

type errUnusedModule struct {
	file string
	name string
}

func (e errUnusedModule) Error() string {
	return fmt.Sprintf("%s: error: unused module %q", e.file, e.name)
}

func isSilent(n syntax.Node, checks ...string) bool {
	scanner := bufio.NewScanner(strings.NewReader(syntax.Doc(n)))
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
