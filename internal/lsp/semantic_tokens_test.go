package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func generateTokenList(t *testing.T, suite *ntt.Suite) *protocol.SemanticTokens {

	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	syntax := suite.ParseWithAllErrors(name)
	return lsp.NewSemanticTokensFromCurrentModule(syntax, name)
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

	list := generateTokenList(t, suite)

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

	list := generateTokenList(t, suite)

	assert.Equal(t, list.Data,
		[]uint32{
			0, 0, 6, uint32(lsp.Keyword), 0,
			2, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 5, uint32(lsp.Struct), 0,
			3, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 6, uint32(lsp.Keyword), 0,
			0, 7, 6, uint32(lsp.Keyword), 0,
			0, 13, 2, uint32(lsp.Keyword), 0,
			0, 11, 3, uint32(lsp.Struct), 0,
			1, 8, 4, uint32(lsp.Keyword), 0,
			0, 5, 3, uint32(lsp.Keyword), 0,
			0, 4, 5, uint32(lsp.Struct), 0,
			3, 8, 8, uint32(lsp.Keyword), 0,
			0, 9, 1, uint32(lsp.Function), 0,
			0, 4, 6, uint32(lsp.Keyword), 0})
}
