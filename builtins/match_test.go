package builtins

import (
	"fmt"
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
		{Record("f1", 3, "f2", 4), Record("f1", 3, "f2", 4), true},
		{Record("f1", 3, "f2", 4), Record("f1", 4, "f2", 3), false},
		{Record("f1", 3, "f2", 4), Record("f2", 4, "f1", 3), true},
		{Record("f1", 3, "f2", false), Record("f1", 3, "f2", false), true},
		{Record("f1", 3, "f2", false), Record("f1", 4, "f2", true), false},
		{Record(), Record(), true},
		{Record(), Record("f1", 2.0), false},
		{Record("f1", 3), runtime.Any, true},
		{Record("f1", 3), runtime.AnyOrNone, true},
		{Record("f1", 3, "f2", 4), Record("f1", runtime.Any, "f2", runtime.Any), true},
		{Record("f1", 3, "f2", 4), Record("f1", 3, "f2", runtime.AnyOrNone), true},

		// set ofs
		{List(runtime.SET_OF), List(runtime.SET_OF), true},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 1, 2, 3), true},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 1, 4, 3), false},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 3), false},
		{List(runtime.SET_OF, 1, 3), List(runtime.SET_OF, 1, 2, 3), false},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 3, runtime.AnyOrNone), true},
		{List(runtime.SET_OF, 1, 3), List(runtime.SET_OF, 1, runtime.AnyOrNone, 3), true},
		{List(runtime.SET_OF, 1, 3), List(runtime.SET_OF, runtime.AnyOrNone, 1, 3), true},
		{List(runtime.SET_OF, 1, 3), List(runtime.SET_OF, 1, 3, runtime.AnyOrNone), true},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, runtime.AnyOrNone), true},
		{List(runtime.SET_OF, 1, 3), List(runtime.SET_OF, 1, runtime.Any, 3), false},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 1, runtime.Any, runtime.Any), true},
		{List(runtime.SET_OF, 1, 2, 3), List(runtime.SET_OF, 1, runtime.Any), false},

		// record ofs
		{List(runtime.RECORD_OF), List(runtime.RECORD_OF), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, 2, 3), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, 3, 2), false},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, 2), false},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, 2, runtime.Any), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, runtime.Any, 2, 3), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, runtime.Any), false},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, 2, runtime.AnyOrNone), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, 1, runtime.AnyOrNone), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, runtime.AnyOrNone), true},
		{List(runtime.RECORD_OF), List(runtime.RECORD_OF, runtime.AnyOrNone), true},
		{List(runtime.RECORD_OF, 1, 2), List(runtime.RECORD_OF, 1, 2, runtime.AnyOrNone), true},
		{List(runtime.RECORD_OF, 1, 2, 3), List(runtime.RECORD_OF, runtime.AnyOrNone, 1, 2, 3), true},
		{List(runtime.RECORD_OF, 1, 2, 3, 2, 3), List(runtime.RECORD_OF, 1, runtime.AnyOrNone, 3), true},
		{List(runtime.RECORD_OF, 1, 2, 3, 2, 1), List(runtime.RECORD_OF, 1, runtime.AnyOrNone, 3), false},
		{List(runtime.RECORD_OF, 1, 2, 4, 2, 3, 5, 4), List(runtime.RECORD_OF, 1, runtime.AnyOrNone, 2, 3, runtime.AnyOrNone, 4), true},
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

func Record(a ...interface{}) *runtime.Record {
	r := runtime.NewRecord()
	for i := 0; i < len(a); i += 2 {
		r.Set(a[i].(string), makeObj(a[i+1]))
	}
	return r
}

func List(lt runtime.ListType, elems ...interface{}) *runtime.List {
	l := runtime.NewList()
	l.ListType = lt
	for _, e := range elems {
		l.Elements = append(l.Elements, makeObj(e))
	}
	return l
}

func makeObj(v interface{}) runtime.Object {
	switch v := v.(type) {
	case string:
		return runtime.NewString(v)
	case int:
		return runtime.NewInt(fmt.Sprint(v))
	case bool:
		return runtime.NewBool(v)
	case float64:
		return runtime.NewFloat(fmt.Sprint(v))
	default:
		return v.(runtime.Object)
	}
}
