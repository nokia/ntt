package lsp_test

import (
	"fmt"
	"path/filepath"
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
	name = name[:len(name)-len(filepath.Ext(name))]
	return lsp.NewCompListItems(suite, pos, nodeStack, name)
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
// TODO: func TestImportTypes(t *testing.T) {}
// TODO: func TestImportTypesCtrlSpc(t *testing.T) {}

func TestImportModulenamesCtrlSpc(t *testing.T) {
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
	assert.Equal(t, []protocol.CompletionItem{{Label: "TestImportModulenamesCtrlSpc_Module_0", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenamesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenamesCtrlSpc_Module_2", Kind: protocol.ModuleCompletion}}, list)
}

func TestImportModulenames(t *testing.T) {
	suite := buildSuite(t, `module Test
	{
		import from Tes
		import from A all;
	  }`, `module A
	  {}`, `module B
	  {}`)

	// Lookup `Msg`
	list := completionAt(t, suite, 33)
	log.Debug(fmt.Sprintf("Node not considered yet: %#v)", list))
	assert.Equal(t, []protocol.CompletionItem{{Label: "TestImportModulenames_Module_0", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenames_Module_1", Kind: protocol.ModuleCompletion},
		{Label: "TestImportModulenames_Module_2", Kind: protocol.ModuleCompletion}}, list)

}

func TestImportBehavioursCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportBehavioursCtrlSpc_Module_1 {
            altstep   }
		import from A all;
	  }`, `module TestImportBehavioursCtrlSpc_Module_1
      {
		  altstep a1() {}
		  altstep a2() {}
	  }`, `module B
	  {}`)

	// Lookup `Msg`
	list := completionAt(t, suite, 99)
	log.Debug(fmt.Sprintf("Node not considered yet: %#v)", list))
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "a1", Kind: protocol.FunctionCompletion},
		{Label: "a2", Kind: protocol.FunctionCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportBehaviours(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportBehaviours_Module_1 {
            testcase t}
		import from A all;
	  }`, `module TestImportBehaviours_Module_1
      {
		  testcase tc1() runs on C0 {}
		  testcase tc2() runs on C0 {}
	  }`, `module B
	  {}`)

	// Lookup `Msg`
	list := completionAt(t, suite, 92)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "tc1", Kind: protocol.FunctionCompletion},
		{Label: "tc2", Kind: protocol.FunctionCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportTemplates(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportTemplates_Module_1 {
            template a_}
		import from A all;
	  }`, `module TestImportTemplates_Module_1
      {
		  template charstring a_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  template R1 a_r1(boolean b) := {f1 := 10, f2 := b}
		  function f1() {template integer a_local_int := 0;}
	  }`, `module B
	  {}`)

	// Lookup `Msg`
	list := completionAt(t, suite, 92)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "a_name", Kind: protocol.ConstantCompletion},
		{Label: "a_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportTemplatesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportTemplatesCtrlSpc_Module_1 {
            template }
		import from A all;
	  }`, `module TestImportTemplatesCtrlSpc_Module_1
      {
		  template charstring a_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  template R1 a_r1(boolean b) := {f1 := 10, f2 := b}
	  }`, `module B
	  {}`)

	list := completionAt(t, suite, 97)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "a_name", Kind: protocol.ConstantCompletion},
		{Label: "a_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportConstantsCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportConstantsCtrlSpc_Module_1 {
            const }
		import from A all;
	  }`, `module TestImportConstantsCtrlSpc_Module_1
      {
		  const charstring c_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  const R1 c_r1 := {f1 := 10, f2 := false}
		  function f1() { const integer c_localInt := 0;}
	  }`)

	list := completionAt(t, suite, 94)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "c_name", Kind: protocol.ConstantCompletion},
		{Label: "c_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportExceptConstantsCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportExceptConstantsCtrlSpc_Module_1 all except {
            const }
		import from A all;
	  }`, `module TestImportExceptConstantsCtrlSpc_Module_1
      {
		  const charstring c_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  const R1 c_r1 := {f1 := 10, f2 := false}
	  }`)

	list := completionAt(t, suite, 111)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "c_name", Kind: protocol.ConstantCompletion},
		{Label: "c_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportExceptConstants(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportExceptConstants_Module_1 all except {
            const c_}
		import from A all;
	  }`, `module TestImportExceptConstants_Module_1
      {
		  const charstring c_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  const R1 c_r1 := {f1 := 10, f2 := false}
	  }`)

	list := completionAt(t, suite, 106)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "c_name", Kind: protocol.ConstantCompletion},
		{Label: "c_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportModuleparsCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportModuleparsCtrlSpc_Module_1 {
            modulepar }
		import from A all;
	  }`, `module TestImportModuleparsCtrlSpc_Module_1
      {
		  modulepar charstring m_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  modulepar R1 m_r1 := {f1 := 10, f2 := true}
		  const integer c_int := 2;
	  }`)

	list := completionAt(t, suite, 101)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "m_name", Kind: protocol.ConstantCompletion},
		{Label: "m_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportExceptModuleparsCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportExceptModuleparsCtrlSpc_Module_1 all except {
            modulepar }
		import from A all;
	  }`, `module TestImportExceptModuleparsCtrlSpc_Module_1
      {
		  modulepar charstring m_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  modulepar R1 m_r1 := {f1 := 10, f2 := false}
	  }`)

	list := completionAt(t, suite, 116)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "m_name", Kind: protocol.ConstantCompletion},
		{Label: "m_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportExceptModulepars(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportExceptModulepars_Module_1 all except {
            modulepar m_}
		import from A all;
	  }`, `module TestImportExceptModulepars_Module_1
      {
		  modulepar charstring m_name := "Justus"
		  type record R1 {integer f1, boolean f2}
		  modulepar R1 m_r1 := {f1 := 10, f2 := false}
	  }`)

	list := completionAt(t, suite, 111)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "m_name", Kind: protocol.ConstantCompletion},
		{Label: "m_r1", Kind: protocol.ConstantCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportTypesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportTypesCtrlSpc_Module_1 {
            type }
		import from A all;
	  }`, `module TestImportTypesCtrlSpc_Module_1
      {
		  type charstring String;
		  type record R1 {integer f1, boolean f2}
		  type set S1 {integer f1, boolean f2}
		  type union U1 {integer f1, boolean f2}
		  type record of integer RoI1;
		  type record length(2..10) of integer RoI2;
		  type record of T MyList <in type T>;
		  type set of integer SoI;
		  type function MyBehavType() return integer;
		  type enumerated E1 {red, green, blue};
		  type port MyPort message {inout E1}
		  type component C0 {}
	  }`)

	list := completionAt(t, suite, 89)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "String", Kind: protocol.StructCompletion},
		{Label: "R1", Kind: protocol.StructCompletion},
		{Label: "S1", Kind: protocol.StructCompletion},
		{Label: "U1", Kind: protocol.StructCompletion},
		{Label: "RoI1", Kind: protocol.StructCompletion},
		{Label: "RoI2", Kind: protocol.StructCompletion},
		{Label: "MyList", Kind: protocol.StructCompletion},
		{Label: "SoI", Kind: protocol.StructCompletion},
		{Label: "MyBehavType", Kind: protocol.StructCompletion},
		{Label: "E1", Kind: protocol.StructCompletion},
		{Label: "MyPort", Kind: protocol.StructCompletion},
		{Label: "C0", Kind: protocol.StructCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestImportTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        import from TestImportTypes_Module_1 {
            type E}
		import from A all;
	  }`, `module TestImportTypes_Module_1
      {
		  type charstring String;
		  type record R1 {integer f1, boolean f2}
		  type set S1 {integer f1, boolean f2}
		  type union U1 {integer f1, boolean f2}
		  type record of integer RoI1;
		  type record length(2..10) of integer RoI2;
		  type record of T MyList <in type T>;
		  type set of integer SoI;
		  type function MyBehavType() return integer;
		  type enumerated E1 {red, green, blue};
		  type port MyPort message {inout E1}
		  type component C0 {}
	  }`)

	list := completionAt(t, suite, 83)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "String", Kind: protocol.StructCompletion},
		{Label: "R1", Kind: protocol.StructCompletion},
		{Label: "S1", Kind: protocol.StructCompletion},
		{Label: "U1", Kind: protocol.StructCompletion},
		{Label: "RoI1", Kind: protocol.StructCompletion},
		{Label: "RoI2", Kind: protocol.StructCompletion},
		{Label: "MyList", Kind: protocol.StructCompletion},
		{Label: "SoI", Kind: protocol.StructCompletion},
		{Label: "MyBehavType", Kind: protocol.StructCompletion},
		{Label: "E1", Kind: protocol.StructCompletion},
		{Label: "MyPort", Kind: protocol.StructCompletion},
		{Label: "C0", Kind: protocol.StructCompletion},
		{Label: "all;", Kind: protocol.KeywordCompletion}}, list)
}

func TestRunsOnTypesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		function f() runs on //
	  }`, `module TestRunsOnTypesCtrlSpc_Module_1
      {
		  type component C0 {}
	  }`, `module TestRunsOnTypesCtrlSpc_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 93)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "B0", Kind: protocol.StructCompletion},
		{Label: "B1", Kind: protocol.StructCompletion},
		{Label: "C0", Kind: protocol.StructCompletion},
		{Label: "A0", Kind: protocol.StructCompletion},
		{Label: "TestRunsOnTypesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion},
		{Label: "TestRunsOnTypesCtrlSpc_Module_2", Kind: protocol.ModuleCompletion}}, list)
}

func TestRunsOnTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		function f() runs on A//
	  }`, `module TestRunsOnTypes_Module_1
      {
		  type component C0 {}
	  }`, `module TestRunsOnTypes_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 94)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "B0", Kind: protocol.StructCompletion},
		{Label: "B1", Kind: protocol.StructCompletion},
		{Label: "C0", Kind: protocol.StructCompletion},
		{Label: "A0", Kind: protocol.StructCompletion},
		{Label: "TestRunsOnTypes_Module_1", Kind: protocol.ModuleCompletion},
		{Label: "TestRunsOnTypes_Module_2", Kind: protocol.ModuleCompletion}}, list)
}

func TestSubTypeDefSegv(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type r
	  }`)

	// Lookup `Msg`
	list := completionAt(t, suite, 32)
	assert.Empty(t, list)
}

func TestNewModuleDef(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        t
	  }`)

	// Lookup `Msg`
	list := completionAt(t, suite, 27)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "import from ", Kind: protocol.KeywordCompletion},
		{Label: "type ", Kind: protocol.KeywordCompletion},
		{Label: "const ", Kind: protocol.KeywordCompletion},
		{Label: "modulepar ", Kind: protocol.KeywordCompletion},
		{Label: "template ", Kind: protocol.KeywordCompletion},
		{Label: "function ", Kind: protocol.KeywordCompletion},
		{Label: "external function ", Kind: protocol.KeywordCompletion},
		{Label: "altstep ", Kind: protocol.KeywordCompletion},
		{Label: "testcase ", Kind: protocol.KeywordCompletion},
		{Label: "control ", Kind: protocol.KeywordCompletion},
		{Label: "signature ", Kind: protocol.KeywordCompletion}}, list)
}
