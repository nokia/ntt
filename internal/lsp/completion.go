package lsp

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/printer"
	"github.com/nokia/ntt/ttcn3/syntax"
)

type FunctionDetails struct {
	Label         string
	Signature     string
	Documentation string
	HasRunsOn     bool
	HasReturn     bool
	HasParameters bool
	TextFormat    protocol.InsertTextFormat
}

type BehavAttrib int

// The list of behaviours.
const (
	NONE        BehavAttrib = iota //neither return nor runs on spec
	WITH_RETURN                    // only retrurn spec
	WITH_RUNSON                    // only runs on spec
)

var (
	moduleDefKw               = []string{"import from ", "type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control ", "signature "}
	importAfterModName        = []string{"all [except {}];", "{}"}
	importAfterModNameSnippet = []string{"${1:all${2: except {$3\\}}};$0", "{$0}"}
	importKinds               = []string{"type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control", "signature "}
	predefinedTypes           = []string{"anytype ", "bitstring ", "boolean ", "charstring ", "default ", "float ", "hexstring ", "integer ", "octetstring ", "universal charstring ", "verdicttype "}
)

func newPredefinedFunctions() []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range PredefinedFunctions {
		markup := protocol.MarkupContent{Kind: "markdown", Value: v.Documentation}
		ret = append(ret, protocol.CompletionItem{
			Label: v.Label, Kind: protocol.FunctionCompletion,
			Detail:           v.Signature,
			InsertTextFormat: v.TextFormat,
			InsertText:       v.InsertText,
			Documentation:    markup})
	}
	return ret
}

func newImportkinds() []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range importKinds {
		ret = append(ret, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return ret
}

func newPredefinedTypes() []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range predefinedTypes {
		ret = append(ret, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return ret
}

func getAllBehavioursFromModule(tree *ttcn3.Tree, kind syntax.Kind, mname string) []*FunctionDetails {
	var ret []*FunctionDetails
	tree.Inspect(func(n syntax.Node) bool {
		node, ok := n.(*syntax.FuncDecl)
		if !ok {
			return true
		}

		if node.Kind.Kind() != kind {
			return false
		}

		var sig bytes.Buffer
		textFormat := protocol.PlainTextTextFormat
		sig.WriteString(node.Kind.String() + " " + mname + "." + node.Name.String())
		len1 := len(sig.String())
		printer.Print(&sig, node.Params)
		hasParams := (len(sig.String()) - len1) > 2
		if hasParams {
			textFormat = protocol.SnippetTextFormat
		}
		if node.RunsOn != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, node.RunsOn)
		}
		if node.System != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, node.System)
		}
		if node.Return != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, node.Return)
		}
		ret = append(ret, &FunctionDetails{
			Label:         node.Name.String(),
			HasRunsOn:     (node.RunsOn != nil),
			HasReturn:     (node.Return != nil),
			Signature:     sig.String(),
			Documentation: syntax.Doc(node),
			HasParameters: hasParams,
			TextFormat:    textFormat})
		return false

	})
	return ret
}

func getAllValueDeclsFromModule(tree *ttcn3.Tree, mname string, kind syntax.Kind) []string {
	var ret []string
	tree.Inspect(func(n syntax.Node) bool {
		if n == nil {
			// called on node exit
			return false
		}

		switch node := n.(type) {
		case *syntax.FuncDecl, *syntax.ComponentTypeDecl:
			// do not descent into TESTCASE, FUNCTION, ALTSTEP,
			// component type
			return false
		case *syntax.ValueDecl:
			if node.Kind.Kind() != kind {
				return false
			}
			return true
		case *syntax.Declarator:
			ret = append(ret, syntax.Name(n))
			return false
		case *syntax.TemplateDecl:
			if kind == syntax.TEMPLATE {
				ret = append(ret, syntax.Name(n))
			}
			return false
		default:
			return true
		}
	})
	return ret
}

func isType(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.BehaviourTypeDecl,
		*syntax.ComponentTypeDecl,
		*syntax.EnumTypeDecl,
		*syntax.PortTypeDecl,
		*syntax.StructTypeDecl,
		*syntax.SubTypeDecl:
		return true
	default:
		return false
	}
}

func getAllTypesFromModule(tree *ttcn3.Tree, mname string) []string {
	var ret []string
	tree.Inspect(func(n syntax.Node) bool {
		if !isType(n) {
			return true
		}
		ret = append(ret, syntax.Name(n))
		return false
	})
	return ret
}

func getAllComponentTypesFromModule(tree *ttcn3.Tree, mname string) []string {
	var ret []string
	tree.Inspect(func(n syntax.Node) bool {
		if _, ok := n.(*syntax.ComponentTypeDecl); !ok {
			return true
		}
		ret = append(ret, syntax.Name(n))
		return false
	})
	return ret
}

func getAllPortTypesFromModule(tree *ttcn3.Tree, mname string) []string {
	var ret []string
	tree.Inspect(func(n syntax.Node) bool {
		if _, ok := n.(*syntax.PortTypeDecl); !ok {
			return true
		}
		ret = append(ret, syntax.Name(n))
		return false
	})
	return ret
}

func newImportBehaviours(tree *ttcn3.Tree, kind syntax.Kind, mname string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range getAllBehavioursFromModule(tree, kind, mname) {
		ret = append(ret, protocol.CompletionItem{Label: v.Label, Kind: protocol.FunctionCompletion})
	}
	ret = append(ret, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return ret
}

func newAllBehavioursFromModule(tree *ttcn3.Tree, kinds []syntax.Kind, attribs []BehavAttrib, mname string, sortPref string) []protocol.CompletionItem {
	var (
		items []*FunctionDetails
		ret   []protocol.CompletionItem
	)

	for _, kind := range kinds {
		items = append(items, getAllBehavioursFromModule(tree, kind, mname)...)
	}

	for _, v := range items {
		insertText := v.Label + "()"
		if v.HasParameters {
			insertText = v.Label + "($1)$0"
		}
		isSelected := false
		for _, attrib := range attribs {
			switch attrib {
			case NONE:
				isSelected = isSelected || !(v.HasReturn || v.HasRunsOn)
			case WITH_RETURN:
				isSelected = isSelected || v.HasReturn
			case WITH_RUNSON:
				isSelected = isSelected || v.HasRunsOn
			}
		}
		if isSelected {
			ret = append(ret, protocol.CompletionItem{Label: v.Label + "()",
				InsertText: insertText,
				Kind:       protocol.FunctionCompletion, SortText: sortPref + v.Label,
				Detail: v.Signature, Documentation: v.Documentation, InsertTextFormat: v.TextFormat})
		}
	}
	return ret
}

func newAllBehaviours(suite *Suite, kinds []syntax.Kind, attribs []BehavAttrib, mname string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		fileName := baseName(f)
		file, err := suite.FindModule(fileName)
		if err != nil {
			log.Debugf("module %s not found\n", mname)
			return nil
		}
		tree := ttcn3.ParseFile(file)

		sortPref := " 2"
		if fileName == mname {
			sortPref = " 1"
		}
		ret = append(ret, newAllBehavioursFromModule(tree, kinds, attribs, fileName, sortPref)...)
	}
	return ret
}

func newValueDeclsFromModule(tree *ttcn3.Tree, mname string, kind syntax.Kind, withDetail bool) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range getAllValueDeclsFromModule(tree, mname, kind) {
		item := protocol.CompletionItem{
			Label: v,
			Kind:  protocol.ConstantCompletion,
		}
		if withDetail {
			item.Detail = mname + "." + v
		}
		ret = append(ret, item)
	}
	return ret
}

func newModuleDefKw() []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, v := range moduleDefKw {
		ret = append(ret, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return ret
}

func newImportAfterModName() []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for i, v := range importAfterModName {
		ret = append(ret, protocol.CompletionItem{Label: v,
			InsertText:       importAfterModNameSnippet[i],
			InsertTextFormat: protocol.SnippetTextFormat, Kind: protocol.KeywordCompletion})
	}
	return ret
}

func newImportCompletions(suite *Suite, kind syntax.Kind, mname string) []protocol.CompletionItem {
	file, err := suite.FindModule(mname)
	if err != nil {
		log.Debugf("module %s not found\n", mname)
		return nil
	}
	tree := ttcn3.ParseFile(file)

	switch kind {
	case syntax.ALTSTEP, syntax.FUNCTION, syntax.TESTCASE:
		return newImportBehaviours(tree, kind, mname)
	case syntax.TEMPLATE, syntax.CONST, syntax.MODULEPAR:
		ret := newValueDeclsFromModule(tree, mname, kind, false)
		ret = append(ret, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
		return ret
	case syntax.TYPE:
		var ret []protocol.CompletionItem
		// NOTE: instead of 'StructCompletion' a better matching kind could be used
		for _, v := range getAllTypesFromModule(tree, mname) {
			ret = append(ret, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion})
		}
		ret = append(ret, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
		return ret
	default:
		log.Debugln(fmt.Sprintf("Kind not considered yet: %#v)", kind))
		return nil
	}
}

func moduleNameListFromSuite(suite *Suite, ownModName string, sortPref string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		fileName := baseName(f)
		if fileName == ownModName {
			continue
		}
		item := protocol.CompletionItem{
			Label: fileName,
			Kind:  protocol.ModuleCompletion,
		}
		if sortPref != "" {
			item.SortText = sortPref + fileName
		}
		ret = append(ret, item)
	}
	return ret
}

func newAllTypesFromModule(suite *Suite, mname string, sortPref string) []protocol.CompletionItem {
	file, err := suite.FindModule(mname)
	if err != nil {
		log.Debugf("module %s not found\n", mname)
		return nil
	}
	tree := ttcn3.ParseFile(file)

	var ret []protocol.CompletionItem
	for _, v := range getAllTypesFromModule(tree, mname) {
		item := protocol.CompletionItem{
			Label:  v + " ",
			Kind:   protocol.StructCompletion,
			Detail: mname + "." + v,
		}
		if sortPref != "" {
			item.SortText = sortPref + v
		}
		ret = append(ret, item)
	}
	return ret
}

func newAllComponentTypesFromModule(suite *Suite, mname string, sortPref string) []protocol.CompletionItem {
	file, err := suite.FindModule(mname)
	if err != nil {
		log.Debugf("module %s not found\n", mname)
		return nil
	}
	tree := ttcn3.ParseFile(file)
	var ret []protocol.CompletionItem
	for _, v := range getAllComponentTypesFromModule(tree, mname) {
		item := protocol.CompletionItem{
			Label:  v,
			Kind:   protocol.StructCompletion,
			Detail: mname + "." + v,
		}
		if sortPref != "" {
			item.SortText = sortPref + v
		}
		ret = append(ret, item)
	}
	return ret
}

func newAllPortTypesFromModule(suite *Suite, mname string, sortPref string) []protocol.CompletionItem {
	file, err := suite.FindModule(mname)
	if err != nil {
		log.Debugf("module %s not found\n", mname)
		return nil
	}
	tree := ttcn3.ParseFile(file)

	var ret []protocol.CompletionItem
	for _, v := range getAllPortTypesFromModule(tree, mname) {
		item := protocol.CompletionItem{
			Label:  v,
			Kind:   protocol.InterfaceCompletion,
			Detail: mname + "." + v,
		}
		if sortPref != "" {
			item.SortText = sortPref + v
		}
		ret = append(ret, item)
	}
	return ret
}

func newAllComponentTypes(suite *Suite, sortPref string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		items := newAllComponentTypesFromModule(suite, baseName(f), sortPref)
		ret = append(ret, items...)
	}
	return ret
}

func newAllPortTypes(suite *Suite, ownModName string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		mName := baseName(f)
		prefix := " 2"
		if mName == ownModName {
			prefix = " 1"
		}
		items := newAllPortTypesFromModule(suite, mName, prefix)
		ret = append(ret, items...)
	}
	return ret
}

func newAllTypes(suite *Suite, ownModName string) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		mName := baseName(f)
		prefix := " 2"
		if mName == ownModName {
			prefix = " 1"
		}
		items := newAllTypesFromModule(suite, mName, prefix)
		ret = append(ret, items...)
	}
	ret = append(ret, newPredefinedTypes()...)
	return ret
}

func newAllValueDecls(suite *Suite, kind syntax.Kind) []protocol.CompletionItem {
	var ret []protocol.CompletionItem
	for _, f := range suite.Files() {
		mname := baseName(f)
		file, err := suite.FindModule(mname)
		if err != nil {
			log.Debugf("module %s not found\n", mname)
			return nil
		}
		tree := ttcn3.ParseFile(file)
		items := newValueDeclsFromModule(tree, baseName(f), kind, true)
		ret = append(ret, items...)
	}
	return ret
}

func baseName(name string) string {
	name = filepath.Base(name)
	name = name[:len(name)-len(filepath.Ext(name))]
	return name
}

func isBehaviourBodyScope(nodes []syntax.Node) bool {
	insideBehav := false
	for _, node := range nodes {
		switch node.(type) {
		case *syntax.FuncDecl:
			insideBehav = true
		case *syntax.BlockStmt:
			if insideBehav {
				return true
			}
		}
	}
	return false
}
func isControlBodyScope(nodes []syntax.Node) bool {
	insideControl := false
	for _, node := range nodes {
		switch node.(type) {
		case *syntax.ControlPart:
			insideControl = true
		case *syntax.BlockStmt:
			if insideControl {
				return true
			}
		}
	}
	return false
}

func isConstDeclScope(nodes []syntax.Node) bool {
	for i := len(nodes) - 1; i > 0; i-- {
		if _, ok := nodes[i].(*syntax.ValueDecl); ok {
			if _, ok := nodes[i-1].(*syntax.ModuleDef); ok {
				return true
			}
		}
	}
	return false
}

func getConstDeclNode(nodes []syntax.Node) *syntax.ValueDecl {
	for _, n := range nodes {
		if val, ok := n.(*syntax.ValueDecl); ok {
			return val
		}
	}
	return nil
}

func isTemplateDeclScope(nodes []syntax.Node) bool {
	for i := len(nodes) - 1; i > 0; i-- {
		if _, ok := nodes[i].(*syntax.TemplateDecl); ok {
			if _, ok := nodes[i-1].(*syntax.ModuleDef); ok {
				return true
			}
		}
	}
	return false
}

func getTemplateDeclNode(nodes []syntax.Node) *syntax.TemplateDecl {
	for _, n := range nodes {
		if templ, ok := n.(*syntax.TemplateDecl); ok {
			return templ
		}
	}
	return nil
}

func isStartId(n syntax.Expr) bool {
	if id, ok := n.(*syntax.Ident); ok {
		return id.Tok.String() == "start"
	}
	return false
}

func isInsideExpression(nodes []syntax.Node, fromModuleDot bool) bool {
	i := len(nodes)
	if i > 1 {
		i--
		if _, ok := nodes[i].(*syntax.Ident); ok {
			i--
		}
		if fromModuleDot {
			if i == 0 {
				return false
			}
			i--
		}
		switch nodes[i].(type) {
		case *syntax.SelectorExpr:
			return false
		case *syntax.ExprStmt:
			return false
		case *syntax.BlockStmt:
			return false
		default:
			return true
		}
	}
	return false
}

func isStartOpArgument(nodes []syntax.Node) bool {
	for i := len(nodes) - 1; i >= 0; i-- {
		if _, ok := nodes[i].(*syntax.ParenExpr); ok {
			if i >= 1 {
				if n, ok := nodes[i-1].(*syntax.CallExpr); ok {
					if fun, ok := n.Fun.(*syntax.SelectorExpr); ok {
						return isStartId(fun.Sel)
					} else {
						return false
					}
				}
			}
		}
	}
	return false
}
func NewCompListItems(suite *Suite, pos loc.Pos, nodes []syntax.Node, ownModName string) []protocol.CompletionItem {
	var list []protocol.CompletionItem
	l := len(nodes)
	switch {
	case isBehaviourBodyScope(nodes):
		modName := ""
		if n, ok := nodes[l-2].(*syntax.SelectorExpr); ok {
			if n.X == nil {
				break
			}
			modName = n.X.LastTok().String()
		}

		kinds := []syntax.Kind{syntax.FUNCTION}
		attrs := []BehavAttrib{WITH_RETURN}

		switch {
		case isStartOpArgument(nodes):
			attrs = append(attrs, WITH_RUNSON)
		case !isInsideExpression(nodes, modName != ""): // less restrictive than isStartOpArgument
			kinds = append(kinds, syntax.ALTSTEP)
			attrs = append(attrs, NONE, WITH_RUNSON)
		}

		if modName != "" {
			file, err := suite.FindModule(modName)
			if err != nil {
				log.Debugf("module %s not found\n", modName)
				return nil
			}
			tree := ttcn3.ParseFile(file)
			list = newAllBehavioursFromModule(tree, kinds, attrs, modName, " 1")
		} else {
			list = newAllBehaviours(suite, kinds, attrs, ownModName)
			list = append(list, newPredefinedFunctions()...)
			list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
		}

	case isControlBodyScope(nodes):

		kinds := []syntax.Kind{syntax.FUNCTION}
		attrs := []BehavAttrib{WITH_RETURN}

		switch n := nodes[l-2].(type) {
		case *syntax.SelectorExpr:
			if n.X == nil {
				break
			}
			modName := n.X.LastTok().String()

			if !isInsideExpression(nodes, true) { // less restrictive than isStartOpArgument
				attrs = append(attrs, NONE)
			}
			file, err := suite.FindModule(modName)
			if err != nil {
				log.Debugf("module %s not found\n", modName)
				return nil
			}
			tree := ttcn3.ParseFile(file)
			list = newAllBehavioursFromModule(tree, kinds, attrs, modName, " 1")
		default:
			if !isInsideExpression(nodes, false) { // less restrictive than isStartOpArgument
				attrs = append(attrs, NONE, WITH_RUNSON)
			}

			list = newAllBehaviours(suite, kinds, attrs, ownModName)
			list = append(list, newPredefinedFunctions()...)
			list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
		}
	case isConstDeclScope(nodes):
		n := getConstDeclNode(nodes)
		if n == nil {
			break
		}

		scndNode, _ := nodes[l-2].(*syntax.SelectorExpr)

		if n.Type == nil || (n.Type != nil && (n.Type.Pos() > pos)) {
			if scndNode != nil && scndNode.X != nil {
				// NOTE: the parser produces a wrong ast under certain circumstances
				// see: func TestTemplateModuleDotType(t *testing.T)
				list = newAllTypesFromModule(suite, scndNode.X.LastTok().String(), "")
			} else {
				list = newAllTypes(suite, ownModName)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		} else {
			switch {
			case scndNode != nil && scndNode.X != nil:
				modName := scndNode.X.LastTok().String()
				file, err := suite.FindModule(modName)
				if err != nil {
					log.Debugf("module %s not found\n", modName)
					return nil
				}
				tree := ttcn3.ParseFile(file)
				list = newAllBehavioursFromModule(tree, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, modName, " 1")
			default:
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		}
	case isTemplateDeclScope(nodes):
		n := getTemplateDeclNode(nodes)
		if n == nil {
			break
		}

		scndNode, _ := nodes[l-2].(*syntax.SelectorExpr)

		switch {
		case n.ModifiesTok != nil && n.AssignTok.Pos() > pos:
			if scndNode != nil && scndNode.X != nil {
				mname := scndNode.X.LastTok().String()
				file, err := suite.FindModule(mname)
				if err != nil {
					log.Debugf("module %s not found\n", mname)
					return nil
				}
				tree := ttcn3.ParseFile(file)
				list = newValueDeclsFromModule(tree, mname, n.TemplateTok.Kind(), false)
			} else {
				list = newAllValueDecls(suite, syntax.TEMPLATE)
				list = append(list, moduleNameListFromSuite(suite, ownModName, "")...)
			}
		case n.Type == nil || n.Name == nil || (n.Name != nil && (n.Name.Pos() > pos)):
			if scndNode != nil && scndNode.X != nil {
				// NOTE: the parser produces a wrong ast under certain circumstances
				// see: func TestTemplateModuleDotType(t *testing.T)
				list = newAllTypesFromModule(suite, scndNode.X.LastTok().String(), "")
			} else {
				list = newAllTypes(suite, ownModName)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		default:
			switch {
			case scndNode != nil && scndNode.X != nil:
				modName := scndNode.X.LastTok().String()
				file, err := suite.FindModule(modName)
				if err != nil {
					log.Debugf("module %s not found\n", modName)
					return nil
				}
				tree := ttcn3.ParseFile(file)
				list = newAllBehavioursFromModule(tree, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, modName, " 1")
			default:
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		}
	default:
		switch n := nodes[l-1].(type) {
		case *syntax.Ident:
			if _, ok := nodes[0].(*syntax.Module); l == 2 && ok {
				list = newModuleDefKw()
			}
			if l <= 2 {
				break
			}

			switch scndNode := nodes[l-2].(type) {
			case *syntax.ModuleDef:
				list = newModuleDefKw()
			case *syntax.ImportDecl:
				switch {
				case scndNode.LBrace != nil:
					list = newImportkinds()
				case n.End() >= pos:
					// look for available modules for import
					list = moduleNameListFromSuite(suite, ownModName, " ")
				default:
					list = newImportAfterModName()
				}
			case *syntax.DefKindExpr:
				// happens after
				// * the altstep/function/testcase kw while typing the identifier
				// * inside the exception list after { while typing the kind
				if l == 8 {
					if _, ok := nodes[l-3].(*syntax.ExceptExpr); ok {
						if scndNode.Kind != nil {
							if impDecl, ok := nodes[l-5].(*syntax.ImportDecl); ok {
								list = newImportCompletions(suite, scndNode.Kind.Kind(), impDecl.Module.Tok.String())
							}
						} else {
							list = newImportkinds()
						}
					}
					break
				}
				if impDecl, ok := nodes[l-3].(*syntax.ImportDecl); ok {
					list = newImportCompletions(suite, scndNode.Kind.Kind(), impDecl.Module.Tok.String())
				}

			case *syntax.ExceptExpr:
				list = newImportkinds()
			case *syntax.RunsOnSpec, *syntax.SystemSpec:
				list = newAllComponentTypes(suite, " 1")
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
			case *syntax.SelectorExpr:
				if scndNode.X != nil {
					switch nodes[l-3].(type) {
					case *syntax.RunsOnSpec, *syntax.SystemSpec, *syntax.ComponentTypeDecl:
						list = newAllComponentTypesFromModule(suite, scndNode.X.LastTok().String(), " 1")
					}
				}
			case *syntax.ComponentTypeDecl:
				// for ctrl+spc, after beginning to type an id after extends Token
				if scndNode.ExtendsTok.LastTok() != nil && scndNode.Body.LBrace.Pos() > pos {
					list = newAllComponentTypes(suite, " 1")
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
				}
			}

		case *syntax.ImportDecl:
			if n.Module == nil {
				// look for available modules for import
				list = moduleNameListFromSuite(suite, ownModName, " ")
			}
		case *syntax.DefKindExpr:
			if n.Kind == nil {
				list = newImportkinds()
			} else {
				if impDecl, ok := nodes[l-2].(*syntax.ImportDecl); ok {
					list = newImportCompletions(suite, n.Kind.Kind(), impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-2].(*syntax.ExceptExpr); ok {
					if impDecl, ok := nodes[l-4].(*syntax.ImportDecl); ok {
						list = newImportCompletions(suite, n.Kind.Kind(), impDecl.Module.Tok.String())
					}
				}
			}
		case *syntax.RunsOnSpec, *syntax.SystemSpec:
			list = newAllComponentTypes(suite, " 1")
			list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
		case *syntax.ErrorNode:
			// i.e. user started typing => syntax.Ident might be detected instead of a kw
			if l <= 1 {
				break
			}
			switch scndNode := nodes[l-2].(type) {
			case *syntax.ModuleDef:
				// start a new module def
				list = newModuleDefKw()
			case *syntax.DefKindExpr:
				// NOTE: not able to reproduce this situation. Maybe it is safe to remove this code.
				// happens streight after the altstep kw if ctrl+space is pressed
				if impDecl, ok := nodes[l-3].(*syntax.ImportDecl); ok {
					list = newImportCompletions(suite, scndNode.Kind.Kind(), impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-3].(*syntax.ExceptExpr); ok {
					if impDecl, ok := nodes[l-5].(*syntax.ImportDecl); ok {
						list = newImportCompletions(suite, scndNode.Kind.Kind(), impDecl.Module.Tok.String())
					}
				}
			}
		case *syntax.Declarator:
			if l <= 2 {
				break
			}
			if valueDecl, ok := nodes[l-2].(*syntax.ValueDecl); ok {
				if valueDecl.Kind.Kind() == syntax.PORT {
					list = newAllPortTypes(suite, ownModName)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		default:
			log.Debugf("Node not considered yet: %#v\n", n)
		}
	}
	return list
}

// LastNonWsToken returns the last node slice before the given position.
func LastNonWsToken(n syntax.Node, pos loc.Pos) []syntax.Node {

	if pos == loc.NoPos {
		return nil
	}

	var (
		completed bool
		isError   bool

		nodeStack, lastStack []syntax.Node
	)

	nodeStack = append(nodeStack, n)
	n.Inspect(func(n syntax.Node) bool {
		if isError {
			return false
		}

		// called on node exit
		if n == nil {
			if len(nodeStack) > 0 {
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			return false
		}

		if errNode, ok := n.(*syntax.ErrorNode); ok {
			if errNode.LastTok().End() == pos {
				isError = true
				nodeStack = append(nodeStack, n)
				lastStack = make([]syntax.Node, len(nodeStack))
				copy(lastStack, nodeStack)
				return false
			}
		}

		// We don't need to descend any deeper if we're passt the
		// position.
		if pos < n.Pos() {
			completed = true
			return false
		}

		nodeStack = append(nodeStack, n)
		lastStack = make([]syntax.Node, len(nodeStack))
		copy(lastStack, nodeStack)
		return !completed
	})

	return lastStack
}

func (s *Server) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Debugln(fmt.Sprintf("Completion took %s.", elapsed))
	}()
	defer func() {
		if err := recover(); err != nil {
			// in case of a panic, just continue as this might be a common situation during typing
			log.Debugln(fmt.Sprintf("Info: %s.", err))
		}
	}()
	if !params.TextDocument.URI.SpanURI().IsFile() {
		log.Printf(fmt.Sprintf("for 'code completion' the new file %q needs to be saved at least once", string(params.TextDocument.URI)))
		return &protocol.CompletionList{}, nil
	}

	defaultModuleId := baseName(params.TextDocument.URI.SpanURI().Filename())

	tree := ttcn3.ParseFile(params.TextDocument.URI.SpanURI().Filename())
	if tree.Root == nil || len(tree.Modules()) == 0 {
		ret := []protocol.CompletionItem{{
			Label:            "module",
			InsertText:       "module ${1:" + defaultModuleId + "} {\n\t${0}\n}",
			InsertTextFormat: protocol.SnippetTextFormat,
			Kind:             protocol.KeywordCompletion,
		}}
		return &protocol.CompletionList{IsIncomplete: false, Items: ret}, nil
	}

	pos := tree.PosFor(int(params.TextDocumentPositionParams.Position.Line+1), int(params.TextDocumentPositionParams.Position.Character+1))
	nodeStack := LastNonWsToken(tree.Root, pos)
	if len(nodeStack) == 0 {
		return nil, nil
	}

	// NOTE: having the current file owned by more then one suite should not
	// import from modules originating from both suites. This would
	// in most ways end up with cyclic imports.
	// Thus 'completion' shall collect items only from one suite.
	// Decision: first suite
	suites := s.Owners(params.TextDocument.URI)
	if len(suites) == 0 {
		return nil, fmt.Errorf("no suite found for file %q", string(params.TextDocument.URI))
	}
	return &protocol.CompletionList{IsIncomplete: false, Items: NewCompListItems(suites[0], pos, nodeStack, defaultModuleId)}, nil
}
