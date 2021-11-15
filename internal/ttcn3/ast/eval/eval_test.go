package eval_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast/eval"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/runtime"
)

func TestInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"-0", 0},
		{"+0", 0},
		{"10", 10},
		{"-10", -10},
		{"+10", 10},
		{"1+2*3", 7},
		{"(1+2)*3", 9},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		if val == nil {
			t.Errorf("Evaluation of %q returned nil", tt.input)
			continue
		}
		testInt(t, val, tt.expected)
	}

}

func TestBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"not true", false},
		{"not not true", true},
		{"not not not true", false},
		{"not false", true},
		{"not not false", false},
		{"not not not false", true},
		{"1<1", false},
		{"1<=1", true},
		{"1<2", true},
		{"1==1", true},
		{"1==2", false},
		{"1!=1", false},
		{"1!=2", true},
		{"2-1 < 2", true},
		{"2+1==1+2", true},
		{"true==false", false},
		{"true!=false", true},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		if val == nil {
			t.Errorf("Evaluation of %q returned nil", tt.input)
			continue
		}
		testBool(t, val, tt.expected)
	}

}

func TestIfStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testInt(t, val, int64(expected))
		default:
			if val != nil {
				t.Errorf("object is not nil. got=%T (%+v)", val, val)
			}
		}
	}
}

func TestReturnStmt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 1;", 1},
		{"return 2; 9", 2},
		{"return 3*4;9", 12},
		{"9; return 5*6; 9", 30},
		{"if (true) { if (true) { return 7 } return 9 }", 7},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		testInt(t, val, tt.expected)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if(1){}", "boolean expression expected. Got integer (1)"},
		{"-true", "unknown operator: -true"},
		{"true==1", "type mismatch: boolean == integer"},
		{"true+true", "unknown operator: boolean + boolean"},
		{"1&1", "unknown operator: integer & integer"},
		{`"a"+"b"`, "unknown operator: charstring + charstring"},
		{"x", "identifier not found: x"},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		err, ok := val.(*runtime.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", val, val)
			continue
		}
		if err.Message != tt.expected {
			t.Errorf("wrong error message. got=%q, want=%q", err.Message, tt.expected)
		}
	}
}

func TestVars(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var integer x := 5; x", 5},
		{"var integer x := 5*5; x", 25},
		{"var integer x := 23; var integer y := x; y", 23},
		{"var integer x := 8; var integer y := x; var integer z := x+y+3; z", 19},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		testInt(t, val, tt.expected)
	}
}

func TestFunc(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"function f(integer x) { x }; f(5)", 5},
		{"function f(integer x) { return x }; f(6)", 6},
		{"function add(integer x, integer y) { return x+y }; add(1,2)", 3},
		{"function add(integer x, integer y) { return x+y }; add(1,add(2,4))", 7},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		testInt(t, val, tt.expected)
	}
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var integer i; i := 2; i", 2},
		{"var integer i := 2; i := i + 1; i", 3},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		testInt(t, val, tt.expected)
	}
}

func TestLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"while(false){1}; 2", 2},
		{"var integer i := 0; while (i<3) {i := i + 1}; i", 3},
		{"var integer i := 1; do { i := 4 } while (false); i", 4},
		{"var integer i; for (i := 0; i < 3; i := i + 1) {}; i", 3},
		{"var integer x; for (var integer i := 0; i < 3; i := i + 1) {x:=i}; x", 2},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		testInt(t, val, tt.expected)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"Hello Wörld!"`, `Hello Wörld!`},
		{`"Hello" & " " & "World"`, `Hello World`},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		str, ok := val.(*runtime.String)
		if !ok {
			t.Errorf("object is not runtime.String. got=%T (%+v)", val, val)
			continue
		}
		if str.Value != tt.expected {
			t.Errorf("object has wrong value. got=%s, want=%s", str.Value, tt.expected)
		}
	}
}

func testEval(t *testing.T, input string) runtime.Object {
	fset := loc.NewFileSet()
	nodes, err := parser.Parse(fset, "<stdin>", input)
	if err != nil {
		t.Fatalf("testEval: %s", err.Error())
	}
	return eval.Eval(nodes, runtime.NewEnv(nil))
}

func testInt(t *testing.T, obj runtime.Object, expected int64) bool {
	i, ok := obj.(runtime.Int)
	if !ok {
		t.Errorf("object is not runtime.Int. got=%T (%+v)", obj, obj)
		return false
	}

	if !i.IsInt64() {
		t.Errorf("object is to big to compare. got=%s", i)
		return false
	}

	if val := i.Int64(); val != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", val, expected)
		return false
	}

	return true
}

func testBool(t *testing.T, obj runtime.Object, expected bool) bool {
	b, ok := obj.(runtime.Bool)
	if !ok {
		t.Errorf("object is not runtime.Bool. got=%T (%+v)", obj, obj)
		return false
	}

	if val := b.Bool(); val != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", val, expected)
		return false
	}

	return true
}
