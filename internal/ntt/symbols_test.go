package ntt_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
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

type Pos struct {
	Line   int
	Column int
}

func gotoDefinition(suite *ntt.Suite, file string, line, column int) Pos {
	id, _ := suite.IdentifierAt(file, line, column)
	if id == nil || id.Def == nil {
		return Pos{}
	}
	return Pos{
		Line:   id.Line(id.Def.Pos()),
		Column: id.Column(id.Def.Pos()),
	}
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
	def := gotoDefinition(suite, "TestStructType_Module_0.ttcn3", 4, 8)
	assert.Equal(t, Pos{Line: 11, Column: 18}, def)

	//// Lookup `r` (LHS)
	def = gotoDefinition(suite, "TestStructType_Module_0.ttcn3", 7, 11)
	assert.Equal(t, Pos{Line: 4, Column: 12}, def)

	//// Lookup `r` (RHS)
	def = gotoDefinition(suite, "TestStructType_Module_0.ttcn3", 8, 4)
	assert.Equal(t, Pos{Line: 4, Column: 12}, def)

	//// Lookup `r.x` (RHS)
	// TODO(5nord) Type resolution not implemented yet.
	//def = gotoDefinition(suite, "TestStructType_Module_0.ttcn3", 8, 6)
	//assert.Equal(t, Pos{Line: 12, Column: 12}, def)
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
	def := gotoDefinition(suite, "TestDecl_Module_0.ttcn3", 5, 23)
	assert.Equal(t, Pos{Line: 6, Column: 18}, def)

	//// Lookup `y`
	def = gotoDefinition(suite, "TestDecl_Module_0.ttcn3", 6, 23)
	assert.Equal(t, Pos{Line: 11, Column: 20}, def)
}
