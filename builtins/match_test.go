package builtins

import (
	"testing"

	"github.com/nokia/ntt/runtime"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		val runtime.Object
		pat runtime.Object
		exp interface{}
	}{
		// booleans
		{runtime.NewBool(true), runtime.NewBool(true), true},
		{runtime.NewBool(true), runtime.NewBool(false), false},
		{runtime.NewBool(false), runtime.NewBool(false), true},
		{runtime.NewBool(false), runtime.NewBool(true), false},
		{runtime.NewBool(true), runtime.Any, true},
		{runtime.NewBool(false), runtime.Any, true},
		{runtime.NewBool(true), runtime.AnyOrNone, true},
		{runtime.NewBool(false), runtime.AnyOrNone, true},

		// integers
		{runtime.NewInt("1"), runtime.NewInt("2"), false},
		{runtime.NewInt("1"), runtime.NewInt("2000"), false},
		{runtime.NewInt("1"), runtime.NewInt("1"), true},
		{runtime.NewInt("2000"), runtime.NewInt("2000"), true},
		{runtime.NewInt("-1"), runtime.NewInt("-1"), true},
		{runtime.NewInt("-1"), runtime.NewInt("2000"), false},
		{runtime.NewInt("1"), runtime.AnyOrNone, true},
		{runtime.NewInt("-1"), runtime.AnyOrNone, true},
		{runtime.NewInt("1"), runtime.Any, true},
		{runtime.NewInt("-1"), runtime.Any, true},
		{runtime.NewInt("2000"), runtime.Any, true},

		// floats
		{runtime.NewFloat("2.2"), runtime.NewFloat("2.2"), true},
		{runtime.NewFloat("2.2"), runtime.NewFloat("2.5"), false},
		{runtime.NewFloat("2.0"), runtime.NewInt("2"), false},
		{runtime.NewFloat("-2.2"), runtime.NewFloat("2.2"), false},
		{runtime.NewFloat("-2.2"), runtime.NewFloat("-2.2"), true},
		{runtime.NewFloat("2.2"), runtime.AnyOrNone, true},
		{runtime.NewFloat("-2.2"), runtime.AnyOrNone, true},
		{runtime.NewFloat("2.2"), runtime.Any, true},
		{runtime.NewFloat("2e2"), runtime.NewFloat("200"), true},
		{runtime.NewFloat("2e-2"), runtime.NewFloat("0.02"), true},

		// Verdicts
		{runtime.PassVerdict, runtime.PassVerdict, true},
		{runtime.PassVerdict, runtime.FailVerdict, false},
		{runtime.PassVerdict, runtime.ErrorVerdict, false},
		{runtime.PassVerdict, runtime.InconcVerdict, false},
		{runtime.PassVerdict, runtime.NoneVerdict, false},
		{runtime.NoneVerdict, runtime.PassVerdict, false},

		{runtime.FailVerdict, runtime.FailVerdict, true},
		{runtime.FailVerdict, runtime.ErrorVerdict, false},
		{runtime.FailVerdict, runtime.InconcVerdict, false},
		{runtime.FailVerdict, runtime.NoneVerdict, false},

		{runtime.ErrorVerdict, runtime.ErrorVerdict, true},
		{runtime.ErrorVerdict, runtime.InconcVerdict, false},
		{runtime.ErrorVerdict, runtime.NoneVerdict, false},

		{runtime.InconcVerdict, runtime.InconcVerdict, true},
		{runtime.InconcVerdict, runtime.NoneVerdict, false},

		{runtime.NoneVerdict, runtime.NoneVerdict, true},

		{runtime.PassVerdict, runtime.Any, true},
		{runtime.FailVerdict, runtime.Any, true},
		{runtime.ErrorVerdict, runtime.Any, true},
		{runtime.InconcVerdict, runtime.Any, true},
		{runtime.NoneVerdict, runtime.Any, true},
		{runtime.PassVerdict, runtime.AnyOrNone, true},
		{runtime.FailVerdict, runtime.AnyOrNone, true},
		{runtime.ErrorVerdict, runtime.AnyOrNone, true},
		{runtime.InconcVerdict, runtime.AnyOrNone, true},
		{runtime.NoneVerdict, runtime.AnyOrNone, true},

		// records
		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")),
			testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")), true},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")),
			testRecord("f1", runtime.NewInt("4"), "f2", runtime.NewInt("3")), false},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")),
			testRecord("f2", runtime.NewInt("4"), "f1", runtime.NewInt("3")), true},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewBool(false)),
			testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewBool(false)), true},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewBool(false)),
			testRecord("f1", runtime.NewInt("4"), "f2", runtime.NewBool(true)), false},

		{testRecord(), testRecord(), true},
		{testRecord(), testRecord("f1", runtime.NewFloat("2")), false},

		{testRecord("f1", runtime.NewInt("3")), runtime.Any, true},
		{testRecord("f1", runtime.NewInt("3")), runtime.AnyOrNone, true},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")),
			testRecord("f1", runtime.Any, "f2", runtime.Any), true},

		{testRecord("f1", runtime.NewInt("3"), "f2", runtime.NewInt("4")),
			testRecord("f1", runtime.NewInt("3"), "f2", runtime.AnyOrNone), true},

		// set ofs

		{testSetOf(), testSetOf(), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.NewInt("4"), runtime.NewInt("3")), false},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("3")), false},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")), false},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("3"), runtime.AnyOrNone), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.AnyOrNone, runtime.NewInt("3")), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("3")),
			testSetOf(runtime.AnyOrNone, runtime.NewInt("1"), runtime.NewInt("3")), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.NewInt("3"), runtime.AnyOrNone), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.AnyOrNone), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.Any, runtime.NewInt("3")), false},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.Any, runtime.Any), true},

		{testSetOf(runtime.NewInt("1"), runtime.NewInt("2"), runtime.NewInt("3")),
			testSetOf(runtime.NewInt("1"), runtime.Any), false},
	}

	for _, test := range tests {
		got, _ := match(test.val, test.pat)
		if want, ok := test.exp.(bool); ok {
			if want != got {
				t.Errorf("want return value %v, got %v", want, got)
			}
		} else {
			// TODO(5nord) Implement error verification.
			t.Errorf("Error verification not implemented yet. Sorry")
		}
	}
}

func testRecord(a ...interface{}) *runtime.Record {
	r := runtime.NewRecord()
	for i := 0; i < len(a); i += 2 {
		r.Set(a[i].(string), a[i+1].(runtime.Object))
	}
	return r
}

func testSetOf(a ...interface{}) *runtime.List {
	r := runtime.NewSetOf()
	for _, v := range a {
		r.Elements = append(r.Elements, v.(runtime.Object))
	}
	return r
}

func testRecordOf(a ...interface{}) *runtime.List {
	r := runtime.NewRecordOf()
	for _, v := range a {
		r.Elements = append(r.Elements, v.(runtime.Object))
	}
	return r
}
