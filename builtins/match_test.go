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
		{runtime.NewFloat("-2.2"), runtime.Any, true},
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
