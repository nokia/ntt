package lsp

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/nokia/ntt/ttcn3/printer"
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
	complList := make([]protocol.CompletionItem, 0, len(PredefinedFunctions))
	for _, v := range PredefinedFunctions {
		markup := protocol.MarkupContent{Kind: "markdown", Value: v.Documentation}
		complList = append(complList, protocol.CompletionItem{
			Label: v.Label, Kind: protocol.FunctionCompletion,
			Detail:           v.Signature,
			InsertTextFormat: v.TextFormat,
			InsertText:       v.InsertText,
			Documentation:    markup})
	}
	return complList
}

func newImportkinds() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(importKinds))
	for _, v := range importKinds {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return complList
}

func newPredefinedTypes() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(predefinedTypes))
	for _, v := range predefinedTypes {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return complList
}

func getAllBehavioursFromModule(suite *Suite, kind syntax.Kind, mname string) []*FunctionDetails {
	list := make([]*FunctionDetails, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		tree := ttcn3.ParseFile(file)
		tree.Root.Inspect(func(n syntax.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *syntax.FuncDecl:
				if node.Kind.Kind() == kind {
					var sig bytes.Buffer
					textFormat := protocol.PlainTextTextFormat
					sig.WriteString(node.Kind.String() + " " + mname + "." + node.Name.String())
					len1 := len(sig.String())
					printer.Print(&sig, tree.FileSet, node.Params)
					hasParams := (len(sig.String()) - len1) > 2
					if hasParams {
						textFormat = protocol.SnippetTextFormat
					}
					if node.RunsOn != nil {
						sig.WriteString("\n  ")
						printer.Print(&sig, tree.FileSet, node.RunsOn)
					}
					if node.System != nil {
						sig.WriteString("\n  ")
						printer.Print(&sig, tree.FileSet, node.System)
					}
					if node.Return != nil {
						sig.WriteString("\n  ")
						printer.Print(&sig, tree.FileSet, node.Return)
					}
					list = append(list, &FunctionDetails{
						Label:         node.Name.String(),
						HasRunsOn:     (node.RunsOn != nil),
						HasReturn:     (node.Return != nil),
						Signature:     sig.String(),
						Documentation: syntax.Doc(tree.FileSet, node),
						HasParameters: hasParams,
						TextFormat:    textFormat})
				}
				return false
			default:
				return true
			}

		})
	}
	log.Debugln(fmt.Sprintf("AltstepCompletion List :%#v", list))
	return list
}

func getAllValueDeclsFromModule(suite *Suite, mname string, kind syntax.Kind) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		tree := ttcn3.ParseFile(file)
		tree.Root.Inspect(func(n syntax.Node) bool {
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
				list = append(list, node.Name.String())
				return false
			case *syntax.TemplateDecl:
				if kind == syntax.TEMPLATE {
					list = append(list, node.Name.String())
				}
				return false
			default:
				return true
			}
		})
	}
	return list
}

func getAllTypesFromModule(suite *Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		tree := ttcn3.ParseFile(file)
		tree.Root.Inspect(func(n syntax.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *syntax.BehaviourTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *syntax.ComponentTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *syntax.EnumTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *syntax.PortTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *syntax.StructTypeDecl:
				list = append(list, node.Name.String())
				return true
			case *syntax.SubTypeDecl:
				// for typpe defs as well as for record of/set of types
				list = append(list, node.Field.Name.String())
				return false
			default:
				return true
			}
		})
	}
	return list
}

func getAllComponentTypesFromModule(suite *Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		tree := ttcn3.ParseFile(file)
		tree.Root.Inspect(func(n syntax.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *syntax.ComponentTypeDecl:
				list = append(list, node.Name.String())
				return false
			default:
				return true
			}
		})
	}
	return list
}

func getAllPortTypesFromModule(suite *Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		tree := ttcn3.ParseFile(file)
		tree.Root.Inspect(func(n syntax.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *syntax.PortTypeDecl:
				list = append(list, node.Name.String())
				return false
			default:
				return true
			}
		})
	}
	return list
}

func newImportBehaviours(suite *Suite, kind syntax.Kind, mname string) []protocol.CompletionItem {
	items := getAllBehavioursFromModule(suite, kind, mname)
	complList := make([]protocol.CompletionItem, 0, len(items)+1)
	for _, v := range items {
		complList = append(complList, protocol.CompletionItem{Label: v.Label, Kind: protocol.FunctionCompletion})
	}
	complList = append(complList, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return complList
}
func newAllBehavioursFromModule(suite *Suite, kinds []syntax.Kind, attribs []BehavAttrib, mname string, sortPref string) []protocol.CompletionItem {

	complList := make([]protocol.CompletionItem, 0, 10)
	var items []*FunctionDetails

	for _, kind := range kinds {
		items = append(items, getAllBehavioursFromModule(suite, kind, mname)...)
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
			complList = append(complList, protocol.CompletionItem{Label: v.Label + "()",
				InsertText: insertText,
				Kind:       protocol.FunctionCompletion, SortText: sortPref + v.Label,
				Detail: v.Signature, Documentation: v.Documentation, InsertTextFormat: v.TextFormat})
		}
	}
	return complList
}
func newAllBehaviours(suite *Suite, kinds []syntax.Kind, attribs []BehavAttrib, mname string) []protocol.CompletionItem {
	var sortPref string

	if files := suite.Files(); len(files) > 0 {
		complList := make([]protocol.CompletionItem, 0, len(files)*2)

		for _, f := range files {
			fileName := filepath.Base(f)
			fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
			if fileName != mname {
				sortPref = " 2"
			} else {
				sortPref = " 1"
			}
			complList = append(complList, newAllBehavioursFromModule(suite, kinds, attribs, fileName, sortPref)...)
		}
		return complList
	}
	return nil
}

func newValueDeclsFromModule(suite *Suite, mname string, kind syntax.Kind, withDetail bool) []protocol.CompletionItem {
	items := getAllValueDeclsFromModule(suite, mname, kind)
	complList := make([]protocol.CompletionItem, 0, len(items)+1)
	for _, v := range items {
		if withDetail {
			complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.ConstantCompletion, Detail: mname + "." + v})
		} else {
			complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.ConstantCompletion})
		}
	}
	return complList
}

func newImportValueDecls(suite *Suite, mname string, kind syntax.Kind) []protocol.CompletionItem {
	complList := newValueDeclsFromModule(suite, mname, kind, false)
	complList = append(complList, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return complList
}

func newImportTypes(suite *Suite, mname string) []protocol.CompletionItem {
	items := getAllTypesFromModule(suite, mname)
	complList := make([]protocol.CompletionItem, 0, len(items)+1)
	// NOTE: instead of 'StructCompletion' a better matching kind could be used
	for _, v := range items {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion})
	}
	complList = append(complList, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return complList
}

func newModuleDefKw() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(moduleDefKw))
	for _, v := range moduleDefKw {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return complList
}

func newImportAfterModName() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(importAfterModName))
	for i, v := range importAfterModName {
		complList = append(complList, protocol.CompletionItem{Label: v,
			InsertText:       importAfterModNameSnippet[i],
			InsertTextFormat: protocol.SnippetTextFormat, Kind: protocol.KeywordCompletion})
	}
	return complList
}

func newImportCompletions(suite *Suite, kind syntax.Kind, mname string) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	switch kind {
	case syntax.ALTSTEP, syntax.FUNCTION, syntax.TESTCASE:
		list = newImportBehaviours(suite, kind, mname)
	case syntax.TEMPLATE, syntax.CONST, syntax.MODULEPAR:
		list = newImportValueDecls(suite, mname, kind)
	case syntax.TYPE:
		list = newImportTypes(suite, mname)
	default:
		log.Debugln(fmt.Sprintf("Kind not considered yet: %#v)", kind))
	}
	return list
}

func getModuleNameSetFromSuite(suite *Suite) *map[string]struct{} {
	var set map[string]struct{}
	if files := suite.Files(); len(files) > 0 {
		set = make(map[string]struct{}, len(files))
		for _, f := range files {
			fileName := filepath.Base(f)
			fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
			set[fileName] = struct{}{}
		}
	}
	return &set
}

func moduleNameListFromSuite(suite *Suite, ownModName string, sortPref string) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	if files := suite.Files(); len(files) > 0 {
		list = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			fileName := filepath.Base(f)
			fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
			if fileName != ownModName {
				if len(sortPref) > 0 {
					list = append(list, protocol.CompletionItem{Label: fileName, Kind: protocol.ModuleCompletion, SortText: sortPref + fileName})
				} else {
					list = append(list, protocol.CompletionItem{Label: fileName, Kind: protocol.ModuleCompletion})
				}
			}
		}
	}
	return list
}

func newAllTypesFromModule(suite *Suite, modName string, sortPref string) []protocol.CompletionItem {
	items := getAllTypesFromModule(suite, modName)
	complList := make([]protocol.CompletionItem, 0, len(items))
	for _, v := range items {
		if len(sortPref) > 0 {
			complList = append(complList, protocol.CompletionItem{Label: v + " ", Kind: protocol.StructCompletion, SortText: sortPref + v, Detail: modName + "." + v})
		} else {
			complList = append(complList, protocol.CompletionItem{Label: v + " ", Kind: protocol.StructCompletion, Detail: modName + "." + v})
		}
	}
	return complList
}

func newAllComponentTypesFromModule(suite *Suite, modName string, sortPref string) []protocol.CompletionItem {
	items := getAllComponentTypesFromModule(suite, modName)
	complList := make([]protocol.CompletionItem, 0, len(items))
	for _, v := range items {
		if len(sortPref) > 0 {
			complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion, SortText: sortPref + v, Detail: modName + "." + v})
		} else {
			complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion, Detail: modName + "." + v})
		}
	}
	return complList
}

func newAllPortTypesFromModule(suite *Suite, modName string, sortPref string) []protocol.CompletionItem {
	items := getAllPortTypesFromModule(suite, modName)
	portList := make([]protocol.CompletionItem, 0, len(items))
	for _, v := range items {
		if len(sortPref) > 0 {
			portList = append(portList, protocol.CompletionItem{Label: v, Kind: protocol.InterfaceCompletion, SortText: sortPref + v, Detail: modName + "." + v})
		} else {
			portList = append(portList, protocol.CompletionItem{Label: v, Kind: protocol.InterfaceCompletion, Detail: modName + "." + v})
		}
	}
	return portList
}

func newAllComponentTypes(suite *Suite, sortPref string) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.Files(); len(files) > 0 {
		complList = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			mName := filepath.Base(f)
			mName = mName[:len(mName)-len(filepath.Ext(mName))]
			items := newAllComponentTypesFromModule(suite, mName, sortPref)
			complList = append(complList, items...)
		}
	}
	return complList
}

func newAllPortTypes(suite *Suite, ownModName string) []protocol.CompletionItem {
	var portList []protocol.CompletionItem = nil
	if files := suite.Files(); len(files) > 0 {
		portList = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			mName := filepath.Base(f)
			mName = mName[:len(mName)-len(filepath.Ext(mName))]
			prefix := int(2)
			if mName == ownModName {
				prefix = 1
			}
			items := newAllPortTypesFromModule(suite, mName, " "+strconv.Itoa(prefix))
			portList = append(portList, items...)
		}
	}
	return portList
}

func newAllTypes(suite *Suite, ownModName string) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.Files(); len(files) > 0 {
		complList = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			mName := filepath.Base(f)
			mName = mName[:len(mName)-len(filepath.Ext(mName))]
			prefix := int(2)
			if mName == ownModName {
				prefix = 1
			}
			items := newAllTypesFromModule(suite, mName, " "+strconv.Itoa(prefix))
			complList = append(complList, items...)
		}
	}
	complList = append(complList, newPredefinedTypes()...)
	return complList
}

func newAllValueDecls(suite *Suite, kind syntax.Kind) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.Files(); len(files) > 0 {
		complList = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			mName := filepath.Base(f)
			mName = mName[:len(mName)-len(filepath.Ext(mName))]
			items := newValueDeclsFromModule(suite, mName, kind, true)
			complList = append(complList, items...)
		}
	}
	return complList
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
	var list []protocol.CompletionItem = nil
	l := len(nodes)
	if nodes == nil || l == 0 {
		return make([]protocol.CompletionItem, 0)
	}
	switch {
	case isBehaviourBodyScope(nodes):
		switch n := nodes[l-2].(type) {
		case *syntax.SelectorExpr:
			if n.X != nil {
				switch {
				case isStartOpArgument(nodes):
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN, WITH_RUNSON}, n.X.LastTok().String(), " 1")
				case isInsideExpression(nodes, true): // less restrictive than isStartOpArgument
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, n.X.LastTok().String(), " 1")
				default:
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION, syntax.ALTSTEP},
						[]BehavAttrib{NONE, WITH_RETURN, WITH_RUNSON}, n.X.LastTok().String(), " 1")
				}
			}
		default:
			switch {
			case isStartOpArgument(nodes):
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN, WITH_RUNSON}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			case isInsideExpression(nodes, false): // less restrictive than isStartOpArgument
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			default:
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION, syntax.ALTSTEP},
					[]BehavAttrib{NONE, WITH_RETURN, WITH_RUNSON}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		}
	case isControlBodyScope(nodes):
		switch n := nodes[l-2].(type) {
		case *syntax.SelectorExpr:
			if n.X != nil {
				switch {
				case isInsideExpression(nodes, true): // less restrictive than isStartOpArgument
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, n.X.LastTok().String(), " 1")
				default:
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{NONE, WITH_RETURN}, n.X.LastTok().String(), " 1")
				}
			}
		default:
			switch {
			case isInsideExpression(nodes, false): // less restrictive than isStartOpArgument
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{WITH_RETURN}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			default:
				list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
					[]BehavAttrib{NONE, WITH_RETURN, WITH_RUNSON}, ownModName)
				list = append(list, newPredefinedFunctions()...)
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
			}
		}
	case isConstDeclScope(nodes):
		if nodec := getConstDeclNode(nodes); nodec != nil {
			scndNode, _ := nodes[l-2].(*syntax.SelectorExpr)

			if nodec.Type == nil || (nodec.Type != nil && (nodec.Type.Pos() > pos)) {
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
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, scndNode.X.LastTok().String(), " 1")
				default:
					list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, ownModName)
					list = append(list, newPredefinedFunctions()...)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		}
	case isTemplateDeclScope(nodes):
		if nodet := getTemplateDeclNode(nodes); nodet != nil {
			scndNode, _ := nodes[l-2].(*syntax.SelectorExpr)

			if nodet.ModifiesTok != nil && nodet.AssignTok.Pos() > pos {
				if scndNode != nil && scndNode.X != nil {
					list = newValueDeclsFromModule(suite, scndNode.X.LastTok().String(), nodet.TemplateTok.Kind(), false)
				} else {
					list = newAllValueDecls(suite, syntax.TEMPLATE)
					list = append(list, moduleNameListFromSuite(suite, ownModName, "")...)
				}
			} else if nodet.Type == nil || nodet.Name == nil || (nodet.Name != nil && (nodet.Name.Pos() > pos)) {
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
					list = newAllBehavioursFromModule(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, scndNode.X.LastTok().String(), " 1")
				default:
					list = newAllBehaviours(suite, []syntax.Kind{syntax.FUNCTION},
						[]BehavAttrib{WITH_RETURN}, ownModName)
					list = append(list, newPredefinedFunctions()...)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		}
	default:
		switch nodet := nodes[l-1].(type) {
		case *syntax.Ident:
			if _, ok := nodes[0].(*syntax.Module); l == 2 && ok {
				list = newModuleDefKw()
			}
			if l > 2 {
				switch scndNode := nodes[l-2].(type) {
				case *syntax.ModuleDef:
					list = newModuleDefKw()
				case *syntax.ImportDecl:
					if scndNode.LBrace != nil {
						list = newImportkinds()
					} else if nodet.End() >= pos {
						// look for available modules for import
						list = moduleNameListFromSuite(suite, ownModName, " ")
					} else {
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
					} else {
						if impDecl, ok := nodes[l-3].(*syntax.ImportDecl); ok {
							list = newImportCompletions(suite, scndNode.Kind.Kind(), impDecl.Module.Tok.String())
						}
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
			}

		case *syntax.ImportDecl:
			if nodet.Module == nil {
				// look for available modules for import
				list = moduleNameListFromSuite(suite, ownModName, " ")
			}
		case *syntax.DefKindExpr:
			if nodet.Kind == nil {
				list = newImportkinds()
			} else {
				if impDecl, ok := nodes[l-2].(*syntax.ImportDecl); ok {
					list = newImportCompletions(suite, nodet.Kind.Kind(), impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-2].(*syntax.ExceptExpr); ok {
					if impDecl, ok := nodes[l-4].(*syntax.ImportDecl); ok {
						list = newImportCompletions(suite, nodet.Kind.Kind(), impDecl.Module.Tok.String())
					}
				}
			}
		case *syntax.RunsOnSpec, *syntax.SystemSpec:
			list = newAllComponentTypes(suite, " 1")
			list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
		case *syntax.ErrorNode:
			// i.e. user started typing => syntax.Ident might be detected instead of a kw
			if l > 1 {
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
			}
		case *syntax.Declarator:
			if l > 2 {
				if valueDecl, ok := nodes[l-2].(*syntax.ValueDecl); ok {
					if valueDecl.Kind.Kind() == syntax.PORT {
						list = newAllPortTypes(suite, ownModName)
						list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
					}
				}
			}
		default:
			log.Debugln(fmt.Sprintf("Node not considered yet: %#v)", nodet))
		}
	}
	return list
}

func LastNonWsToken(n syntax.Node, pos loc.Pos) []syntax.Node {
	// TODO(5nord): Replace this function with n.LastTok() and n.PrevTok()

	var (
		completed bool       = false
		nodeStack []syntax.Node = make([]syntax.Node, 0, 10)
		lastStack []syntax.Node = nil
		isError   bool       = false
	)

	n.Inspect(func(n syntax.Node) bool {
		if isError {
			return false
		}
		if n == nil {
			// called on node exit
			if !isError {
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			return false
		}
		log.Debugln(fmt.Sprintf("looking for %d In node[%d .. %d] (node: %#v)", pos, n.Pos(), n.End(), n))
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
	log.Debugln(fmt.Sprintf("Completion at lastNode :%#v NodeStack: %#v", lastStack[len(lastStack)-1], lastStack))
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

	fileName := filepath.Base(params.TextDocument.URI.SpanURI().Filename())
	defaultModuleId := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	suites := s.Owners(params.TextDocument.URI)
	// NOTE: having the current file owned by more then one suite should not
	// import from modules originating from both suites. This would
	// in most ways end up with cyclic imports.
	// Thus 'completion' shall collect items only from one suite.
	// Decision: first suite
	tree := ttcn3.ParseFile(params.TextDocument.URI.SpanURI().Filename())
	log.Debugln(fmt.Sprintf("Completion after Parse :%p", &tree.Root))
	if tree.Root == nil || len(tree.Modules()) == 0 {
		complList := make([]protocol.CompletionItem, 1)
		complList = append(complList, protocol.CompletionItem{Label: "module",
			InsertText:       "module ${1:" + defaultModuleId + "} {\n\t${0}\n}",
			InsertTextFormat: protocol.SnippetTextFormat, Kind: protocol.KeywordCompletion})
		elapsed := time.Since(start)
		log.Debugln(fmt.Sprintf("Completion took %s.", elapsed))

		return &protocol.CompletionList{IsIncomplete: false, Items: complList}, nil
	}
	pos := tree.Pos(int(params.TextDocumentPositionParams.Position.Line+1), int(params.TextDocumentPositionParams.Position.Character+1))
	nodeStack := LastNonWsToken(tree.Root, pos)

	return &protocol.CompletionList{IsIncomplete: false, Items: NewCompListItems(suites[0], pos, nodeStack, defaultModuleId)}, nil
}
