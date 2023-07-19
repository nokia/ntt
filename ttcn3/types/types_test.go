package types_test

import (
	"testing"

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
		{skip: true, Output: "anytype", Type: types.Predefined["anytype"]},
		{Output: "bitstring", Type: types.Predefined["bitstring"]},
		{Output: "boolean", Type: types.Predefined["boolean"]},
		{Output: "charstring", Type: types.Predefined["charstring"]},
		{skip: true, Output: "default", Type: types.Predefined["default"]},
		{Output: "float", Type: types.Predefined["float"]},
		{Output: "hexstring", Type: types.Predefined["hexstring"]},
		{Output: "integer", Type: types.Predefined["integer"]},
		{Output: "octetstring", Type: types.Predefined["octetstring"]},
		{skip: true, Output: "timer", Type: types.Predefined["timer"]},
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

		{skip: true, Output: "record of any",
			Type: &types.ListType{}},

		{skip: true, Output: "record of any",
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

		{skip: true, Output: "any[]",
			Type: &types.ListType{
				Kind: types.Array,
			}},

		{skip: true, Output: "any[1..10]",
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

		{skip: true, Output: "map from any to any",
			Type: &types.ListType{Kind: types.Map}},

		{Output: "map from integer to any",
			Type: &types.ListType{
				Kind:        types.Map,
				ElementType: types.Predefined["integer"],
			}},

		{skip: true, Output: "map from any to any",
			Type: &types.ListType{
				Kind:        types.Map,
				ElementType: &types.PairType{},
			}},

		{Output: "map from integer to charstring",
			Type: &types.ListType{
				Kind: types.Map,
				ElementType: &types.PairType{
					First:  types.Predefined["integer"],
					Second: types.Predefined["charstring"],
				},
			}},

		{Output: "(map from integer to charstring)[]",
			Type: &types.ListType{
				Kind: types.Array,
				ElementType: &types.ListType{
					Kind: types.Map,
					ElementType: &types.PairType{
						First:  types.Predefined["integer"],
						Second: types.Predefined["charstring"],
					},
				},
			}},

		{Output: "map from record of integer[] to map from charstring to (set of float)[]",
			Type: &types.ListType{
				Kind: types.Map,
				ElementType: &types.PairType{
					First: &types.ListType{
						Kind: types.RecordOf,
						ElementType: &types.ListType{
							Kind:        types.Array,
							ElementType: types.Predefined["integer"],
						},
					},
					Second: &types.ListType{
						Kind: types.Map,
						ElementType: &types.PairType{
							First: types.Predefined["charstring"],
							Second: &types.ListType{
								Kind: types.Array,
								ElementType: &types.ListType{
									Kind:        types.SetOf,
									ElementType: types.Float,
								}},
						},
					},
				},
			}},

		// Named types

		{skip: true, Output: "foo [any]",
			Type: &types.PrimitiveType{Name: "foo"}},

		{skip: true, Output: "foo [record of any]",
			Type: &types.ListType{Name: "foo"}},

		{skip: true, Output: "foo [record of integer]",
			Type: &types.ListType{
				Name: "foo",
				ElementType: &types.PrimitiveType{
					Name: "bar",
					Kind: types.Integer,
				},
			}},

		{skip: true, Output: "record of integer",
			Type: &types.ListType{
				ElementType: &types.PrimitiveType{
					Name: "foo",
					Kind: types.Integer,
				},
			}},

		// Structured types

		{skip: true, Output: "record {}", Type: &types.StructuredType{}},
		{skip: true, Output: "record {}", Type: &types.StructuredType{Kind: types.Record}},

		{skip: true, Output: "record {integer, float optional}",
			Type: &types.StructuredType{
				Kind: types.Record,
				Fields: []types.Field{
					{Name: "a", Type: types.Predefined["integer"]},
					{Name: "b", Type: types.Predefined["float"], Optional: true},
				},
			}},

		{skip: true, Output: "set {integer (5) optional} (1..10)",
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

		{skip: true, Output: "T", Type: &types.StructuredType{Kind: types.Component, Name: "T"}},
		{skip: true, Output: "trait", Type: &types.StructuredType{Kind: types.Trait}},
		{skip: true, Output: "C", Type: &types.StructuredType{Kind: types.Component, Name: "C"}},
		{skip: true, Output: "component", Type: &types.StructuredType{Kind: types.Component}},
		{skip: true, Output: "C", Type: &types.StructuredType{Kind: types.Object, Name: "C"}},
		{skip: true, Output: "class", Type: &types.StructuredType{Kind: types.Object}},
		{skip: true, Output: "P", Type: &types.StructuredType{Kind: types.Port, Name: "P"}},
		{skip: true, Output: "port", Type: &types.StructuredType{Kind: types.Port}},

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
