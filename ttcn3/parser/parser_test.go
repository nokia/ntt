package parser

import (
	"testing"
    "github.com/nokia/ntt/ttcn3/token"
    "fmt"
)

func TestModule(t *testing.T) {
	input := `
        // Module foo is empty  
        module foo /* language "ttcn3" */ { }
	 `

    fset := token.NewFileSet()
    m, err := ParseModule(fset, "<stdin>", input, 0)
    if err != nil {
        t.Errorf("TestModule: %s", err)
    }
    fmt.Println(m.Doc)
    fmt.Println(m.Comments)
}
