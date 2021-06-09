package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func generateSymbols(t *testing.T, suite *ntt.Suite) (*ntt.ParseInfo, []protocol.DocumentSymbol) {

	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	syntax := suite.ParseWithAllErrors(name)
	list := lsp.NewAllDefinitionSymbolsFromCurrentModule(syntax)
	ret := make([]protocol.DocumentSymbol, 0, len(list))
	for _, l := range list {
		if l, ok := l.(protocol.DocumentSymbol); ok {
			ret = append(ret, l)
		}
	}

	return syntax, ret
}

func setRange(syntax *ntt.ParseInfo, begin loc.Pos, end loc.Pos) protocol.Range {
	b := syntax.Position(begin)
	e := syntax.Position(end)
	ret := protocol.Range{
		Start: protocol.Position{Line: float64(b.Line - 1), Character: float64(b.Column)},
		End:   protocol.Position{Line: float64(e.Line - 1), Character: float64(e.Column)}}

	return ret
}
func TestFunctionDefWithModuleDotRunsOn(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		function f() runs on TestSystemModuleDotTypes_Module_1.C0 system B0 {}
	  }`, `module TestSystemModuleDotTypes_Module_1
      {
		  type component C0 {}
	  }`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "f", Kind: protocol.Method, Detail: "function definition",
			Range:          setRange(syntax, 49, 119),
			SelectionRange: setRange(syntax, 49, 119),
			Children: []protocol.DocumentSymbol{
				{Name: "runs on", Detail: "TestSystemModuleDotTypes_Module_1.C0", Kind: protocol.Class,
					Range:          setRange(syntax, 70, 106),
					SelectionRange: setRange(syntax, 70, 106)},
				{Name: "system", Detail: "B0", Kind: protocol.Class,
					Range:          setRange(syntax, 114, 116),
					SelectionRange: setRange(syntax, 114, 116)}}}}, list)
}
