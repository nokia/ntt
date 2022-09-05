package syntax_test

import (
	"fmt"

	"github.com/nokia/ntt/ttcn3/syntax"
)

func main() {
	root := syntax.Parse([]byte("1+2*3"))
	root.Inspect(func(n syntax.Node) bool {
		fmt.Printf(n.Kind().String())
		return true
	})
}
