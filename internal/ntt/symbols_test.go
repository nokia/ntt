package ntt_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func symbols(t *testing.T, strs ...string) (*ntt.Suite, []*ntt.Module, []error) {
	suite := &ntt.Suite{}
	mods := make([]*ntt.Module, len(strs))
	errs := make([]error, len(strs))

	for i, s := range strs {
		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
		file := suite.File(name)
		file.SetBytes([]byte(s))
		mods[i], errs[i] = suite.Symbols(name)
	}

	return suite, mods, errs
}

// Import handling.
// TODO: func TestImport(t *testing.T) {}
// TODO: func TestGroup(t *testing.T)  {}
// TODO: func TestFriend(t *testing.T) {}

// Types
// TODO: func TestSubType(t *testing.T)       {}
// TODO: func TestPortType(t *testing.T)      {}
// TODO: func TestComponentType(t *testing.T) {}
// TODO: func TestStructType(t *testing.T)    {}
// TODO: func TestEnumType(t *testing.T)      {}
// TODO: func TestBehaviourType(t *testing.T) {}

// Other
// TODO: func TestTemplates(t *testing.T) {}
// TODO: func TestSignature(t *testing.T) {}
// TODO: func TestGoto(t *testing.T)      {}
// TODO: func TestModulePar(t *testing.T) {}
// TODO: func TestFunc(t *testing.T)      {}

func TestDecl(t *testing.T) {
	// TODO(5nord) Check the errors described in the module.
	_, mods, _ := symbols(t, `module Test
	{
	    control {
			var boolean x;         // ERR: Shadowing is not permitted.
			const integer a := b;  // ERR: In local scopes, b must be declared before a.
			const integer b := 23;
			const integer b := 5;  // ERR: Redeclaration of b.
	    }

	    const integer x := y;
	    const integer y := 23;

    }`)

	m := mods[0]
	assert.NotNil(t, m)

	assert.Equal(t, []string{"x", "y"}, m.Names())
	scope, v := m.Lookup("x")
	assert.Equal(t, "x", v.Name())
	assert.Equal(t, "Test", scope.(*ntt.Module).Name())
}
