package types_test

import (
	"testing"

	"github.com/nokia/ntt/internal/ntttest"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/nokia/ntt/ttcn3/types"
	"github.com/stretchr/testify/assert"
)

// Kind strings are used to construct error messages and should therefore have
// a tested string representation.
func TestKindStrings(t *testing.T) {
	assert.Equal(t, "any", types.Kind(0).String(), "any is the zero value")
	assert.Equal(t, "integer", types.Integer.String())
	assert.Equal(t, "boolean", types.Boolean.String())
	assert.Equal(t, "unknown", types.Kind(1<<63).String(), "unknown types should have a string representation")
	assert.Equal(t, "boolean", (types.Boolean | types.Kind(1<<63)).String(), "only known types should be represented")
	assert.Equal(t, "error or boolean or integer", (types.Boolean | types.Integer | types.Error).String(), "multiple root types should be concatenated")
}

func TestTypeStrings(t *testing.T) {
	tree := ttcn3.Parse("5, 1..10")
	if tree.Err != nil {
		t.Fatal(tree.Err)
	}

	intVal := tree.Nodes[0].(syntax.Expr)
	rangeVal := tree.Nodes[1].(syntax.Expr)

	tests := []struct {
		Type   types.Type
		Output string
		skip   bool
	}{

		// Predefined types

		{Output: "any", Type: types.Predefined["any"]},
		{Output: "anytype", Type: types.Predefined["anytype"]},
		{Output: "bitstring", Type: types.Predefined["bitstring"]},
		{Output: "boolean", Type: types.Predefined["boolean"]},
		{Output: "charstring", Type: types.Predefined["charstring"]},
		{skip: true, Output: "default", Type: types.Predefined["default"]},
		{Output: "float", Type: types.Predefined["float"]},
		{Output: "hexstring", Type: types.Predefined["hexstring"]},
		{Output: "integer", Type: types.Predefined["integer"]},
		{Output: "octetstring", Type: types.Predefined["octetstring"]},
		{Output: "timer", Type: types.Predefined["timer"]},
		{Output: "universal charstring", Type: types.Predefined["universal charstring"]},

		// Primitive types

		{Output: "any",
			Type: &types.PrimitiveType{}},

		{Output: "any ()",
			Type: &types.PrimitiveType{
				ValueConstraints: []types.Value{},
			}},

		{Output: "any (5, 1..10)",
			Type: &types.PrimitiveType{
				ValueConstraints: []types.Value{
					{Expr: intVal},
					{Expr: rangeVal},
				},
			}},

		{Output: "integer",
			Type: &types.PrimitiveType{
				Kind: types.Integer,
			}},

		{Output: "integer (5)",
			Type: &types.PrimitiveType{
				Kind:             types.Integer,
				ValueConstraints: []types.Value{{Expr: intVal}},
			}},

		// List types

		{Output: "record of any",
			Type: &types.ListType{}},

		{Output: "record of any",
			Type: &types.ListType{Kind: types.RecordOf}},

		{Output: "hexstring",
			Type: &types.ListType{Kind: types.Hexstring}},

		{Output: "set of integer",
			Type: &types.ListType{
				Kind:        types.SetOf,
				ElementType: types.Predefined["integer"],
			}},

		{Output: "set of record of integer",
			Type: &types.ListType{
				Kind: types.SetOf,
				ElementType: &types.ListType{
					Kind:        types.RecordOf,
					ElementType: types.Predefined["integer"],
				},
			}},

		{Output: "record of charstring length(5)",
			Type: &types.ListType{
				Kind: types.RecordOf,
				ElementType: &types.ListType{
					Kind:             types.Charstring,
					LengthConstraint: types.Value{Expr: intVal},
				},
			}},

		{Output: "record length(1..10) of charstring",
			Type: &types.ListType{
				Kind:             types.RecordOf,
				LengthConstraint: types.Value{Expr: rangeVal},
				ElementType: &types.ListType{
					Kind: types.Charstring,
				},
			}},

		{Output: "record length(1..10) of charstring length(5)",
			Type: &types.ListType{
				Kind:             types.RecordOf,
				LengthConstraint: types.Value{Expr: rangeVal},
				ElementType: &types.ListType{
					Kind:             types.Charstring,
					LengthConstraint: types.Value{Expr: intVal},
				},
			}},

		// Arrays

		{Output: "any[]",
			Type: &types.ListType{
				Kind: types.Array,
			}},

		{Output: "any[1..10]",
			Type: &types.ListType{
				Kind:             types.Array,
				LengthConstraint: types.Value{Expr: rangeVal},
			}},

		{Output: "integer[5][1..10]",
			Type: &types.ListType{
				Kind:             types.Array,
				LengthConstraint: types.Value{Expr: rangeVal},
				ElementType: &types.ListType{
					Kind:             types.Array,
					LengthConstraint: types.Value{Expr: intVal},
					ElementType:      types.Predefined["integer"],
				},
			}},

		{Output: "integer[]",
			Type: &types.ListType{
				Kind:        types.Array,
				ElementType: types.Predefined["integer"],
			}},

		{Output: "record of integer[]", // record of integer-arrays
			Type: &types.ListType{
				Kind: types.RecordOf,
				ElementType: &types.ListType{
					Kind: types.Array,
					ElementType: &types.PrimitiveType{
						Kind: types.Integer},
				},
			}},

		{Output: "(record of integer)[]", // array of integer-record-ofs
			Type: &types.ListType{
				Kind: types.Array,
				ElementType: &types.ListType{
					Kind: types.RecordOf,
					ElementType: &types.PrimitiveType{
						Kind: types.Integer},
				},
			}},

		{Output: "map from any to any",
			Type: &types.MapType{}},

		{Output: "map from integer to any",
			Type: &types.MapType{
				From: types.Predefined["integer"],
			}},

		{Output: "map from integer to charstring",
			Type: &types.MapType{
				From: types.Predefined["integer"],
				To:   types.Predefined["charstring"],
			}},

		{Output: "(map from integer to charstring)[]",
			Type: &types.ListType{
				Kind: types.Array,
				ElementType: &types.MapType{
					From: types.Predefined["integer"],
					To:   types.Predefined["charstring"],
				},
			}},

		{Output: "map from record of integer[] to map from charstring to (set of float)[]",
			Type: &types.MapType{
				From: &types.ListType{
					Kind: types.RecordOf,
					ElementType: &types.ListType{
						Kind:        types.Array,
						ElementType: types.Predefined["integer"],
					},
				},
				To: &types.MapType{
					From: types.Predefined["charstring"],
					To: &types.ListType{
						Kind: types.Array,
						ElementType: &types.ListType{
							Kind:        types.SetOf,
							ElementType: types.Predefined["float"],
						},
					},
				},
			}},

		// Named types

		{Output: "foo [any]",
			Type: &types.PrimitiveType{Name: "foo"}},

		{Output: "foo [record of any]",
			Type: &types.ListType{Name: "foo"}},

		{Output: "foo [record of integer]",
			Type: &types.ListType{
				Name: "foo",
				ElementType: &types.PrimitiveType{
					Name: "bar",
					Kind: types.Integer,
				},
			}},

		{Output: "record of integer",
			Type: &types.ListType{
				ElementType: &types.PrimitiveType{
					Name: "foo",
					Kind: types.Integer,
				},
			}},

		// Structured types

		{Output: "record {}", Type: &types.StructuredType{}},
		{Output: "record {}", Type: &types.StructuredType{Kind: types.Record}},

		{Output: "record {integer, float optional}",
			Type: &types.StructuredType{
				Kind: types.Record,
				Fields: []types.Field{
					{Name: "a", Type: types.Predefined["integer"]},
					{Name: "b", Type: types.Predefined["float"], Optional: true},
				},
			}},

		{Output: "set {integer (5) optional} (1..10)",
			Type: &types.StructuredType{
				Kind: types.Set,
				Fields: []types.Field{{
					Name: "a",
					Type: &types.PrimitiveType{
						Kind:             types.Integer,
						ValueConstraints: []types.Value{{Expr: intVal}}},
					Optional: true,
				}},
				ValueConstraints: []types.Value{{Expr: rangeVal}},
			}},

		{Output: "T", Type: &types.StructuredType{Kind: types.Trait, Name: "T"}},
		{Output: "trait", Type: &types.StructuredType{Kind: types.Trait}},
		{Output: "C", Type: &types.StructuredType{Kind: types.Component, Name: "C"}},
		{Output: "component", Type: &types.StructuredType{Kind: types.Component}},
		{Output: "C", Type: &types.StructuredType{Kind: types.Object, Name: "C"}},
		{Output: "class", Type: &types.StructuredType{Kind: types.Object}},
		{Output: "P", Type: &types.StructuredType{Kind: types.Port, Name: "P"}},
		{Output: "port", Type: &types.StructuredType{Kind: types.Port}},

		// TODO(5nord) add tests for enumerated types, unions and subtypes
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Output, func(t *testing.T) {
			t.Parallel()
			if tt.skip {
				t.Skip()
			}
			assert.Equal(t, tt.Output, tt.Type.String())
		})
	}
}

func TestTypeInference(t *testing.T) {
	tests := []struct {
		input  string
		expect string
		skip   bool
	}{
		// Identifiers
		{skip: true, input: `integer`, expect: `integer`},
		{skip: true, input: `float`, expect: `float`},
		{skip: true, input: `boolean`, expect: `boolean`},

		// ValueLiterals
		{input: `0`, expect: `integer`},
		{input: `0.0`, expect: `float`},
		{skip: true, input: `infinity`, expect: `float`},
		{input: `not_a_number`, expect: `float`},
		{input: `true`, expect: `boolean`},
		{input: `false`, expect: `boolean`},
		{input: `"hello"`, expect: `charstring`},
		{input: `"wörld"`, expect: `universal charstring`},
		{input: `'111'H`, expect: `hexstring`},
		{input: `'111'B`, expect: `bitstring`},
		{input: `'111'O`, expect: `octetstring`},
		{input: `pass`, expect: `verdicttype`},

		// Unary Expressions
		{input: `+0`, expect: `integer`},
		{input: `-0`, expect: `integer`},
		{input: `not4b '111'B`, expect: `bitstring`},

		// Binary Expressions
		{input: `1+2`, expect: `integer`},
		{input: `1+2-3`, expect: `integer`},
		{input: `1.0+2.0`, expect: `float`},
		{input: `not true or false`, expect: `boolean`},
		{input: `1+2 <= 3`, expect: `boolean`},

		{skip: true, input: `{1}`, expect: `record of integer`},
		{skip: true, input: `{1,2,3}`, expect: `record of integer`},
		{skip: true, input: `{1+2,2,3}`, expect: `record of integer`},
		{skip: true, input: `{(1+2),2,(3)}`, expect: `record of integer`},
		{skip: true, input: `{{1},{2},{}}`, expect: `record of record of integer`},
		{skip: true, input: `{"A",""}`, expect: `record of charstring`},
		{skip: true, input: `{"A","Ä"}`, expect: `record of universal charstring`},
		{skip: true, input: `{x := 2, y := true}`, expect: `record { integer, boolean }`},
		{skip: true, input: `{x := 2, y := omit}`, expect: `record { integer, boolean optional}`},
		{skip: true, input: `{[1] := 1, [3] := 2, [2] := 3}`, expect: `record of integer`},
		{skip: true, input: `{[1.0] := 1, [3.0] := 2, [2.0] := 3}`, expect: `map from float to integer`},

		//{skip: true, input: `x := 2`, expect: `void`},
		//{skip: true, input: `mtc == null`, expect: `boolean`},
		//{skip: true, input: `null == null`, expect: `boolean`},
		//{skip: true, input: `all component`, expect: `component`},
		//{skip: true, input: `__MODULE__`, expect: `charstring`},
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			if tt.skip {
				t.Skip()
			}
			input, cursor := ntttest.CutCursor(tt.input)
			tree := ttcn3.Parse(input)
			if err := tree.Err; err != nil {
				t.Fatal(err)
			}
			if len(tree.Nodes) < 1 {
				t.Fatal("no nodes")
			}
			var expr syntax.Expr
			if cursor >= 0 {
				pos := tree.Position(cursor)
				expr = tree.ExprAt(pos.Line, pos.Column)
			} else {
				switch n := tree.Nodes[len(tree.Nodes)-1].(type) {
				case syntax.Expr:
					expr = n
				case *syntax.ExprStmt:
					expr = n.Expr
				default:
					t.Fatalf("last statement is not an expression: %T", n)
				}
			}

			typ := types.TypeOf(expr)
			if typ == nil {
				t.Fatal("type is nil")
			}
			assert.Equal(t, tt.expect, typ.String())

		})
	}

}
