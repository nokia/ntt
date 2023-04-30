package syntax

import (
	"fmt"
)

// Error represents a syntax error.
type Error struct {
	Node
	Msg string
}

func (e Error) Error() string {
	if spn := SpanOf(e.Node); spn.Begin.IsValid() {
		return fmt.Sprintf("%s: %s", spn, e.Msg)
	}

	return e.Msg
}
