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
