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

	y := scp.Var("y")
	if _, ok := y.Type.(*types.List); !ok {
		t.Errorf("y.Type is not an array. got=%T", x.Type)
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

func TestSubType(t *testing.T) {
	input := `type integer int; var int x`

	scp, _, _ := makeScope(t, input)
	assert.Equal(t, types.Integer, scp.NamedType("int"))
}

func TestRecord(t *testing.T) {
	input := `type record R { record { integer x } x }`

	scp, _, _ := makeScope(t, input)
	typ := scp.Type("R").(*types.NamedType)
	R, ok := typ.Type.(*types.Struct)
	if !ok {
		t.Fatalf("R is not a struct type. got=%T", typ)
	}

	x := R.Lookup("x").(*types.NamedType)
	if _, ok := x.Type.(*types.Struct); !ok {
		t.Fatalf("R.x is not a struct type. got=%T", x.Type)
	}
}

func TestEnum(t *testing.T) {
	input := `
		type enumerated A { E1, E2, E3, E4 }
		type enumerated B { E1, E2, E3, E4 }

		type record   E1 { integer E2 }
		const E1      E2 := { E2 := 23 }
		const integer E3 := 23;
		const A       E4 := E1; // not allowed, because constant is of type A`
	scp, _, _ := makeScope(t, input)
	if _, ok := scp.NamedType("A").(*types.Struct); !ok {
		t.Errorf("A is not a enumerated type. got=%T", scp.NamedType("A"))
	}
}

func TestScopes(t *testing.T) {
	_ = []struct {
		input  string
		reject bool
	}{
		// Standards states that every identifier must be unique in scope hierarchy.
		// All bellow tests should be rejected, actually.
		{input: `type enumerated X { X }`},
		{input: `type record X { integer X }`},
		{input: `type record X { record { integer X } x }`},
		{input: `type record X { enumerated { X } x }`},
		{input: `type integer X; type enumerated A { X }`, reject: false}, // allowed by standard example.
		{input: `type integer X; type record A { integer X }`},
		{input: `type integer X; function A() { var integer X }`, reject: true}, // forbidden by standard
		{input: `type integer X; function A(integer X) { }`},
	}
}

func TestNestedTypes(t *testing.T) {
	input := `
		type integer x[-1]
		type record { integer x, integer y } r
		type record of record { integer x, integer y } r2[23]`

	scp, _, _ := makeScope(t, input)

	typ := scp.Type("x").(*types.NamedType)
	if l, ok := typ.Type.(*types.List); ok {
		assert.Equal(t, types.Integer, l.ElemType)
	} else {
		t.Errorf("x is not a list type. got=%T", typ)
	}

	typ = scp.Type("r").(*types.NamedType)
	if s, ok := typ.Type.(*types.Struct); ok {
		assert.Equal(t, []string{"x", "y"}, s.Names())
	} else {
		t.Errorf("x is not a struct type. got=%T", typ)
	}

	typ = scp.Type("r2").(*types.NamedType)
	if l, ok := typ.Type.(*types.List); ok {
		if _, ok := l.ElemType.(*types.List); !ok {
			t.Errorf("element type of r2 is not a list type. got=%T", l.ElemType)
		}
	} else {
		t.Errorf("x is not a list type. got=%T", typ)
	}

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
	assert.Equal(t, []string{"x", "y"}, m.Names())

	m2 := scp.Module("m2")
	assert.Equal(t, []string{"x"}, m2.Names())

	m3 := scp.Lookup("m3")
	if _, ok := m3.(*types.Var); !ok {
		t.Errorf("m3 is not a variable. got=%T", m3)
	}

}
