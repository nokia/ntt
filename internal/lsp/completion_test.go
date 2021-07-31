package lsp_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

var auxModulesMap = map[string]bool{"CCS": true, "CcsConnector": true, "ConversionFunc": true, "ExtProc_PortType": true, "Internet": true, "IpcCommunication": true, "LPTENB": true, "LameCodec": true,
	"ProtoBaseTypes": true, "SCTP": true, "SctpMapParameters": true, "TestManager": true, "ZmqMapParameters": true, "asnExtFunctions": true, "asnPreload": true, "config": true, "controlMessages": true,
	"math": true, "os": true, "snow3g": true, "strings": true, "tcpsConnector": true, "testCcsStub": true, "udpsLogger": true, "utilityFunctions": true, "zuc": true, "SutControl": true, "TestcaseExecutor": true}

func filterContentOfAuxModules(input []protocol.CompletionItem) []protocol.CompletionItem {
	ret := make([]protocol.CompletionItem, 0, len(input))
	for _, v := range input {
		mod := strings.Split(v.Detail, ".")
		if len(mod) > 1 {
			if _, ok := auxModulesMap[mod[0]]; ok {
				continue
			}
		}
		if _, ok := auxModulesMap[v.Label]; ok && v.Kind == protocol.ModuleCompletion {
			continue
		}
		ret = append(ret, v)
	}
	return ret
}
func buildSuite(t *testing.T, strs ...string) *ntt.Suite {
	suite := &ntt.Suite{}
	for i, s := range strs {
		name := fmt.Sprintf("%s_Module_%d.ttcn3", t.Name(), i)
		//file := fs.Open(name)
		//file.SetBytes([]byte(s))
		suite.AddSources(name)
		srcs, _ := suite.Sources()
		fh := fs.Open(srcs[len(srcs)-1])
		fh.SetBytes([]byte(s))
	}
	return suite
}

type Pos struct {
	Line   int
	Column int
}

func completionAt(t *testing.T, suite *ntt.Suite, pos loc.Pos) []protocol.CompletionItem {
	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	syntax := suite.ParseWithAllErrors(name)
	nodeStack := lsp.LastNonWsToken(syntax.Module, pos)
	name = name[:len(name)-len(filepath.Ext(name))]
	return lsp.NewCompListItems(suite, pos, nodeStack, name)
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
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "TestImportModulenamesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion, SortText: " TestImportModulenamesCtrlSpc_Module_1"},
		{Label: "TestImportModulenamesCtrlSpc_Module_2", Kind: protocol.ModuleCompletion, SortText: " TestImportModulenamesCtrlSpc_Module_2"}}, filterContentOfAuxModules(list))
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
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "TestImportModulenames_Module_1", Kind: protocol.ModuleCompletion, SortText: " TestImportModulenames_Module_1"},
		{Label: "TestImportModulenames_Module_2", Kind: protocol.ModuleCompletion, SortText: " TestImportModulenames_Module_2"}}, filterContentOfAuxModules(list))

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
		{Label: "B0", Kind: protocol.StructCompletion, SortText: " 1B0", Detail: "TestRunsOnTypesCtrlSpc_Module_0.B0"},
		{Label: "B1", Kind: protocol.StructCompletion, SortText: " 1B1", Detail: "TestRunsOnTypesCtrlSpc_Module_0.B1"},
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestRunsOnTypesCtrlSpc_Module_1.C0"},
		{Label: "A0", Kind: protocol.StructCompletion, SortText: " 1A0", Detail: "TestRunsOnTypesCtrlSpc_Module_2.A0"},
		{Label: "TestRunsOnTypesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion, SortText: " 2TestRunsOnTypesCtrlSpc_Module_1"},
		{Label: "TestRunsOnTypesCtrlSpc_Module_2", Kind: protocol.ModuleCompletion, SortText: " 2TestRunsOnTypesCtrlSpc_Module_2"}}, filterContentOfAuxModules(list))
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
		{Label: "B0", Kind: protocol.StructCompletion, SortText: " 1B0", Detail: "TestRunsOnTypes_Module_0.B0"},
		{Label: "B1", Kind: protocol.StructCompletion, SortText: " 1B1", Detail: "TestRunsOnTypes_Module_0.B1"},
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestRunsOnTypes_Module_1.C0"},
		{Label: "A0", Kind: protocol.StructCompletion, SortText: " 1A0", Detail: "TestRunsOnTypes_Module_2.A0"},
		{Label: "TestRunsOnTypes_Module_1", Kind: protocol.ModuleCompletion, SortText: " 2TestRunsOnTypes_Module_1"},
		{Label: "TestRunsOnTypes_Module_2", Kind: protocol.ModuleCompletion, SortText: " 2TestRunsOnTypes_Module_2"}}, filterContentOfAuxModules(list))
}

func TestRunsOnModuleDotTypesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		function f() runs on TestRunsOnModuleDotTypesCtrlSpc_Module_1.//
	  }`, `module TestRunsOnModuleDotTypesCtrlSpc_Module_1
      {
		  type component C0 {}
	  }`, `module TestRunsOnModuleDotTypesCtrlSpc_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 135)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestRunsOnModuleDotTypesCtrlSpc_Module_1.C0"}}, list)
}

func TestRunsOnModuleDotTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		function f() runs on TestRunsOnModuleDotTypes_Module_1.C//
	  }`, `module TestRunsOnModuleDotTypes_Module_1
      {
		  type component C0 {}
	  }`, `module TestRunsOnModuleDotTypes_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 128)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestRunsOnModuleDotTypes_Module_1.C0"}}, list)
}

func TestSystemTypesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		testcase f() runs on C0 system //
	  }`, `module TestSystemTypesCtrlSpc_Module_1
      {
		  type component C0 {}
	  }`, `module TestSystemTypesCtrlSpc_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 103)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "B0", Kind: protocol.StructCompletion, SortText: " 1B0", Detail: "TestSystemTypesCtrlSpc_Module_0.B0"},
		{Label: "B1", Kind: protocol.StructCompletion, SortText: " 1B1", Detail: "TestSystemTypesCtrlSpc_Module_0.B1"},
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestSystemTypesCtrlSpc_Module_1.C0"},
		{Label: "A0", Kind: protocol.StructCompletion, SortText: " 1A0", Detail: "TestSystemTypesCtrlSpc_Module_2.A0"},
		{Label: "TestSystemTypesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion, SortText: " 2TestSystemTypesCtrlSpc_Module_1"},
		{Label: "TestSystemTypesCtrlSpc_Module_2", Kind: protocol.ModuleCompletion, SortText: " 2TestSystemTypesCtrlSpc_Module_2"}}, filterContentOfAuxModules(list))
}

func TestSystemModuleDotTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		function f() runs on TestSystemModuleDotTypes_Module_1.C0 system TestSystemModuleDotTypes_Module_1.C//
	  }`, `module TestSystemModuleDotTypes_Module_1
      {
		  type component C0 {}
	  }`, `module TestSystemModuleDotTypes_Module_2
      {
		  type component A0 {}
	  }`)

	list := completionAt(t, suite, 172)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestSystemModuleDotTypes_Module_1.C0"}}, list)
}

func TestExtendsTypesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		type component B2 extends //
	  }`, `module TestExtendsTypesCtrlSpc_Module_1
      {
		  type component C0 {}
	  }`)

	list := completionAt(t, suite, 98)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "B0", Kind: protocol.StructCompletion, SortText: " 1B0", Detail: "TestExtendsTypesCtrlSpc_Module_0.B0"},
		{Label: "B1", Kind: protocol.StructCompletion, SortText: " 1B1", Detail: "TestExtendsTypesCtrlSpc_Module_0.B1"},
		{Label: "B2", Kind: protocol.StructCompletion, SortText: " 1B2", Detail: "TestExtendsTypesCtrlSpc_Module_0.B2"}, // TODO: filter 'self' out
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestExtendsTypesCtrlSpc_Module_1.C0"},
		{Label: "TestExtendsTypesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion, SortText: " 2TestExtendsTypesCtrlSpc_Module_1"}}, filterContentOfAuxModules(list))
}

func TestExtendsTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		type component B2 extends B//
	  }`, `module TestExtendsTypes_Module_1
      {
		  type component C0 {}
	  }`)

	list := completionAt(t, suite, 99)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "B0", Kind: protocol.StructCompletion, SortText: " 1B0", Detail: "TestExtendsTypes_Module_0.B0"},
		{Label: "B1", Kind: protocol.StructCompletion, SortText: " 1B1", Detail: "TestExtendsTypes_Module_0.B1"},
		{Label: "B2", Kind: protocol.StructCompletion, SortText: " 1B2", Detail: "TestExtendsTypes_Module_0.B2"},
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestExtendsTypes_Module_1.C0"},
		{Label: "TestExtendsTypes_Module_1", Kind: protocol.ModuleCompletion, SortText: " 2TestExtendsTypes_Module_1"}}, filterContentOfAuxModules(list))
}

func TestExtendsModuleDotTypes(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
        type component B0 {}
		type component B1 {}
		type component B2 extends TestExtendsModuleDotTypes_Module_1.//
	  }`, `module TestExtendsModuleDotTypes_Module_1
      {
		  type component C0 {}
	  }`)

	list := completionAt(t, suite, 133)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "C0", Kind: protocol.StructCompletion, SortText: " 1C0", Detail: "TestExtendsModuleDotTypes_Module_1.C0"}}, list)
}

func TestModifiesCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		import from TestModifiesCtrlSpc_Module_1 all;
		template R t_r := *;
		template integer t_i := ?;
		template (omit) R t_rmod modifies //
	  }`, `module TestModifiesCtrlSpc_Module_1
      {
		  type record R {integer f1, boolean f2 optional}
		  template (value) R t_r2 := {10, omit}
	  }`)

	list := completionAt(t, suite, 155)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "t_r", Kind: protocol.ConstantCompletion, Detail: "TestModifiesCtrlSpc_Module_0.t_r"},
		{Label: "t_i", Kind: protocol.ConstantCompletion, Detail: "TestModifiesCtrlSpc_Module_0.t_i"},       // TODO: implement filter on Compatible Type
		{Label: "t_rmod", Kind: protocol.ConstantCompletion, Detail: "TestModifiesCtrlSpc_Module_0.t_rmod"}, // TODO: implement filter for self
		{Label: "t_r2", Kind: protocol.ConstantCompletion, Detail: "TestModifiesCtrlSpc_Module_1.t_r2"},
		{Label: "TestModifiesCtrlSpc_Module_1", Kind: protocol.ModuleCompletion}}, filterContentOfAuxModules(list))
}

func TestModifiesParseErrorCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test {
		   template integer t_base := ?
		   template in//

		   function setup_ports()
			runs on CpMctMain
		{
			map(system:ifF1Sctp[0], mtc:ifF1Sctp[0]) param(t_f1SctpMapParam1);
			map(system:ifF1Sctp[1], mtc:ifF1Sctp[1]) param(t_f1SctpMapParam2);
			map(system:ifF1Sctp[2], mtc:ifF1Sctp[2]) param(t_f1SctpMapParam3);
			map(system:ifF1Sctp[3], mtc:ifF1Sctp[3]) param(t_f1SctpMapParam4);
			map(system:ifF1Sctp[4], mtc:ifF1Sctp[4]) param(t_f1SctpMapParam5);
			map(system:ifF1Sctp[5], mtc:ifF1Sctp[5]) param(t_f1SctpMapParam6);
			map(system:ifF1Sctp[6], mtc:ifF1Sctp[6]) param(t_f1SctpMapParam7);
			map(system:ifX2Sctp[0], mtc:ifX2Sctp[0]) param(t_x2SctpMapParam);
			map(system:ifXNSctp[0], mtc:ifXNSctp[0]) param(t_xnSctpMapParam);
		}
	}`)

	list := completionAt(t, suite, 65)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "anytype ", Kind: protocol.KeywordCompletion},
		{Label: "bitstring ", Kind: protocol.KeywordCompletion},
		{Label: "boolean ", Kind: protocol.KeywordCompletion},
		{Label: "charstring ", Kind: protocol.KeywordCompletion},
		{Label: "default ", Kind: protocol.KeywordCompletion},
		{Label: "float ", Kind: protocol.KeywordCompletion},
		{Label: "hexstring ", Kind: protocol.KeywordCompletion},
		{Label: "integer ", Kind: protocol.KeywordCompletion},
		{Label: "octetstring ", Kind: protocol.KeywordCompletion},
		{Label: "universal charstring ", Kind: protocol.KeywordCompletion},
		{Label: "verdicttype ", Kind: protocol.KeywordCompletion}}, filterContentOfAuxModules(list))
}

func TestTemplateTypeCtrlSpc(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type integer Byte(0..255);
		template //
		template integer a_i := ?;
	}`, `module TestTemplateTypeCtrlSpc_Module_1
      {
		  type record R {integer f1, boolean f2 optional}
		  template (value) R t_r2 := {10, omit}
	  }`)

	list := completionAt(t, suite, 59)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "Byte ", Kind: protocol.StructCompletion, SortText: " 1Byte", Detail: "TestTemplateTypeCtrlSpc_Module_0.Byte"},
		{Label: "R ", Kind: protocol.StructCompletion, SortText: " 2R", Detail: "TestTemplateTypeCtrlSpc_Module_1.R"},
		{Label: "anytype ", Kind: protocol.KeywordCompletion},
		{Label: "bitstring ", Kind: protocol.KeywordCompletion},
		{Label: "boolean ", Kind: protocol.KeywordCompletion},
		{Label: "charstring ", Kind: protocol.KeywordCompletion},
		{Label: "default ", Kind: protocol.KeywordCompletion},
		{Label: "float ", Kind: protocol.KeywordCompletion},
		{Label: "hexstring ", Kind: protocol.KeywordCompletion},
		{Label: "integer ", Kind: protocol.KeywordCompletion},
		{Label: "octetstring ", Kind: protocol.KeywordCompletion},
		{Label: "universal charstring ", Kind: protocol.KeywordCompletion},
		{Label: "verdicttype ", Kind: protocol.KeywordCompletion},
		{Label: "TestTemplateTypeCtrlSpc_Module_1", Kind: protocol.ModuleCompletion, SortText: " 3TestTemplateTypeCtrlSpc_Module_1"}}, filterContentOfAuxModules(list))
}

func TestTemplateType(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type integer Byte(0..255);
		template h//
		template integer a_i := ?;
	}`, `module TestTemplateType_Module_1
      {
		  type record R {integer f1, boolean f2 optional}
		  template (value) R t_r2 := {10, omit}
	  }`)

	list := completionAt(t, suite, 60)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "Byte ", Kind: protocol.StructCompletion, SortText: " 1Byte", Detail: "TestTemplateType_Module_0.Byte"},
		{Label: "R ", Kind: protocol.StructCompletion, SortText: " 2R", Detail: "TestTemplateType_Module_1.R"},
		{Label: "anytype ", Kind: protocol.KeywordCompletion},
		{Label: "bitstring ", Kind: protocol.KeywordCompletion},
		{Label: "boolean ", Kind: protocol.KeywordCompletion},
		{Label: "charstring ", Kind: protocol.KeywordCompletion},
		{Label: "default ", Kind: protocol.KeywordCompletion},
		{Label: "float ", Kind: protocol.KeywordCompletion},
		{Label: "hexstring ", Kind: protocol.KeywordCompletion},
		{Label: "integer ", Kind: protocol.KeywordCompletion},
		{Label: "octetstring ", Kind: protocol.KeywordCompletion},
		{Label: "universal charstring ", Kind: protocol.KeywordCompletion},
		{Label: "verdicttype ", Kind: protocol.KeywordCompletion},
		{Label: "TestTemplateType_Module_1", Kind: protocol.ModuleCompletion, SortText: " 3TestTemplateType_Module_1"}}, filterContentOfAuxModules(list))
}

func TestTemplateModuleDotType(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type integer Byte(0..255);
		template TestTemplateModuleDotType_Module_1.//
		template integer a_i := ?;  // NOTE: this template kw is apparently consumed by the parser leading to integer being interpreted as Name!!!
	}`, `module TestTemplateModuleDotType_Module_1
      {
		  type record R {integer f1, boolean f2 optional}
		  template (value) R t_r2 := {10, omit}
	  }`)

	list := completionAt(t, suite, 94)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "R ", Kind: protocol.StructCompletion, Detail: "TestTemplateModuleDotType_Module_1.R"}}, filterContentOfAuxModules(list))
}

func TestModifiesModuleDot(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		import from TestModifiesModuleDot_Module_1 all;
		template R t_r := *;
		template integer t_i := ?;
		template (omit) R t_rmod modifies TestModifiesModuleDot_Module_1.//
	  }`, `module TestModifiesModuleDot_Module_1
      {
		  type record R {integer f1, boolean f2 optional}
		  template (value) R t_r2 := {10, omit}
	  }`)

	list := completionAt(t, suite, 188)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "t_r2", Kind: protocol.ConstantCompletion}}, filterContentOfAuxModules(list))
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

func TestPortTypeInsideComponent(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
		type port P1 message {
			inout charstring
		}
        type component B0 {
			port P//
		}
	  }`, `module TestPortTypeInsideComponent_Module_1
      {
		  type port P2 message {
			  in integer
			  out float
		  }
	  }`)

	list := completionAt(t, suite, 104)
	assert.Equal(t, []protocol.CompletionItem{
		{Label: "P1", Kind: protocol.InterfaceCompletion, SortText: " 1P1", Detail: "TestPortTypeInsideComponent_Module_0.P1"},
		{Label: "P2", Kind: protocol.InterfaceCompletion, SortText: " 2P2", Detail: "TestPortTypeInsideComponent_Module_1.P2"},
		{Label: "TestPortTypeInsideComponent_Module_1", Kind: protocol.ModuleCompletion, SortText: " 3TestPortTypeInsideComponent_Module_1"}}, filterContentOfAuxModules(list))
}

// TODO: fixing this issue requires more effort.
/*
func TestSyntaxErrorProvokingInvalidPos(t *testing.T) {
	suite := buildSuite(t, `module Test
    {
    type component Ptc {}
    type component Sys {}
    function setColor(integer p_color) runs on {
        log(p_color);
    }
    testcase tc1() runs on test.Ptc system Sys { }
    c
	  }`)
	name := fmt.Sprintf("%s_Module_0.ttcn3", t.Name())
	syntax := suite.ParseWithAllErrors(name)
	pos := syntax.Pos(9, 6)
	assert.Equal(t, pos, 203)
}
*/
