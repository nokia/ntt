package runtime

import (
	"bytes"

	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type Function struct {
	Params *ast.FormalPars
	Body   *ast.BlockStmt
	Env    *Env
}

func (f *Function) Type() ObjectType { return FUNCTION }
func (f *Function) Inspect() string {
	var buf bytes.Buffer
	buf.WriteString("function(\"")
	for i, p := range f.Params.List {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(p.Name.String())
	}
	buf.WriteString(")")
	return buf.String()
}
