package parser

import (
	"github.com/nokia/ntt/ttcn3/token"
	"testing"
)

func TestModule(t *testing.T) {
	input := `
        // Module foo is empty  
        module foo /* language "ttcn3" */
		{
			private const integer x := 1
		}
	 `

	fset := token.NewFileSet()
	_, err := ParseModule(fset, "<stdin>", input, Trace)
	if err != nil {
		t.Errorf("TestModule: %s", err)
	}
}
