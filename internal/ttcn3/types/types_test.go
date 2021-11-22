package types_test

import (
	"bytes"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/internal/ttcn3/types"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"1", types.Typ[types.Integer]},
		{"1.0", types.Typ[types.Float]},
		{"true", types.Typ[types.Boolean]},
		{"false", types.Typ[types.Boolean]},
		{"+1", types.Typ[types.Integer]},
		{"-1", types.Typ[types.Integer]},
		{"not true", types.Typ[types.Boolean]},
		{"- true", `invalid type "boolean", expected "integer"`},
	}
	for _, tt := range tests {
		testType(t, tt.input, tt.expected)
	}
}

func testType(t *testing.T, input string, expected interface{}) {
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

	errs := &multierror.Error{}
	errs.ErrorFormat = func(errs []error) string {
		var buf bytes.Buffer
		for i, err := range errs {
			if i != 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}

	info := &types.Info{
		Fset: fset,
		Error: func(err error) {
			errs = multierror.Append(errs, err)
		},
	}
	info.CollectInfo(nodes)
	switch expected := expected.(type) {
	case types.Type:
		if errs.ErrorOrNil() != nil {
			t.Errorf("unexpected error: %s", errs)
		}
		if info.Types[actual] != expected {
			t.Errorf("expected type %v, got %v", expected, info.Types[actual])
		}

	case string:
		if errs.ErrorOrNil() == nil || errs.Error() != expected {
			t.Errorf("expected error %q, got %q", expected, errs)
		}
	default:
		t.Fatalf("unexpected expected type %T", expected)
	}
}
