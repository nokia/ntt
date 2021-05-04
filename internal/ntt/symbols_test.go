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
	id, _ := suite.DefinitionAt(file, line, column)
	if id == nil || id.Def == nil {
		return Pos{}
	}
	return Pos{
		Line:   id.Def.Position.Line,
		Column: id.Def.Position.Column,
	}
}

// Import handling.
// TODO: func TestImport(t *testing.T) {}
// TODO: func TestGroup(t *testing.T)  {}
// TODO: func TestFriend(t *testing.T) {}

func TestPortType(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		type record Msg{}
		type record Msg2{}

		type port P1 message {
			in charstring
			out Msg, Msg2
		}

		type port P4 message {
			in Msg
			map param (Msg p1)
		}

		type port P5 message {
			in Msg
			map param (int Msg)
		}

		type port P6 mixed { }
		type port P7 procedure { }

		// It's a common mistake to forget the port kind (message, procedure or
		// mixed). To go defintion should work despite that.
		type port P8 {
			inout Msg
		}
	  }`)

	// Lookup `Msg`
	def := gotoDefinition(suite, "TestPortType_Module_0.ttcn3", 8, 8)
	assert.Equal(t, Pos{Line: 3, Column: 15}, def)

	// Lookup `Msg2`
	def = gotoDefinition(suite, "TestPortType_Module_0.ttcn3", 8, 13)
	assert.Equal(t, Pos{Line: 4, Column: 15}, def)

	// Lookup `Msg`
	def = gotoDefinition(suite, "TestPortType_Module_0.ttcn3", 13, 15)
	assert.Equal(t, Pos{Line: 3, Column: 15}, def)

	// Lookup `Msg`
	def = gotoDefinition(suite, "TestPortType_Module_0.ttcn3", 18, 19)
	assert.Equal(t, Pos{Line: 0, Column: 0}, def)
}

func TestComponentType(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
        type component CompA {
           integer compA_var1;
        }

        type component CompB extends CompA {
           integer comp_var2
        }

        type component CompC extends CompA {
           bool comp_var2
        }

        type component CompD extends CompB, CompC {
           integer compD_var3
        }

        type component CompE<type T, C> extends C {
           integer compE_var1
		   T       compE_var2
        }

		type component CompF extends CompE<integer> {}

		function f1() runs on CompE {}

	}`)

	// Lookup `CompA`
	def := gotoDefinition(suite, "TestComponentType_Module_0.ttcn3", 7, 39)
	assert.Equal(t, Pos{Line: 3, Column: 24}, def)

	// Lookup `CompC`
	def = gotoDefinition(suite, "TestComponentType_Module_0.ttcn3", 15, 45)
	assert.Equal(t, Pos{Line: 11, Column: 24}, def)

	// Lookup `C`
	// TODO(5nord) Fix parser bug
	//def = gotoDefinition(suite, "TestComponentType_Module_0.ttcn3", 19, 49)
	//assert.Equal(t, Pos{Line: 19, Column: 32}, def)

	// Lookup `CompE`
	//def = gotoDefinition(suite, "TestComponentType_Module_0.ttcn3", 26, 25)
	//assert.Equal(t, Pos{Line: 19, Column: 9}, def)
}

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
			Rec     next,      // ERR: Recursive field is non-optional
			record {
				int x
			} y
		}

	  }`)

	// Lookup `Rec`
	def := gotoDefinition(suite, "TestStructType_Module_0.ttcn3", 4, 8)
	assert.Equal(t, Pos{Line: 11, Column: 18}, def)

	// Lookup `r` (LHS)
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

func TestEnumType(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		type enumerated RGB {
			RED,
			GREEN,
			BLUE
		}

		type enumerated Colors {
			BLACK,
			RED,
			ORANGE,
			YELLOW,
			GREEN,
			BLUE,
			PURPLE,
			WHITE
		}

		function f1(RGB p1) {}
		function f2(Colors p1) {}

	    control {
			var Colors vc := YELLOW;
			f1(GREEN);
			f2(GREEN);
			log(GREEN);
	    }
	}`)

	// Lookup `Rec`
	def := gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 20, 15)
	assert.Equal(t, Pos{Line: 3, Column: 19}, def)

	// in the middle of 'Colors'
	def = gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 24, 9)
	assert.Equal(t, Pos{Line: 9, Column: 19}, def)

	def = gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 24, 23) //YELLOW
	assert.Equal(t, Pos{Line: 13, Column: 4}, def)

	def = gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 25, 7) //RGB::GREEN
	assert.Equal(t, Pos{Line: 5, Column: 4}, def)                     //

	def = gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 26, 7) //Colors::GREEN
	//assert.Equal(t, Pos{Line: 12, Column: 4}, def)                  // TODO(5nord) resolve ambiguities
	assert.Equal(t, Pos{Line: 5, Column: 4}, def) //

	def = gotoDefinition(suite, "TestEnumType_Module_0.ttcn3", 27, 8) //GREEN
	assert.Equal(t, Pos{Line: 5, Column: 4}, def)                     // ambiguous: should show 2 defs
}

// Other
func TestTemplates(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		type integer int;

		control {
			template int t0 := 5;
			template int t1(int x) modifies t0 :=  23;
			var template int x;
		}
	}`)

	// Lookup `int`
	def := gotoDefinition(suite, "TestTemplates_Module_0.ttcn3", 6, 13)
	assert.Equal(t, Pos{Line: 3, Column: 16}, def)

	// Lookup `t0`
	def = gotoDefinition(suite, "TestTemplates_Module_0.ttcn3", 7, 36)
	assert.Equal(t, Pos{Line: 6, Column: 4}, def)
}

// TODO: func TestSignature(t *testing.T) {}
// TODO: func TestGoto(t *testing.T)      {}
// TODO: func TestModulePar(t *testing.T) {}

func TestFunc(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		function f1(integer x := y) {
			f2(x)
		}

		function f2() {
			x := 19;
			f2(f1())
		}

		const integer y := 23
	}`)

	// Lookup `f2`
	def := gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 4, 4)
	assert.Equal(t, Pos{Line: 7, Column: 3}, def)

	// Lookup `f2` recursivly
	def = gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 9, 4)
	assert.Equal(t, Pos{Line: 7, Column: 3}, def)

	// Lookup `f1`
	def = gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 9, 7)
	assert.Equal(t, Pos{Line: 3, Column: 3}, def)

	// Lookup unknown `x`
	def = gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 8, 4)
	assert.Equal(t, Pos{Line: 0, Column: 0}, def)

	// Lookup `x` parameter.
	def = gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 4, 7)
	assert.Equal(t, Pos{Line: 3, Column: 23}, def)

	// Lookup `y` constant.
	def = gotoDefinition(suite, "TestFunc_Module_0.ttcn3", 3, 28)
	assert.Equal(t, Pos{Line: 12, Column: 17}, def)
}

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
