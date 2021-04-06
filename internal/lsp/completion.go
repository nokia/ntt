package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

var (
	moduleDefKw               = []string{"import from ", "type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control", "signature "}
	importAfterModName        = []string{"all [except {}];", "{}"}
	importAfterModNameSnippet = []string{"${1:all${2: except {$3\\}}};$0", "{$0}"}
	importKinds               = []string{"type ", "const ", "modulepar ", "template ", "function ", "external function ", "altstep ", "testcase ", "control", "signature "}
)

func newImportkinds() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(importKinds))
	for _, v := range importKinds {
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

func getAllTemplatesFromModule(suite *ntt.Suite, mname string) []string {
	list := make([]string, 0, 10)
	if file, err := suite.FindModule(mname); err == nil {
		syntax := suite.Parse(file)
		ast.Inspect(syntax.Module, func(n ast.Node) bool {
			if n == nil {
				// called on node exit
				return false
			}

			switch node := n.(type) {
			case *ast.TemplateDecl:
				list = append(list, node.Name.String())
				return false
			default:
				return true
			}

		})
	}
	log.Debug(fmt.Sprintf("AltstepCompletion List :%#v", list))
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

func newImportTemplates(suite *ntt.Suite, mname string) []protocol.CompletionItem {
	items := getAllTemplatesFromModule(suite, mname)
	complList := make([]protocol.CompletionItem, 0, len(items)+1)
	for _, v := range items {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.ConstantCompletion})
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

func moduleNameListFromSuite(suite *ntt.Suite) []protocol.CompletionItem {
	var list []protocol.CompletionItem = nil
	if files, err := suite.Files(); err == nil {
		list = make([]protocol.CompletionItem, 0, len(files))
		for _, f := range files {
			fileName := filepath.Base(f)
			fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
			list = append(list, protocol.CompletionItem{Label: fileName, Kind: protocol.ModuleCompletion})
		}
	}
	return list
}

func NewCompListItems(suite *ntt.Suite, pos loc.Pos, nodes []ast.Node) []protocol.CompletionItem {
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
			if _, ok := nodes[l-2].(*ast.ModuleDef); ok {
				list = newModuleDefKw()
			}
			if importDecl, ok := nodes[l-2].(*ast.ImportDecl); ok {
				if importDecl.LBrace.IsValid() {
					list = newImportkinds()
				} else if nodet.End() >= pos {
					// look for available modules for import
					list = moduleNameListFromSuite(suite)
				} else {
					list = newImportAfterModName()
				}
			}
			if defKind, ok := nodes[l-2].(*ast.DefKindExpr); ok {
				// happens after
				// * the altstep/function/testcase kw while typing the identifier
				// * inside the exception list after { while typing the kind
				if l == 7 {
					if _, ok := nodes[l-3].(*ast.ExceptExpr); ok {
						if defKind.Kind.IsValid() {
							switch defKind.Kind.Kind {
							case token.ALTSTEP, token.FUNCTION, token.TESTCASE:
								if impDecl, ok := nodes[l-5].(*ast.ImportDecl); ok {
									list = newImportBehaviours(suite, defKind.Kind.Kind, impDecl.Module.Tok.String())
								}
							case token.TEMPLATE:
								if impDecl, ok := nodes[l-5].(*ast.ImportDecl); ok {
									list = newImportTemplates(suite, impDecl.Module.Tok.String())
								}
							}
						} else {
							list = newImportkinds()
						}
					}
				} else {
					switch defKind.Kind.Kind {
					case token.ALTSTEP, token.FUNCTION, token.TESTCASE:
						if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
							list = newImportBehaviours(suite, defKind.Kind.Kind, impDecl.Module.Tok.String())
						}
					case token.TEMPLATE:
						if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
							list = newImportTemplates(suite, impDecl.Module.Tok.String())
						}
					}
				}
			}
			if _, ok := nodes[l-2].(*ast.ExceptExpr); ok {
				list = newImportkinds()
			}
		}

	case *ast.ImportDecl:
		if nodet.Module == nil {
			// look for available modules for import
			list = moduleNameListFromSuite(suite)
		}
	case *ast.DefKindExpr:
		if !nodet.Kind.IsValid() {
			list = newImportkinds()
		} else {
			switch nodet.Kind.Kind {
			case token.ALTSTEP, token.FUNCTION, token.TESTCASE:
				if impDecl, ok := nodes[l-2].(*ast.ImportDecl); ok {
					list = newImportBehaviours(suite, nodet.Kind.Kind, impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-2].(*ast.ExceptExpr); ok {
					if impDecl, ok := nodes[l-4].(*ast.ImportDecl); ok {
						list = newImportBehaviours(suite, nodet.Kind.Kind, impDecl.Module.Tok.String())
					}
				}
			case token.TEMPLATE:
				if impDecl, ok := nodes[l-2].(*ast.ImportDecl); ok {
					list = newImportTemplates(suite, impDecl.Module.Tok.String())
				}
			default:
				log.Debug(fmt.Sprintf("Kind not considered yet: %#v)", nodet.Kind.Kind))
			}
		}
	case *ast.ErrorNode:
		// i.e. user started typing => ast.Ident might be detected instead of a kw
		if _, ok := nodes[l-2].(*ast.ModuleDef); l > 1 && ok {
			// start a new module def
			list = newModuleDefKw()
		} else if defKind, ok := nodes[l-2].(*ast.DefKindExpr); ok {
			// happens streight after the altstep kw if ctrl+space is pressed
			switch defKind.Kind.Kind {
			case token.ALTSTEP, token.FUNCTION, token.TESTCASE:
				if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
					list = newImportBehaviours(suite, defKind.Kind.Kind, impDecl.Module.Tok.String())
				} else if _, ok := nodes[l-3].(*ast.ExceptExpr); ok {
					if impDecl, ok := nodes[l-5].(*ast.ImportDecl); ok {
						list = newImportBehaviours(suite, defKind.Kind.Kind, impDecl.Module.Tok.String())
					}
				}
			case token.TEMPLATE:
				if impDecl, ok := nodes[l-3].(*ast.ImportDecl); ok {
					list = newImportTemplates(suite, impDecl.Module.Tok.String())
				}
			default:
				log.Debug(fmt.Sprintf("Kind not considered yet: %#v)", defKind))
			}
		}
	default:
		log.Debug(fmt.Sprintf("Node not considered yet: %#v)", nodet))

	}
	log.Debug(fmt.Sprintf("Completion List :%#v", list))
	return list
}

func LastNonWsToken(n ast.Node, pos loc.Pos) []ast.Node {
	var (
		completed bool       = false
		nodeStack []ast.Node = make([]ast.Node, 0, 10)
		lastStack []ast.Node = nil
	)

	ast.Inspect(n, func(n ast.Node) bool {
		if n == nil {
			// called on node exit
			nodeStack = nodeStack[:len(nodeStack)-1]
			return false
		}

		log.Debug(fmt.Sprintf("looking for %d In node[%d .. %d] (node: %#v)", pos, n.Pos(), n.End(), n))
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
	fileName := filepath.Base(params.TextDocument.URI.SpanURI().Filename())
	defaultModuleId := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	suites := s.Owners(params.TextDocument.URI)
	// NOTE: having the current file owned by more then one suite should not
	// import from modules originating from both suites. This would
	// in most ways end up with cyclic imports.
	// Thus 'completion' shall collect items only from one suite.
	// Decision: first suite
	syntax := suites[0].Parse(params.TextDocument.URI.SpanURI().Filename())
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

	return &protocol.CompletionList{IsIncomplete: false, Items: NewCompListItems(suites[0], pos, nodeStack)}, nil //notImplemented("Completion")
}
