package interpreter_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/interpreter"
	"github.com/nokia/ntt/runtime"
	"github.com/nokia/ntt/ttcn3/parser"
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
		{"goto L10", "goto statement not implemented"},
		{"break", "break or continue statements not allowed outside loops"},
		{"continue", "break or continue statements not allowed outside loops"},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		err, ok := val.(*runtime.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", val, val)
			continue
		}
		if err.Error() != tt.expected {
			t.Errorf("wrong error message. got=%q, want=%q", err.Error(), tt.expected)
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
		{"var integer i := 2; while(false){i:=1};i", 2},
		{"var integer i := 0; while (i<3) {i := i + 1}; i", 3},
		{"var integer i := 1; do { i := 4 } while (false); i", 4},
		{"var integer i; for (i := 0; i < 3; i := i + 1) {}; i", 3},
		{"var integer x; for (var integer i := 0; i < 3; i := i + 1) {x:=i}; x", 2},
		{"var integer i := 5; while(true) { break; i := 2}; i", 5},
		{"var integer i := 1; while(true) { while(true) { break; i := 2} i:= 6; break}; i", 6},
		{"var integer x := 7; for (var integer i := 0; i< 3; i:= i + 1) {continue; x:=i}; x", 7},
		{"var integer x := 9; do { continue } while(false); x", 9},
		{"var integer i; for ( i:= 0; true; i := i + 1) {break}; i", 0},
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

func TestBitstring(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"'011'b", "'11'B"},
		{"not4b '01'b", "'10'B"},
		{"'0011'b and4b '0101'b", "'1'B"},
		{"'0011'b or4b  '0101'b", "'111'B"},
		{"'0011'b xor4b '0101'b", "'110'B"},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		str, ok := val.(*runtime.Bitstring)
		if !ok {
			t.Errorf("object is not runtime.Bitstring. got=%T (%+v)", val, val)
			continue
		}
		if str.Inspect() != tt.expected {
			t.Errorf("object has wrong value. got=%s, want=%s", str.Inspect(), tt.expected)
		}
	}
}

func TestBuiltinFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`lengthof("")`, 0},
		{`lengthof("fnord")`, 5},
		{`lengthof(1)`, "integer arguments not supported"},
		{`lengthof("hello", "world")`, "wrong number of arguments. got=2, want=1"},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testInt(t, val, int64(expected))
		case string:
			err, ok := val.(*runtime.Error)
			if !ok {
				t.Errorf("object is not runtime.Error. got=%T (%+v)", val, val)
				continue
			}
			if err.Error() != expected {
				t.Errorf("wrong error message. got=%q, want=%s", err.Error(), expected)
			}
		}
	}
}

func TestList(t *testing.T) {
	input := "var integer a[3] := {1, 1+1, 3}; a"
	val := testEval(t, input)
	l, ok := val.(*runtime.List)
	if !ok {
		t.Errorf("object is not runtime.List. got=%T (%+v)", val, val)
		return
	}
	testInt(t, l.Elements[0], 1)
	testInt(t, l.Elements[1], 2)
	testInt(t, l.Elements[2], 3)
}

func TestIndexExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"var integer a[3] := {1, 1+1, 3}; a[0] + a[1] + a[2]", 6},
		{"var integer a[3] := {1, 1+1, 3}; a[3]", nil},
		{"var integer a[3] := {1, 1+1, 3}; a[-1]", nil},
		{"var integer a[3] := {1, 1+1, 3}; var integer i := 2; a[i]", 3},
		{"var integer x := {2,4,8}[1]; x", 4},
		{`var integer m := { ["foo"] := 23, [ 1+2 ] := 5}; m["foo"] + m[3]`, 28},
		{`var integer r := { x := 2, y := 3}; r.x+r.y`, 5},
	}
	for _, tt := range tests {
		val := testEval(t, tt.input)
		expected, ok := tt.expected.(int)
		if ok {
			testInt(t, val, int64(expected))
		} else {
			if val != runtime.Undefined {
				t.Errorf("object is not undefined. got=%T (%+v)", val, val)
			}
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`1==1`, true},
		{`2==1`, false},
		{`"one"=="one"`, true},
		{`"one"=="two"`, false},
		{`true==true`, true},
		{`false==false`, true},
		{`var RoI a:={1}, b:={1}; a==b`, true},
		{`var RoI a:={1,2}, b:={2,1}; a==b`, false},
		{`var RoI a:={}, b:={2,1}; a==b`, false},
		{`var RoI a:={}, b:={}; a==b`, true},
	}

	for _, tt := range tests {
		val := testEval(t, tt.input)
		testBool(t, val, tt.expected)
	}
}

func TestMapExpr(t *testing.T) {
	input := `var Map m := { ["foo"] := 1, ["bar"] := 2 }; m`
	val := testEval(t, input)
	m, ok := val.(*runtime.Map)
	if !ok {
		t.Errorf("object is not runtime.Map. got=%T (%+v)", val, val)
		return
	}

	tests := []struct {
		key      runtime.Object
		expected int64
	}{
		{&runtime.String{Value: "foo"}, 1},
		{&runtime.String{Value: "bar"}, 2},
	}
	for _, tt := range tests {
		val, ok := m.Get(tt.key)
		if !ok {
			t.Errorf("key %s not found", tt.key.(runtime.Object).Inspect())
			continue
		}
		testInt(t, val, tt.expected)
	}
}

func testEval(t *testing.T, input string) runtime.Object {
	fset := loc.NewFileSet()
	nodes, _, _, err := parser.Parse(fset, "<stdin>", input)
	if err != nil {
		t.Fatalf("%s\n %s", input, err.Error())
	}
	return interpreter.Eval(nodes, runtime.NewEnv(nil))
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
