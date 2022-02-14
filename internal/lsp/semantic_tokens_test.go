package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/ttcn3"
	"github.com/stretchr/testify/assert"
)

func SetRange(startLine uint32, startCol uint32, endLine uint32, endCol uint32) *protocol.Range {
	return &protocol.Range{
		Start: protocol.Position{Line: startLine, Character: startCol},
		End:   protocol.Position{Line: endLine, Character: endCol}}
}
func DefaultRange(tree *ttcn3.Tree) *protocol.Range {
	modEnd := tree.Position(tree.Root.End())
	return &protocol.Range{Start: protocol.Position{Line: 0, Character: 0}, End: protocol.Position{Line: uint32(modEnd.Line - 1), Character: uint32(modEnd.Column - 1)}}
}
func generateTokenList(t *testing.T, suite *ntt.Suite, rng *protocol.Range) *protocol.SemanticTokens {
	lsp.EmitKeywords = true
	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	tree := ttcn3.ParseFile(name)

	// index
	db := &ttcn3.DB{}
	if srcs, _ := suite.Sources(); len(srcs) > 0 {
		for _, src := range srcs {
			db.Index(src)
		}
	}

	if rng == nil {
		rng = DefaultRange(tree)
	}

	return lsp.NewSemanticTokensFromCurrentModule(tree, db, suite, name, *rng)
}

func TestFullModuleKwOnly(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 extends C0, C1 {
			var integer i := 1;
			timer t1 := 2.0;
			port P p;
		}
		function f() runs on TestFunctionDefWithModuleDotRunsOn_Module_0.C0 system B0 return integer {}
	}`)

	list := generateTokenList(t, suite, nil)

	assert.NotEqual(t, len(list.Data), 0)
}

func TestFullModuleKwTypeId(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type record MyRec {
			integer i
		}
        type record length(0..2) of integer RoI;
        type set MySet {
			integer i
		}
        function f() return integer {}
	  }`)

	list := generateTokenList(t, suite, nil)

	assert.Equal(t,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Namespace), uint32(lsp.Definition),
			2, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 5, uint32(lsp.Struct), uint32(lsp.Definition),
			1, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			2, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 6, uint32(lsp.Keyword), 0,
			0, 13, 2, uint32(lsp.Keyword), 0,
			0, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 3, uint32(lsp.Type), uint32(lsp.Definition),
			1, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 3, uint32(lsp.Keyword), 0,
			0, 4, 5, uint32(lsp.Struct), uint32(lsp.Definition),
			1, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			2, 8, 8, uint32(lsp.Keyword), 0,
			0, 9, 1, uint32(lsp.Function), uint32(lsp.Definition),
			0, 4, 6, uint32(lsp.Keyword), 0,
			0, 7, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary)}, list.Data)
}

func TestConstAndTemplDecl(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type integer X;
		type X Y;
		type record MyRec {
			integer i
		}
        const float pi := 3.14;
		template MyRec a_rec1 := *;
		template (value) MyRec a_rec2(template integer p_i) := {
			i := p_i
		}
 	  }`)

	list := generateTokenList(t, suite, nil)

	assert.Equal(t,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Namespace), uint32(lsp.Definition),
			2, 2, 4, uint32(lsp.Keyword), 0,
			0, 5, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 1, uint32(lsp.Type), uint32(lsp.Definition),
			1, 2, 4, uint32(lsp.Keyword), 0,
			0, 5, 1, uint32(lsp.Type), 0,
			0, 2, 1, uint32(lsp.Type), uint32(lsp.Definition),
			1, 2, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 5, uint32(lsp.Struct), uint32(lsp.Definition),
			1, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			2, 8, 5, uint32(lsp.Keyword), 0,
			0, 6, 5, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 6, 2, uint32(lsp.Variable), uint32(lsp.Declaration | lsp.Readonly),
			1, 2, 8, uint32(lsp.Keyword), 0,
			0, 9, 5, uint32(lsp.Type), uint32(lsp.Undefined),
			0, 6, 6, uint32(lsp.Variable), uint32(lsp.Declaration | lsp.Readonly),
			1, 2, 8, uint32(lsp.Keyword), 0,
			0, 10, 5, uint32(lsp.Keyword), 0,
			0, 7, 5, uint32(lsp.Type), uint32(lsp.Undefined),
			0, 6, 6, uint32(lsp.Variable), uint32(lsp.Declaration | lsp.Readonly),
			0, 7, 8, uint32(lsp.Keyword), 0,
			0, 9, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 3, uint32(lsp.Parameter), uint32(lsp.Declaration),
			1, 8, 3, uint32(lsp.Parameter), 0,
		}, list.Data)
}

func TestModuleIds(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		import from TestModuleIds_Module_1 all;
		type TestModuleIds_Module_1.Byte MyByte;
		function f(TestModuleIds_Module_1.Byte pb) {
			var TestModuleIds_Module_1.MyRec r;
			r.i := pb;
		}
	}`,
		`module TestModuleIds_Module_1 {
		type integer Byte;
		type record MyRec {integer i};
	}`)

	list := generateTokenList(t, suite, nil)

	assert.Equal(t,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Namespace), uint32(lsp.Definition),
			2, 2, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Keyword), 0,
			0, 5, 22, uint32(lsp.Namespace), 0,
			0, 23, 3, uint32(lsp.Keyword), 0, //all
			1, 2, 4, uint32(lsp.Keyword), 0, // type line 3
			0, 5, 22, uint32(lsp.Namespace), 0, //
			0, 23, 4, uint32(lsp.Type), 0,
			0, 5, 6, uint32(lsp.Type), uint32(lsp.Definition),
			1, 2, 8, uint32(lsp.Keyword), 0, // line 4
			0, 9, 1, uint32(lsp.Function), uint32(lsp.Definition),
			0, 2, 22, uint32(lsp.Namespace), 0,
			0, 23, 4, uint32(lsp.Type), 0,
			0, 5, 2, uint32(lsp.Parameter), uint32(lsp.Declaration), // pb
			1, 3, 3, uint32(lsp.Keyword), 0, //var line 5
			0, 4, 22, uint32(lsp.Namespace), 0,
			0, 23, 5, uint32(lsp.Type), uint32(lsp.Undefined),
			0, 6, 1, uint32(lsp.Variable), uint32(lsp.Declaration),
			1, 10, 2, uint32(lsp.Parameter), 0,
		}, list.Data)
}

func TestDeltaModuleKwTypeId(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type record MyRec {
			integer i
		}
        type record length(0..2) of integer RoI;
        type set MySet {
			integer i
		}
        function f() return integer {}
	  }`)

	list := generateTokenList(t, suite, SetRange(5, 8, 8, 3))

	assert.Equal(t,
		[]uint32{
			0, 0, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 6, uint32(lsp.Keyword), 0,
			0, 13, 2, uint32(lsp.Keyword), 0,
			0, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 3, uint32(lsp.Type), uint32(lsp.Definition),
			1, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 3, uint32(lsp.Keyword), 0,
			0, 4, 5, uint32(lsp.Struct), uint32(lsp.Definition),
			1, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary)}, list.Data)
}

func TestUniversalCharstring(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type record MyRec {
			universal charstring  uch,
			charstring ch
		}
		type universal charstring UcharT;
		function f() return universal charstring {}
	  }`)

	list := generateTokenList(t, suite, nil)

	assert.Equal(t,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Namespace), uint32(lsp.Definition),
			2, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 5, uint32(lsp.Struct), uint32(lsp.Definition),
			1, 3, 20, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			1, 3, 10, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			2, 2, 4, uint32(lsp.Keyword), 0,
			0, 5, 20, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 21, 6, uint32(lsp.Type), uint32(lsp.Definition),
			1, 2, 8, uint32(lsp.Keyword), 0,
			0, 9, 1, uint32(lsp.Function), uint32(lsp.Definition),
			0, 4, 6, uint32(lsp.Keyword), 0,
			0, 7, 20, uint32(lsp.Type), uint32(lsp.DefaultLibrary)}, list.Data)
}

func TestPortTypeDeclSemTok(t *testing.T) {
	suite := buildSuite(t, `module Test
{
	import from TestPortTypeDeclSemTok_Module_1 all;
	type charstring AddrType;
	type boolean Type1;
	type port MyPort message {
		in integer, universal charstring
		out Type1, TestPortTypeDeclSemTok_Module_1.Type2
		map param (in integer pi, boolean pb);
		address AddrType
	}
}`, `module TestPortTypeDeclSemTok_Module_1
{
	type integer Type2;
}`)

	list := generateTokenList(t, suite, nil)

	assert.Equal(t,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Namespace), uint32(lsp.Definition),
			2, 1, 6, uint32(lsp.Keyword), 0,
			0, 7, 4, uint32(lsp.Keyword), 0,
			0, 5, 31, uint32(lsp.Namespace), 0,
			0, 32, 3, uint32(lsp.Keyword), 0,
			1, 1, 4, uint32(lsp.Keyword), 0,
			0, 5, 10, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 11, 8, uint32(lsp.Type), uint32(lsp.Definition),
			1, 1, 4, uint32(lsp.Keyword), 0,
			0, 5, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 5, uint32(lsp.Type), uint32(lsp.Definition),
			1, 1, 4, uint32(lsp.Keyword), 0,
			0, 5, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Interface), uint32(lsp.Definition),
			0, 7, 7, uint32(lsp.Keyword), 0,
			1, 2, 2, uint32(lsp.Keyword), 0,
			0, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 9, 20, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			1, 2, 3, uint32(lsp.Keyword), 0,
			0, 4, 5, uint32(lsp.Type), 0,
			0, 7, 31, uint32(lsp.Namespace), 0,
			0, 32, 5, uint32(lsp.Type), 0,
			1, 2, 3, uint32(lsp.Keyword), 0,
			0, 4, 5, uint32(lsp.Keyword), 0,
			0, 7, 2, uint32(lsp.Keyword), 0,
			0, 3, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 2, uint32(lsp.Parameter), uint32(lsp.Declaration),
			0, 4, 7, uint32(lsp.Type), uint32(lsp.DefaultLibrary),
			0, 8, 2, uint32(lsp.Parameter), uint32(lsp.Declaration),
			1, 2, 7, uint32(lsp.Keyword), 0,
			0, 8, 8, uint32(lsp.Type), 0}, list.Data)
}
