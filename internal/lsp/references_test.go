package lsp_test

import (
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/stretchr/testify/assert"
)

func StripPathFromURI(list []protocol.Location) []protocol.Location {
	ret := make([]protocol.Location, 0, len(list))
	for _, elem := range list {
		fileName := elem.URI[strings.LastIndexByte(string(elem.URI), '/')+1:]
		ret = append(ret, protocol.Location{URI: fileName, Range: elem.Range})
	}
	return ret
}

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
	db := &ttcn3.DB{}
	db.Index(project.FindAllFiles(suite)...)

	// Lookup `Msg`
	list := lsp.NewAllIdsWithSameName(db, "Byte")

	assert.Equal(t, []protocol.Location{
		{URI: "TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 15}, End: protocol.Position{Line: 2, Character: 15}}},
		{URI: "TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 22}, End: protocol.Position{Line: 3, Character: 22}}},
		{URI: "TestFindAllTypeDefs_Module_0.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 4, Character: 7}, End: protocol.Position{Line: 4, Character: 7}}},
		{URI: "TestFindAllTypeDefs_Module_1.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 12}, End: protocol.Position{Line: 3, Character: 12}}},
		{URI: "TestFindAllTypeDefs_Module_2.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 50}, End: protocol.Position{Line: 2, Character: 50}}},
		{URI: "TestFindAllTypeDefs_Module_2.ttcn3",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 8}, End: protocol.Position{Line: 3, Character: 8}}}}, StripPathFromURI(list))
}
