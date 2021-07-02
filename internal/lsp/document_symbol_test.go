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
		function f() runs on TestFunctionDefWithModuleDotRunsOn_Module_1.C0 system B0 return integer {}
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
			Range:          setRange(syntax, 123, 218),
			SelectionRange: setRange(syntax, 123, 218),
			Children: []protocol.DocumentSymbol{
				{Name: "runs on", Detail: "TestFunctionDefWithModuleDotRunsOn_Module_1.C0", Kind: protocol.Class,
					Range:          setRange(syntax, 144, 190),
					SelectionRange: setRange(syntax, 144, 190)},
				{Name: "system", Detail: "B0", Kind: protocol.Class,
					Range:          setRange(syntax, 198, 200),
					SelectionRange: setRange(syntax, 198, 200)},
				{Name: "return", Detail: "integer", Kind: protocol.Struct,
					Range:          setRange(syntax, 208, 215),
					SelectionRange: setRange(syntax, 208, 215)}}}}, list)
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
		template R t_r4 modifies t_r1 := {f2:= true}
	  }`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "c_r", Kind: protocol.Constant, Detail: "const R",
			Range:          setRange(syntax, 26, 51),
			SelectionRange: setRange(syntax, 26, 51),
			Children:       nil},
		{Name: "m_r", Kind: protocol.Constant, Detail: "modulepar R",
			Range:          setRange(syntax, 54, 83),
			SelectionRange: setRange(syntax, 54, 83),
			Children:       nil},
		{Name: "t_r1", Kind: protocol.Constant, Detail: "template R",
			Range:          setRange(syntax, 86, 106),
			SelectionRange: setRange(syntax, 86, 106),
			Children:       nil},
		{Name: "t_r2", Kind: protocol.Constant, Detail: "template R",
			Range:          setRange(syntax, 110, 139),
			SelectionRange: setRange(syntax, 110, 139),
			Children:       nil},
		{Name: "t_r3", Kind: protocol.Constant, Detail: "template R",
			Range:          setRange(syntax, 142, 187),
			SelectionRange: setRange(syntax, 142, 187),
			Children:       nil},
		{Name: "t_r4", Kind: protocol.Constant, Detail: "template R",
			Range:          setRange(syntax, 190, 234),
			SelectionRange: setRange(syntax, 190, 234),
			Children: []protocol.DocumentSymbol{
				{Name: "t_r1", Kind: protocol.Constant, Detail: "template",
					Range:          setRange(syntax, 215, 219),
					SelectionRange: setRange(syntax, 215, 219),
					Children:       nil}}}}, list)
}

func TestPortTypeDecl(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type port Pmessage message{
			address AddressType;
			in aModule.Msg1, integer, Msg2;
			out Msg3, Msg4;
			inout Msg5;
			map param(PmessageMapType1 p1, PmessageMapType2 p2[]);
			unmap(PmessageUnmapType1 p1);
		}
	  }`)

	syntax, list := generateSymbols(t, suite)
	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "Pmessage", Kind: protocol.Interface, Detail: "message port type",
			Range:          setRange(syntax, 20, 235),
			SelectionRange: setRange(syntax, 20, 235),
			Children: []protocol.DocumentSymbol{
				{Name: "address", Kind: protocol.Struct, Detail: "AddressType type",
					Range:          setRange(syntax, 51, 70),
					SelectionRange: setRange(syntax, 51, 70),
					Children:       nil},
				{Name: "in", Kind: protocol.Array,
					Range:          setRange(syntax, 75, 105),
					SelectionRange: setRange(syntax, 75, 105),
					Children: []protocol.DocumentSymbol{
						{Name: "aModule.Msg1", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 78, 90),
							SelectionRange: setRange(syntax, 78, 90),
							Children:       nil},
						{Name: "integer", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 92, 99),
							SelectionRange: setRange(syntax, 92, 99),
							Children:       nil},
						{Name: "Msg2", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 101, 105),
							SelectionRange: setRange(syntax, 101, 105),
							Children:       nil}}},
				{Name: "out", Kind: protocol.Array,
					Range:          setRange(syntax, 110, 124),
					SelectionRange: setRange(syntax, 110, 124),
					Children: []protocol.DocumentSymbol{
						{Name: "Msg3", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 114, 118),
							SelectionRange: setRange(syntax, 114, 118),
							Children:       nil},
						{Name: "Msg4", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 120, 124),
							SelectionRange: setRange(syntax, 120, 124),
							Children:       nil}}},
				{Name: "inout", Kind: protocol.Array,
					Range:          setRange(syntax, 129, 139),
					SelectionRange: setRange(syntax, 129, 139),
					Children: []protocol.DocumentSymbol{{Name: "Msg5", Kind: protocol.Struct, Detail: "type",
						Range:          setRange(syntax, 135, 139),
						SelectionRange: setRange(syntax, 135, 139),
						Children:       nil}}}}}}, list)
}

func TestSignatureDecl(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		signature MyRemoteProcOne ();
		signature MyRemoteProcTwo () noblock;
		signature MyRemoteProcThree (in integer Par1, out float Par2, inout integer Par3);
		signature MyRemoteProcFour (in integer Par1) return integer;
		signature MyRemoteProcFive (inout float Par1) return integer
			exception (ExceptionType1, ExceptionType2);
		signature MyRemoteProcSix (in integer Par1) noblock
			exception (integer, float);
	}`)

	syntax, list := generateSymbols(t, suite)

	assert.Equal(t, []protocol.DocumentSymbol{
		{Name: "MyRemoteProcOne", Kind: protocol.Function, Detail: "blocking signature",
			Range:          setRange(syntax, 20, 48),
			SelectionRange: setRange(syntax, 20, 48),
			Children:       nil},
		{Name: "MyRemoteProcTwo", Kind: protocol.Function, Detail: "non-blocking signature",
			Range:          setRange(syntax, 52, 88),
			SelectionRange: setRange(syntax, 52, 88),
			Children:       nil},
		{Name: "MyRemoteProcThree", Kind: protocol.Function, Detail: "blocking signature",
			Range:          setRange(syntax, 92, 173),
			SelectionRange: setRange(syntax, 92, 173),
			Children:       nil},
		{Name: "MyRemoteProcFour", Kind: protocol.Function, Detail: "blocking signature",
			Range:          setRange(syntax, 177, 236),
			SelectionRange: setRange(syntax, 177, 236),
			Children: []protocol.DocumentSymbol{
				{Name: "integer", Kind: protocol.Struct, Detail: "return type",
					Range:          setRange(syntax, 229, 236),
					SelectionRange: setRange(syntax, 229, 236),
					Children:       nil}}},
		{Name: "MyRemoteProcFive", Kind: protocol.Function, Detail: "blocking signature",
			Range:          setRange(syntax, 240, 346),
			SelectionRange: setRange(syntax, 240, 346),
			Children: []protocol.DocumentSymbol{
				{Name: "integer", Kind: protocol.Struct, Detail: "return type",
					Range:          setRange(syntax, 293, 300),
					SelectionRange: setRange(syntax, 293, 300),
					Children:       nil},
				{Name: "Exceptions", Kind: protocol.Array,
					Range:          setRange(syntax, 304, 346),
					SelectionRange: setRange(syntax, 304, 346),
					Children: []protocol.DocumentSymbol{
						{Name: "ExceptionType1", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 315, 329),
							SelectionRange: setRange(syntax, 315, 329),
							Children:       nil},
						{Name: "ExceptionType2", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 331, 345),
							SelectionRange: setRange(syntax, 331, 345),
							Children:       nil}}}}},
		{Name: "MyRemoteProcSix", Kind: protocol.Function, Detail: "non-blocking signature",
			Range:          setRange(syntax, 350, 431),
			SelectionRange: setRange(syntax, 350, 431),
			Children: []protocol.DocumentSymbol{
				{Name: "Exceptions", Kind: protocol.Array,
					Range:          setRange(syntax, 405, 431),
					SelectionRange: setRange(syntax, 405, 431),
					Children: []protocol.DocumentSymbol{
						{Name: "integer", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 416, 423),
							SelectionRange: setRange(syntax, 416, 423),
							Children:       nil},
						{Name: "float", Kind: protocol.Struct, Detail: "type",
							Range:          setRange(syntax, 425, 430),
							SelectionRange: setRange(syntax, 425, 430),
							Children:       nil}}}}}}, list)
}
