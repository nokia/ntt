package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/stretchr/testify/assert"
)

func TestFindAllTypeDefs(t *testing.T) {
	input1 := `
	module A {
		type integer Byte(0..255);
		function f() return Byte {
		var Byte ret := 100;
		}
	}
	module B {
		import from A all;
		template Byte a_byte := ?;
	}`
	input2 := `
	module C {
		import from A {type Byte}
		type Byte AliasByte;
	}`

	name1 := fmt.Sprintf("test://%s_input1", t.Name())
	name2 := fmt.Sprintf("test://%s_input2", t.Name())

	fs.SetContent(name1, []byte(input1))
	fs.SetContent(name2, []byte(input2))

	db := &ttcn3.DB{}
	db.Index(name1, name2)

	// Lookup `Msg`
	list := lsp.NewAllIdsWithSameName(db, "Byte")

	assert.Equal(t, []protocol.Location{
		{URI: "test://TestFindAllTypeDefs_input1",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 15}, End: protocol.Position{Line: 2, Character: 19}}},
		{URI: "test://TestFindAllTypeDefs_input1",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 22}, End: protocol.Position{Line: 3, Character: 26}}},
		{URI: "test://TestFindAllTypeDefs_input1",
			Range: protocol.Range{Start: protocol.Position{Line: 4, Character: 6}, End: protocol.Position{Line: 4, Character: 10}}},
		{URI: "test://TestFindAllTypeDefs_input1",
			Range: protocol.Range{Start: protocol.Position{Line: 9, Character: 11}, End: protocol.Position{Line: 9, Character: 15}}},
		{URI: "test://TestFindAllTypeDefs_input2",
			Range: protocol.Range{Start: protocol.Position{Line: 2, Character: 22}, End: protocol.Position{Line: 2, Character: 26}}},
		{URI: "test://TestFindAllTypeDefs_input2",
			Range: protocol.Range{Start: protocol.Position{Line: 3, Character: 7}, End: protocol.Position{Line: 3, Character: 11}}}}, list)
}
