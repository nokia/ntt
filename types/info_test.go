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
	if y.Type.Kind() != types.ArrayType {
		t.Errorf("y.Type is not an array. got=%T", x.Type)
	}

	z := scp.Var("z")
	if z.Type.Kind() != types.TypeReference {
		t.Errorf("z.Type is not a type reference. got=%T", z.Type)
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
	input := `type record R { set { integer x } x }`

	scp, _, _ := makeScope(t, input)
	R := scp.NamedType("R")
	if R.Kind() != types.RecordType {
		t.Fatalf("R is not a record type. got=%T", R)
	}

	x := R.(types.Scope).Lookup("x").(*types.NamedType)
	if x.Type.Kind() != types.SetType {
		t.Fatalf("R.x is not a set type. got=%T", x.Type)
	}
}

func TestUnion(t *testing.T) {
	input := `type union U { integer x }`
	scp, _, _ := makeScope(t, input)
	U := scp.NamedType("U")
	if U.Kind() != types.UnionType {
		t.Fatalf("U is not a struct type. got=%T", U)
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
	A := scp.NamedType("A")
	if A.Kind() != types.EnumeratedType {
		t.Errorf("A is not a enumerated type. got=%T", A)
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
		type record of union { integer x, integer y } r2[23]`

	scp, _, _ := makeScope(t, input)

	x := scp.NamedType("x")
	if x.Kind() != types.ArrayType {
		t.Errorf("x is not an array. got=%T", x)
	}
	assert.Equal(t, types.Integer, x.(*types.List).ElemType)

	r := scp.NamedType("r")
	if r.Kind() != types.RecordType {
		t.Errorf("x is not a record type. got=%T", r)
	}
	assert.Equal(t, []string{"x", "y"}, r.(types.Scope).Names())

	r2 := scp.NamedType("r2")
	if r2.Kind() != types.ArrayType {
		t.Errorf("r2 is not an array. got=%T", r2)
	}
	assert.Equal(t, types.RecordOfType, r2.(*types.List).ElemType.Kind())
	assert.Equal(t, types.UnionType, r2.(*types.List).ElemType.(*types.List).ElemType.Kind())
}

func TestComponents(t *testing.T) {
	t.Skip("Test requires extending component references to be resolved")
	input := `
		type component A {
			var integer x
			var integer y
		}

		type component B extends A {
			var boolean x
		}

		type component C extends A {
			var boolean y
		}

		type component D extends B, C {
		}
`
	scp, _, _ := makeScope(t, input)
	scp.Component("A")
	scp.Component("B")
	scp.Component("C")
	D := scp.Component("D")
	x, ok := D.Lookup("x").(*types.Var)
	if !ok {
		t.Fatalf("D.x is not a var. got=%T", D.Lookup("x"))
	}
	assert.Equal(t, types.Boolean, x.Type)

	y, ok := D.Lookup("y").(*types.Var)
	if !ok {
		t.Fatalf("D.y is not a var. got=%T", y)
	}
	assert.Equal(t, types.Integer, y.Type)
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
