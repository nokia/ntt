package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

var (
	moduleDefKw               = []string{"import from ", "type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control ", "signature "}
	importAfterModName        = []string{"all [except {}];", "{}"}
	importAfterModNameSnippet = []string{"${1:all${2: except {$3\\}}};$0", "{$0}"}
	importKinds               = []string{"type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control", "signature "}
	predefinedTypes           = []string{"anytype ", "bitstring ", "boolean ", "charstring ", "default ", "float ", "hexstring ", "integer ", "octetstring ", "universal charstring ", "verdicttype "}
)

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

func getAllBehavioursFromModule(suite *ntt.Suite, kind token.Kind, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Kind.Kind == kind {
					list = append(list, node.Name.String())
				}
				return false
			default:
				return true
			}

		})
	}
	log.Debug(fmt.Sprintf("AltstepCompletion List :%#v", list))
	return list
}

func getAllValueDeclsFromModule(suite *ntt.Suite, mname string, kind token.Kind) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.FuncDecl, *ast.ComponentTypeDecl:
				// do not descent into TESTCASE, FUNCTION, ALTSTEP,
				// component type
				return false
			case *ast.ValueDecl:
				if node.Kind.Kind != kind {
					return false
				}
				return true
			case *ast.Declarator:
				list = append(list, node.Name.String())
				return false
			case *ast.TemplateDecl:
				if kind == token.TEMPLATE {
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

func getAllTypesFromModule(suite *ntt.Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.BehaviourTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *ast.ComponentTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *ast.EnumTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *ast.PortTypeDecl:
				list = append(list, node.Name.String())
				return false
			case *ast.StructTypeDecl:
				list = append(list, node.Name.String())
				return true
			case *ast.SubTypeDecl:
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

func getAllComponentTypesFromModule(suite *ntt.Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.ComponentTypeDecl:
				list = append(list, node.Name.String())
				return false
			default:
				return true
			}
		})
	}
	return list
}

func getAllPortTypesFromModule(suite *ntt.Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.PortTypeDecl:
				list = append(list, node.Name.String())
				return false
			default:
				return true
			}
		})
	}
	return list
}

func newImportBehaviours(suite *ntt.Suite, kind token.Kind, mname string) []protocol.CompletionItem {
	items := getAllBehavioursFromModule(suite, kind, mname)
	complList := make([]protocol.CompletionItem, 0, len(items)+1)
	for _, v := range items {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.FunctionCompletion})
	}
	complList = append(complList, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return complList
}

func newValueDeclsFromModule(suite *ntt.Suite, mname string, kind token.Kind, withDetail bool) []protocol.CompletionItem {
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

func newImportValueDecls(suite *ntt.Suite, mname string, kind token.Kind) []protocol.CompletionItem {
	complList := newValueDeclsFromModule(suite, mname, kind, false)
	complList = append(complList, protocol.CompletionItem{Label: "all;", Kind: protocol.KeywordCompletion})
	return complList
}

func newImportTypes(suite *ntt.Suite, mname string) []protocol.CompletionItem {
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

func newImportCompletions(suite *ntt.Suite, kind token.Kind, mname string) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	switch kind {
	case token.ALTSTEP, token.FUNCTION, token.TESTCASE:
		list = newImportBehaviours(suite, kind, mname)
	case token.TEMPLATE, token.CONST, token.MODULEPAR:
		list = newImportValueDecls(suite, mname, kind)
	case token.TYPE:
		list = newImportTypes(suite, mname)
	default:
		log.Debug(fmt.Sprintf("Kind not considered yet: %#v)", kind))
	}
	return list
}

func moduleNameListFromSuite(suite *ntt.Suite, ownModName string, sortPref string) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	if files := suite.FindAllFiles(); len(files) > 0 {
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

func newAllTypesFromModule(suite *ntt.Suite, modName string, sortPref string) []protocol.CompletionItem {
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

func newAllComponentTypesFromModule(suite *ntt.Suite, modName string, sortPref string) []protocol.CompletionItem {
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

func newAllPortTypesFromModule(suite *ntt.Suite, modName string, sortPref string) []protocol.CompletionItem {
	items := getAllPortTypesFromModule(suite, modName)
	portList := make([]protocol.CompletionItem, 0, len(items))
	for _, v := range items {
		if len(sortPref) > 0 {
			portList = append(portList, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion, SortText: sortPref + v, Detail: modName + "." + v})
		} else {
			portList = append(portList, protocol.CompletionItem{Label: v, Kind: protocol.StructCompletion, Detail: modName + "." + v})
		}
	}
	return portList
}

func newAllComponentTypes(suite *ntt.Suite, sortPref string) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.FindAllFiles(); len(files) > 0 {
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

func newAllPortTypes(suite *ntt.Suite, ownModName string) []protocol.CompletionItem {
	var portList []protocol.CompletionItem = nil
	if files := suite.FindAllFiles(); len(files) > 0 {
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

func newAllTypes(suite *ntt.Suite, ownModName string) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.FindAllFiles(); len(files) > 0 {
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

func newAllValueDecls(suite *ntt.Suite, kind token.Kind) []protocol.CompletionItem {
	var complList []protocol.CompletionItem = nil
	if files := suite.FindAllFiles(); len(files) > 0 {
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

func NewCompListItems(suite *ntt.Suite, pos loc.Pos, nodes []ast.Node, ownModName string) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	l := len(nodes)
	if nodes == nil || l == 0 {
		return make([]protocol.CompletionItem, 0)
	}
	switch nodet := nodes[l-1].(type) {
	case *ast.Ident:
		if _, ok := nodes[0].(*ast.Module); l == 2 && ok {
			list = newModuleDefKw()
		}
		if l > 2 {
			switch scndNode := nodes[l-2].(type) {
			case *ast.ModuleDef:
				list = newModuleDefKw()
			case *ast.ImportDecl:
				if scndNode.LBrace.IsValid() {
					list = newImportkinds()
				} else if nodet.End() >= pos {
					// look for available modules for import
					list = moduleNameListFromSuite(suite, ownModName, " ")
				} else {
					list = newImportAfterModName()
				}
			case *ast.DefKindExpr:
				// happens after
				// * the altstep/function/testcase kw while typing the identifier
				// * inside the exception list after { while typing the kind
				if l == 7 {
					if _, ok := nodes[l-3].(*ast.ExceptExpr); ok {
						if scndNode.Kind.IsValid() {
							if impDecl, ok := nodes[l-5].(*ast.ImportDecl); ok {
								list = newImportCompletions(suite, scndNode.Kind.Kind, impDecl.Module.Tok.String())
							}
						} else {
							list = newImportkinds()
						}
					}
				} else {
					if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
						list = newImportCompletions(suite, scndNode.Kind.Kind, impDecl.Module.Tok.String())
					}
				}

			case *ast.ExceptExpr:
				list = newImportkinds()
			case *ast.RunsOnSpec, *ast.SystemSpec:
				list = newAllComponentTypes(suite, " 1")
				list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
			case *ast.SelectorExpr:
				if scndNode.X != nil {
					switch thrdNode := nodes[l-3].(type) {
					case *ast.RunsOnSpec, *ast.SystemSpec, *ast.ComponentTypeDecl:
						list = newAllComponentTypesFromModule(suite, scndNode.X.LastTok().String(), " 1")
					case *ast.TemplateDecl:
						if thrdNode.ModifiesTok.IsValid() {
							list = newValueDeclsFromModule(suite, scndNode.X.LastTok().String(), thrdNode.TemplateTok.Kind, false)
						} else /*if thrdNode.Name == nil */ {
							// NOTE: the parser produces a wrong ast under certain circumstances
							// see: func TestTemplateModuleDotType(t *testing.T)
							list = newAllTypesFromModule(suite, scndNode.X.LastTok().String(), "")
						}
					}
				}
			case *ast.ComponentTypeDecl:
				// for ctrl+spc, after beginning to type an id after extends Token
				if scndNode.ExtendsTok.LastTok().IsValid() && scndNode.Body.LBrace.Pos() > pos {
					list = newAllComponentTypes(suite, " 1")
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
				}
			case *ast.TemplateDecl:
				if scndNode.ModifiesTok.LastTok().IsValid() && scndNode.AssignTok.Pos() > pos {
					list = newAllValueDecls(suite, token.TEMPLATE)
					list = append(list, moduleNameListFromSuite(suite, ownModName, "")...)
				} else if scndNode.Name == nil {
					list = newAllTypes(suite, ownModName)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		}

	case *ast.ImportDecl:
		if nodet.Module == nil {
			// look for available modules for import
			list = moduleNameListFromSuite(suite, ownModName, " ")
		}
	case *ast.DefKindExpr:
		if !nodet.Kind.IsValid() {
			list = newImportkinds()
		} else {
			if impDecl, ok := nodes[l-2].(*ast.ImportDecl); ok {
				list = newImportCompletions(suite, nodet.Kind.Kind, impDecl.Module.Tok.String())
			} else if _, ok := nodes[l-2].(*ast.ExceptExpr); ok {
				if impDecl, ok := nodes[l-4].(*ast.ImportDecl); ok {
					list = newImportCompletions(suite, nodet.Kind.Kind, impDecl.Module.Tok.String())
				}
			}
		}
	case *ast.RunsOnSpec, *ast.SystemSpec:
		list = newAllComponentTypes(suite, " 1")
		list = append(list, moduleNameListFromSuite(suite, ownModName, " 2")...)
	case *ast.TemplateDecl:
		if nodet.ModifiesTok.LastTok().IsValid() && nodet.AssignTok.Pos() > pos {
			list = newAllValueDecls(suite, token.TEMPLATE)
			list = append(list, moduleNameListFromSuite(suite, ownModName, "")...)
		} else if nodet.Type == nil || nodet.Name == nil {
			list = newAllTypes(suite, ownModName)
			list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
		}
	case *ast.ErrorNode:
		// i.e. user started typing => ast.Ident might be detected instead of a kw
		if l > 1 {
			switch scndNode := nodes[l-2].(type) {
			case *ast.ModuleDef:
				// start a new module def
				list = newModuleDefKw()
			case *ast.DefKindExpr:
				// NOTE: not able to reproduce this situation. Maybe it is safe to remove this code.
				// happens streight after the altstep kw if ctrl+space is pressed
				if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
					list = newImportCompletions(suite, scndNode.Kind.Kind, impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-3].(*ast.ExceptExpr); ok {
					if impDecl, ok := nodes[l-5].(*ast.ImportDecl); ok {
						list = newImportCompletions(suite, scndNode.Kind.Kind, impDecl.Module.Tok.String())
					}
				}
			case *ast.TemplateDecl:
				if scndNode.Type == nil || scndNode.Name == nil {
					list = newAllTypes(suite, ownModName)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		}
	case *ast.Declarator:
		if l > 2 {
			if valueDecl, ok := nodes[l-2].(*ast.ValueDecl); ok {
				if valueDecl.Kind.Kind == token.PORT {
					list = newAllPortTypes(suite, ownModName)
					list = append(list, moduleNameListFromSuite(suite, ownModName, " 3")...)
				}
			}
		}
	default:
		log.Debug(fmt.Sprintf("Node not considered yet: %#v)", nodet))

	}
	return list
}

func LastNonWsToken(n ast.Node, pos loc.Pos) []ast.Node {
	var (
		completed bool       = false
		nodeStack []ast.Node = make([]ast.Node, 0, 10)
		lastStack []ast.Node = nil
		isError   bool       = false
	)

	ast.Inspect(n, func(n ast.Node) bool {
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
		log.Debug(fmt.Sprintf("looking for %d In node[%d .. %d] (node: %#v)", pos, n.Pos(), n.End(), n))
		if errNode, ok := n.(*ast.ErrorNode); ok {
			if errNode.LastTok().End() == pos {
				isError = true
				nodeStack = append(nodeStack, n)
				lastStack = make([]ast.Node, len(nodeStack))
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
		lastStack = make([]ast.Node, len(nodeStack))
		copy(lastStack, nodeStack)
		return !completed
	})
	log.Debug(fmt.Sprintf("Completion at lastNode :%#v NodeStack: %#v", lastStack[len(lastStack)-1], lastStack))
	return lastStack
}

func (s *Server) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("Completion took %s.", elapsed))
	}()
	defer func() {
		if err := recover(); err != nil {
			// in case of a panic, just continue as this might be a common situation during typing
			log.Debug(fmt.Sprintf("Info: %s.", err))
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
	syntax := suites[0].ParseWithAllErrors(params.TextDocument.URI.SpanURI().Filename())
	log.Debug(fmt.Sprintf("Completion after Parse :%p", &syntax.Module))
	if syntax.Module == nil {
		return nil, syntax.Err
	}

	if syntax.Module.Name == nil {
		complList := make([]protocol.CompletionItem, 1)
		complList = append(complList, protocol.CompletionItem{Label: "module",
			InsertText:       "module ${1:" + defaultModuleId + "} {\n\t${0}\n}",
			InsertTextFormat: protocol.SnippetTextFormat, Kind: protocol.KeywordCompletion})
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("Completion took %s.", elapsed))

		return &protocol.CompletionList{IsIncomplete: false, Items: complList}, nil
	}
	pos := syntax.Pos(int(params.TextDocumentPositionParams.Position.Line+1), int(params.TextDocumentPositionParams.Position.Character+1))
	nodeStack := LastNonWsToken(syntax.Module, pos)

	return &protocol.CompletionList{IsIncomplete: false, Items: NewCompListItems(suites[0], pos, nodeStack, defaultModuleId)}, nil
}
