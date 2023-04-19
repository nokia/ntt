// Package ntttest provides utilities for TTCN-3 testing.
package ntttest

import (
	"fmt"
	"strings"

	"github.com/nokia/ntt/ttcn3/syntax"
)

// CURSOR is a special character (pilcrow "¶") that is used to mark the cursor position in
// the input string. See function ExtractCursor for more details.
//
// The pilcrow can be entered in the source code by pressing ^Kpp in Vim.
const CURSOR = "¶"

// CutCursor returns s without the first instance of the cursor character ("¶")
// and its position as second return value. If there's no cursor in the input
// string, the position is -1.
func CutCursor(s string) (string, int) {
	pos := strings.Index(s, CURSOR)
	s = strings.Replace(s, CURSOR, "", 1)
	return s, pos
}

// NodeString returns a string representation of the given node.
func NodeString(n syntax.Node) string {
	s := fmt.Sprintf("%T", n)
	if n := syntax.Name(n); n != "" {
		s += fmt.Sprintf("(%s)", n)
	}
	return s
}
