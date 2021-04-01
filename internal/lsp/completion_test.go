package lsp_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func buildSuite(t *testing.T, strs ...string) *ntt.Suite {
	suite := &ntt.Suite{}
	for i, s := range strs {
		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
		//file := suite.File(name)
		//file.SetBytes([]byte(s))
		suite.AddSources(name)
		fh, _ := suite.Sources()
		fh[len(fh)-1].SetBytes([]byte(s))
	}
	return suite
}

type Pos struct {
	Line   int
	Column int
}

func completionAt(t *testing.T, suite *ntt.Suite, pos loc.Pos) []protocol.CompletionItem {
	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	syntax := suite.Parse(name)
	nodeStack := lsp.LastNonWsToken(syntax.Module, pos)

	return lsp.NewCompListItems(suite, pos, nodeStack)
}
func gotoDefinition(suite *ntt.Suite, file string, line, column int) Pos {
	id, _ := suite.IdentifierAt(file, line, column)
	if id == nil || id.Def == nil {
		return Pos{}
	}
	return Pos{
		Line:   id.Def.Position.Line,
		Column: id.Def.Position.Column,
	}
}

// Completion within Import statement.
// TODO: func TestImportModulenamesCtrlSpc(t *testing.T) {}
// TODO: func TestImportBehaviours(t *testing.T)  {}
// TODO: func TestImportBehavioursCtrlSpc(t *testing.T)  {}
// TODO: func TestImportTypes(t *testing.T) {}
// TODO: func TestImportTypesCtrlSpc(t *testing.T) {}

func TestImportModulenames(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		import from
		import from A all;
	  }`, `module A
	  {}`, `module B
	  {}`)

	// Lookup `Msg`
	list := completionAt(t, suite, 30)
	log.Debug(fmt.Sprintf("Node not considered yet: %#v)", list))
	assert.Equal(t, []protocol.CompletionItem{{Label: "TestImportModulenames_Module_0", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenames_Module_1", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenames_Module_2", Kind: protocol.ModuleCompletion}}, list)

}
