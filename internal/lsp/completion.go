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
)

var moduleDefKw = []string{"import from", "type", "const", "modulepar", "template", "function", "external function", "altstep", "testcase", "control", "signature"}

func newModuleDefKw() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, 0, len(moduleDefKw))
	for _, v := range moduleDefKw {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return complList
}

func newCompListItems(suite *ntt.Suite, pos loc.Pos, nodes []ast.Node) []protocol.CompletionItem {
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
			if _, ok := nodes[l-2].(*ast.ImportDecl); ok {
				// look for available modules for import
				if files, err := suite.Files(); err == nil {
					list = make([]protocol.CompletionItem, 0, len(files))
					for _, f := range files {
						fileName := filepath.Base(f)
						fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
						list = append(list, protocol.CompletionItem{Label: fileName, Kind: protocol.ModuleCompletion})
					}
				}
			}
		}
	case *ast.ImportDecl:
		if nodet.Module == nil {
			// look for available modules for import
			if files, err := suite.Files(); err == nil {
				list = make([]protocol.CompletionItem, 0, len(files))
				for _, f := range files {
					fileName := filepath.Base(f)
					fileName = fileName[:len(fileName)-len(filepath.Ext(fileName))]
					list = append(list, protocol.CompletionItem{Label: fileName, Kind: protocol.ModuleCompletion})
				}
			}
		}
	case *ast.ErrorNode:
		// i.e. user started typing => ast.Ident might be detected instead of a kw
		if _, ok := nodes[l-2].(*ast.ModuleDef); l > 1 && ok {
			// start a new module def
			list = newModuleDefKw()
		}
	default:
		log.Debug(fmt.Sprintf("Node not considered yet: %#v)", nodet))

	}
	return list
}

func lastNonWsToken(n ast.Node, pos loc.Pos) []ast.Node {
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
		log.Debug(fmt.Sprintf("Not a Token :%#v", n))
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

	syntax := s.suite.Parse(params.TextDocument.URI.SpanURI().Filename())
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
	nodeStack := lastNonWsToken(syntax.Module, pos)

	return &protocol.CompletionList{IsIncomplete: false, Items: newCompListItems(s.suite, pos, nodeStack)}, nil //notImplemented("Completion")
}
