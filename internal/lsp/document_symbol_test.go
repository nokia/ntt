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
        type component B0 extends C0, C1 {
			var integer i := 1;
			timer t1 := 2.0;
			port P p;
		}
		function f() runs on TestFunctionDefWithModuleDotRunsOn_Module_1.C0 system B0 {}
	  }`, `module TestFunctionDefWithModuleDotRunsOn_Module_1
      {
		  type component C0 {}
		  type component C1 {}
	  }`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "B0", Kind: protocol.Class, Detail: "component type",
			Range:          setRange(syntax, 26, 120),
			SelectionRange: setRange(syntax, 26, 120),
			Children: []protocol.DocumentSymbol{
				{Name: "extends", Kind: protocol.Array,
					Range:          setRange(syntax, 52, 58),
					SelectionRange: setRange(syntax, 52, 58),
					Children: []protocol.DocumentSymbol{
						{Name: "C0", Kind: protocol.Class,
							Range:          setRange(syntax, 52, 54),
							SelectionRange: setRange(syntax, 52, 54)},
						{Name: "C1", Kind: protocol.Class,
							Range:          setRange(syntax, 56, 58),
							SelectionRange: setRange(syntax, 56, 58)}}},
				{Name: "i", Detail: "var integer", Kind: protocol.Variable,
					Range:          setRange(syntax, 64, 82),
					SelectionRange: setRange(syntax, 64, 82)},
				{Name: "t1", Detail: "timer", Kind: protocol.Event,
					Range:          setRange(syntax, 87, 102),
					SelectionRange: setRange(syntax, 87, 102)},
				{Name: "p", Detail: "port P", Kind: protocol.Interface,
					Range:          setRange(syntax, 107, 115),
					SelectionRange: setRange(syntax, 107, 115)}}},
		{Name: "f", Kind: protocol.Method, Detail: "function definition",
			Range:          setRange(syntax, 123, 203),
			SelectionRange: setRange(syntax, 123, 203),
			Children: []protocol.DocumentSymbol{
				{Name: "runs on", Detail: "TestFunctionDefWithModuleDotRunsOn_Module_1.C0", Kind: protocol.Class,
					Range:          setRange(syntax, 144, 190),
					SelectionRange: setRange(syntax, 144, 190)},
				{Name: "system", Detail: "B0", Kind: protocol.Class,
					Range:          setRange(syntax, 198, 200),
					SelectionRange: setRange(syntax, 198, 200)}}}}, list)
}

func TestRecordOfTypeDefWithTypeRef(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type integer Byte(0..255)
		type record of Byte Octets
	  }`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "Byte", Kind: protocol.Struct, Detail: "subtype",
			Range:          setRange(syntax, 26, 51),
			SelectionRange: setRange(syntax, 26, 51),
			Children:       nil},
		{Name: "Octets", Kind: protocol.Array, Detail: "record of type",
			Range:          setRange(syntax, 54, 80),
			SelectionRange: setRange(syntax, 54, 80),
			Children: []protocol.DocumentSymbol{
				{Name: "Byte", Detail: "element type", Kind: protocol.Struct,
					Range:          setRange(syntax, 69, 73),
					SelectionRange: setRange(syntax, 69, 73)}}}}, list)
}

func TestConstTemplModulePar(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        const R c_r := {10, true}
		modulepar R m_r := {10, true}
		template R t_r1 := ?;
		template R t_r2 := {10, true}
		template R t_r3(integer pi:=10) := {pi, true}
		template R t_r4 extends t_r1:= {f2:= true}
	  }`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "c_r", Kind: protocol.Constant, Detail: "const",
			Range:          setRange(syntax, 26, 51),
			SelectionRange: setRange(syntax, 26, 51),
			Children:       nil},
		{Name: "m_r", Kind: protocol.Constant, Detail: "modulepar",
			Range:          setRange(syntax, 54, 83),
			SelectionRange: setRange(syntax, 54, 83),
			Children:       nil},
		{Name: "t_r1", Kind: protocol.Constant, Detail: "template",
			Range:          setRange(syntax, 86, 106),
			SelectionRange: setRange(syntax, 86, 106),
			Children:       nil},
		{Name: "t_r2", Kind: protocol.Constant, Detail: "template",
			Range:          setRange(syntax, 110, 139),
			SelectionRange: setRange(syntax, 110, 139),
			Children:       nil},
		{Name: "t_r3", Kind: protocol.Constant, Detail: "template",
			Range:          setRange(syntax, 142, 187),
			SelectionRange: setRange(syntax, 142, 187),
			Children:       nil},
		{Name: "t_r4", Kind: protocol.Constant, Detail: "template",
			Range:          setRange(syntax, 190, 232),
			SelectionRange: setRange(syntax, 190, 232),
			Children: []protocol.DocumentSymbol{
				{Name: "t_r1", Kind: protocol.Constant, Detail: "template",
					Range:          setRange(syntax, 214, 218),
					SelectionRange: setRange(syntax, 214, 218),
					Children:       nil}}}}, list)
}
