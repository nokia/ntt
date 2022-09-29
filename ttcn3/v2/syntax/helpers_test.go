package syntax

import "strings"

const CURSOR = "¶"

// ExtractCursor removes the cursor from the given string and returns the
// string and the cursor position. If the cursor is not found, the cursor
// position is -1.
func ExtractCursor(s string) (string, int) {
	return strings.Replace(s, CURSOR, "", -1), strings.Index(s, CURSOR)
}
