package types_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/internal/ttcn3/types"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{"1", types.Typ[types.Integer]},
		{"1.0", types.Typ[types.Float]},
		{"true", types.Typ[types.Boolean]},
		{"false", types.Typ[types.Boolean]},
	}
	for _, tt := range tests {
		testType(t, tt.input, tt.expected)
	}
}

func testType(t *testing.T, input string, expected types.Type) {
	fset := loc.NewFileSet()
	nodes, err := parser.Parse(fset, "<test>", input)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	if len(nodes) == 0 {
		t.Fatalf("no syntax nodes")
	}
	actual, ok := nodes[len(nodes)-1].(ast.Expr)
	if !ok {
		t.Fatalf("expected expression, got %T", nodes[len(nodes)-1])
	}
	info := &types.Info{
		Fset: fset,
	}
	info.CollectInfo(nodes)
	if info.Types[actual] != expected {
		t.Errorf("expected type %v, got %v", expected, info.Types[actual])
	}
}
