package builtins

import "github.com/nokia/ntt/runtime"

// match returns true given objects match.
func match(a, b runtime.Object) (bool, error) {
	if a == runtime.Any || a == runtime.AnyOrNone || b == runtime.Any || b == runtime.AnyOrNone {
		return true, nil
	}

	if a.Type() != b.Type() {
		return false, runtime.Errorf("type mismatch: %s != %s", a.Type(), b.Type())
	}

	return a.Equal(b), nil
}
