package ntt_test

import (
	"testing"
)

//func buildSuite(t *testing.T, strs ...string) *ntt.Suite {
//	suite := &ntt.Suite{}
//	for i, s := range strs {
//		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
//		file := suite.File(name)
//		file.SetBytes([]byte(s))
//	}
//	return suite
//}
//
//func symbols(t *testing.T, strs ...string) (*ntt.Suite, []*ntt.Module) {
//	suite := buildSuite(t, strs...)
//
//	mods := make([]*ntt.Module, len(strs))
//	errs := make([]error, len(strs))
//
//	for i, s := range strs {
//		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
//		mods[i] = suite.symbols(name)
//	}
//
//	return suite, mods
//}

// Import handling.
// TODO: func TestImport(t *testing.T) {}
// TODO: func TestGroup(t *testing.T)  {}
// TODO: func TestFriend(t *testing.T) {}

// Types
// TODO: func TestSubType(t *testing.T)       {}
// TODO: func TestPortType(t *testing.T)      {}
// TODO: func TestComponentType(t *testing.T) {}
// TODO: func TestEnumType(t *testing.T)      {}
// TODO: func TestBehaviourType(t *testing.T) {}
func TestStructType(t *testing.T) {
	//	_, mods, err := symbols(t, `module Test
	//	{
	//	    control {
	//			var Rec r
	//			r.x := 23
	//			r.next.next.x := 5
	//			r.x := r.y.x
	//			r.a := 4         // ERR: No such field a.
	//	    }
	//
	//	    type record Rec {
	//			integer x,
	//			Rec     next      // ERR: Recursive field is non-optional
	//			struct {
	//				int x
	//			} y
	//		}
	//
	//    }`)
	//
	//	m := mods[0]
	//	assert.NotNil(t, m)
	//	q.Q(err)
}

// Other
// TODO: func TestTemplates(t *testing.T) {}
// TODO: func TestSignature(t *testing.T) {}
// TODO: func TestGoto(t *testing.T)      {}
// TODO: func TestModulePar(t *testing.T) {}
// TODO: func TestFunc(t *testing.T)      {}

func TestDecl(t *testing.T) {
	//	suite, mods, _ := symbols(t, `module Test
	//	{
	//	    control {
	//			var boolean x;         // ERR: Shadowing is not permitted.
	//			const integer a := b;  // ERR: In local scopes, b must be declared before a.
	//			const integer b := 23;
	//			const integer b := 5;  // ERR: Redeclaration of b.
	//	    }
	//
	//	    const integer x := y;
	//	    const integer y := 23;
	//
	//    }`)
	//
	//	m := mods[0]
	//	assert.NotNil(t, m)
	//
	//	assert.Equal(t, []string{"x", "y"}, m.Names())
	//	v := m.Lookup("x")
	//	assert.Equal(t, "x", v.Name())
	//	assert.Equal(t, 10, suite.Position("TestDecl_Module_0.ttcn3", v.Pos()).Line)
	//	assert.Equal(t, m, v.Parent())
}
