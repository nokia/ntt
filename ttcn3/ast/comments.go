// +build !go1.10

package ast

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/ttcn3/token"
)

func joinComments(trivs []Trivia) string {
	var b bytes.Buffer
	b.Grow(1024)

	for _, triv := range trivs {
		if triv.Kind == token.COMMENT {
			fmt.Fprint(&b, triv.Lit)
		}
	}
	return b.String()
}
