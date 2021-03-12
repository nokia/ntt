package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

var moduleDefKw = []string{"import from", "type", "const", "modulepar", "template", "function", "external function", "altstep", "testcase", "control", "signature"}

func newModuleDefKw() []protocol.CompletionItem {
	complList := make([]protocol.CompletionItem, len(moduleDefKw))
	for _, v := range moduleDefKw {
		complList = append(complList, protocol.CompletionItem{Label: v, Kind: protocol.KeywordCompletion})
	}
	return complList
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
	id, _ := s.suite.IdentifierAt(syntax.Module.Name.String(), int(params.TextDocumentPositionParams.Position.Line), int(params.TextDocumentPositionParams.Position.Character))
	log.Debug(fmt.Sprintf("Completion at id :%#v", id))
	return &protocol.CompletionList{IsIncomplete: false, Items: newModuleDefKw()}, nil //notImplemented("Completion")
}
