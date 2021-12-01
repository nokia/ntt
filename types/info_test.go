package types_test

import (
	"testing"

	"github.com/nokia/ntt/types"
)

func TestValueDecl(t *testing.T) {
	input := `const integer x := y, y[x] := z; const x z`

	scp, _, err := makeScope(t, input)
	if err != nil {
		t.Fatal(err)
	}

	x, ok := scp.Lookup("x").(*types.Var)
	if !ok {
		t.Fatalf("x is not a variable. got=%T", scp.Lookup("x"))
	}
	if x.Type != types.Integer {
		t.Errorf("x.Type is not integer. got=%T", x.Type)
	}

	z, ok := scp.Lookup("z").(*types.Var)
	if !ok {
		t.Fatalf("z is not a variable. got=%T", scp.Lookup("z"))
	}

	if _, ok := z.Type.(*types.Ref); !ok {
		t.Errorf("z.Type is not integer. got=%T", z.Type)
	}
}
