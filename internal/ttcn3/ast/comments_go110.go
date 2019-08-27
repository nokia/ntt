// +build go1.10

package ast

import (
	"fmt"
	"strings"

	"github.com/nokia/ntt/internal/ttcn3/token"
)

func joinComments(trivs []Trivia) string {
	var b strings.Builder
	b.Grow(1024)

	for _, triv := range trivs {
		if triv.Kind == token.COMMENT {
			fmt.Fprint(&b, triv.Lit)
		}
	}
	return b.String()
}
