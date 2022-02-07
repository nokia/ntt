package lsp_test

import (
	"testing"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestFindAllTypeDefs(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		type integer Byte(0..255);
		function f() return Byte {
			var Byte ret := 100;
		}
	  }`, `module A
	  {
			import from TestFindAllTypeDefs_Module_0 all;
			template Byte a_byte := ?;
	  }`, `module B
	  {
			import from TestFindAllTypeDefs_Module_0 {type Byte}
			type Byte AliasByte;
	  }`)

	// Lookup `Msg`
	list := lsp.NewAllIdsWithSameName(suite, "Byte")

	assert.Equal(t, []protocol.Location{
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 15}, End: protocol.Position{Line: 2, Character: 15}}},
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 22}, End: protocol.Position{Line: 3, Character: 22}}},
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 4, Character: 7}, End: protocol.Position{Line: 4, Character: 7}}},
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_1.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 12}, End: protocol.Position{Line: 3, Character: 12}}},
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_2.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 50}, End: protocol.Position{Line: 2, Character: 50}}},
		{URI: "file:///home/ut/ntt/internal/lsp/TestFindAllTypeDefs_Module_2.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 8}, End: protocol.Position{Line: 3, Character: 8}}}}, list)
}
