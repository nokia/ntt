package ntt_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
	"github.com/y0ssar1an/q"
)

func buildSuite(t *testing.T, strs ...string) *ntt.Suite {
	suite := &ntt.Suite{}
	for i, s := range strs {
		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
		file := suite.File(name)
		file.SetBytes([]byte(s))
	}
	return suite
}

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
	suite := buildSuite(t, `module Test
	{
	    control {
			var Rec r
			r.x := 23
			r.next.next.x := 5
			r.x := r.y.x
			r.a := 4         // ERR: No such field a.
	    }
	
	    type record Rec {
			integer x,
			Rec     next      // ERR: Recursive field is non-optional
			struct {
				int x
			} y
		}
	
	  }`)

	// Lookup `Rec`
	id, _ := suite.IdentifierAt("TestStructType_Module_0.ttcn3", 4, 8)
	assert.Equal(t, 11, id.Line(id.Def.Pos()))
	assert.Equal(t, 18, id.Column(id.Def.Pos()))

	// Lookup `r` (LHS)
	id, _ = suite.IdentifierAt("TestStructType_Module_0.ttcn3", 7, 11)
	assert.Equal(t, 4, id.Line(id.Def.Pos()))
	assert.Equal(t, 12, id.Column(id.Def.Pos()))

	// Lookup `r` (RHS)
	id, _ = suite.IdentifierAt("TestStructType_Module_0.ttcn3", 8, 4)
	assert.Equal(t, 4, id.Line(id.Def.Pos()))
	assert.Equal(t, 12, id.Column(id.Def.Pos()))

	// Lookup `r.x` (RHS)
	id, _ = suite.IdentifierAt("TestStructType_Module_0.ttcn3", 8, 6)
	q.Q(id)
	assert.Equal(t, 12, id.Line(id.Def.Pos()))
	assert.Equal(t, 12, id.Column(id.Def.Pos()))
}

// Other
// TODO: func TestTemplates(t *testing.T) {}
// TODO: func TestSignature(t *testing.T) {}
// TODO: func TestGoto(t *testing.T)      {}
// TODO: func TestModulePar(t *testing.T) {}
// TODO: func TestFunc(t *testing.T)      {}

func TestDecl(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
	    control {
			var boolean x;         // ERR: Shadowing is not permitted.
			const integer a := b;  // ERR: In local scopes, b must be declared before a.
			const integer b := y;
			const integer b := 5;  // ERR: Redeclaration of b.
	    }
	
	    const integer x := y;
	    const integer y := 23;
	
	}`)

	// Lookup `b`
	id, _ := suite.IdentifierAt("TestDecl_Module_0.ttcn3", 5, 23)
	assert.Equal(t, 6, id.Line(id.Def.Pos()))
	assert.Equal(t, 18, id.Column(id.Def.Pos()))

	// Lookup `y`
	id, _ = suite.IdentifierAt("TestDecl_Module_0.ttcn3", 6, 23)
	assert.Equal(t, 11, id.Line(id.Def.Pos()))
	assert.Equal(t, 20, id.Column(id.Def.Pos()))
}
