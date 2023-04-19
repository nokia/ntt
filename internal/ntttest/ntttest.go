// Package ntttest provides utilities for TTCN-3 testing.
package ntttest

import (
	"strings"
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
