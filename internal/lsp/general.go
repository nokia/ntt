package lsp

import (
	"context"
	"os"
	"path"

	"github.com/nokia/ntt/internal/jsonrpc2"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
)

func (s *Server) initialize(ctx context.Context, params *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	s.stateMu.Lock()
	state := s.state
	s.stateMu.Unlock()
	if state >= serverInitializing {
		return nil, jsonrpc2.NewErrorf(jsonrpc2.CodeInvalidRequest, "server already initialized")
	}
	s.stateMu.Lock()
	s.state = serverInitializing
	s.stateMu.Unlock()

	s.pendingFolders = params.WorkspaceFolders
	if len(s.pendingFolders) == 0 {
		s.pendingFolders = []protocol.WorkspaceFolder{{
			URI:  params.RootURI,
			Name: path.Base(params.RootURI),
		}}
	}

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CodeActionProvider:         false,
			CompletionProvider:         protocol.CompletionOptions{},
			DefinitionProvider:         true,
			TypeDefinitionProvider:     false,
			ImplementationProvider:     false,
			DocumentFormattingProvider: false,
			DocumentSymbolProvider:     false,
			WorkspaceSymbolProvider:    false,
			FoldingRangeProvider:       false,
			HoverProvider:              false,
			DocumentHighlightProvider:  false,
			DocumentLinkProvider:       protocol.DocumentLinkOptions{},
			ReferencesProvider:         false,
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				Change:    protocol.Full,
				OpenClose: true,
				Save: protocol.SaveOptions{
					IncludeText: false,
				},
			},
			Workspace: protocol.WorkspaceGn{
				WorkspaceFolders: protocol.WorkspaceFoldersGn{
					Supported:           true,
					ChangeNotifications: "workspace/didChangeWorkspaceFolders",
				},
			},
		},
	}, nil
}

func (s *Server) initialized(ctx context.Context, params *protocol.InitializedParams) error {
	s.stateMu.Lock()
	s.state = serverInitialized
	s.stateMu.Unlock()

	// Create Session and add folders, the first workspace folder is considered
	// as root folder, which might contain a manifest file (package.yml)
	s.suite = &ntt.Suite{}
	if len(s.pendingFolders) >= 1 {
		s.suite.SetRoot(s.pendingFolders[0].URI)
		for i := range s.pendingFolders[1:] {
			s.suite.AddImports(s.pendingFolders[i].URI)
		}
	}

	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state < serverInitialized {
		return jsonrpc2.NewErrorf(jsonrpc2.CodeInvalidRequest, "server not initialized")
	}
	s.state = serverShutDown
	return nil
}

func (s *Server) exit(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state != serverShutDown {
		os.Exit(1)
	}
	os.Exit(0)
	return nil
}
