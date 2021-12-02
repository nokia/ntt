package types_test

import (
	"testing"

	"github.com/nokia/ntt/types"
	"github.com/stretchr/testify/assert"
)

func TestValueDecl(t *testing.T) {
	input := `const integer x := y, y[x] := z; const x z`

	scp, _, err := makeScope(t, input)
	if err != nil {
		t.Fatal(err)
	}

	x := scp.Var("x")
	if x.Type != types.Integer {
		t.Errorf("x.Type is not integer. got=%T", x.Type)
	}

	z := scp.Var("z")
	if _, ok := z.Type.(*types.Ref); !ok {
		t.Errorf("z.Type is not integer. got=%T", z.Type)
	}
}

func TestTemplateDecl(t *testing.T) {
	input := `var template integer x := y, y[x] := z; template x z(integer p) := p`

	scp, _, err := makeScope(t, input)
	if err != nil {
		t.Fatal(err)
	}

	x := scp.Var("x")
	if x.Type != types.Integer {
		t.Errorf("x.Type is not integer. got=%T", x.Type)
	}

	scp.Var("z")
}

func TestModule(t *testing.T) {
	input := `
		// A sningle TTCN-3 input is allowed to have multiple modules.
		module m  { const integer x }
		module m2 { const integer x }

		// A single module is allowed to be defined multiple times. For
		// example, a directory with multiple TTCN-3 files define a
		// single module.
		module m  { const integer x, y }

		// This is not allowed.
		modulepar integer m3 := 23;
		module m3 { const ineteger x}`

	scp, _, _ := makeScope(t, input)

	m := scp.Module("m")
	assert.Equal(t, m.Names(), []string{"x", "y"})

	m2 := scp.Module("m2")
	assert.Equal(t, m2.Names(), []string{"x"})

	m3 := scp.Lookup("m3")
	if _, ok := m3.(*types.Var); !ok {
		t.Errorf("m3 is not a variable. got=%T", m3)
	}

}
