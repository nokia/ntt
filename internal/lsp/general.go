package lsp

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/jsonrpc2"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/project"
	errors "golang.org/x/xerrors"
)

func registerSemanticTokens() *protocol.SemanticTokensRegistrationOptions {
	semTok := os.Getenv("NTT_SEMANTIC_TOKENS")
	if strings.ToLower(semTok) == "on" {
		return &protocol.SemanticTokensRegistrationOptions{

			TextDocumentRegistrationOptions: protocol.TextDocumentRegistrationOptions{
				DocumentSelector: protocol.DocumentSelector{
					protocol.DocumentFilter{Language: "ttcn3", Scheme: "file", Pattern: "**/*.ttcn3"},
				},
			},
			SemanticTokensOptions: protocol.SemanticTokensOptions{
				Legend: protocol.SemanticTokensLegend{
					TokenTypes:     TokenTypes,
					TokenModifiers: TokenModifiers,
				},
				Range: true,
				Full:  true,
			}}
	}
	return nil
}
func (s *Server) initialize(ctx context.Context, params *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	s.stateMu.Lock()
	state := s.state
	s.stateMu.Unlock()
	if state >= serverInitializing {
		return nil, errors.Errorf("%w: initialize called while server in %v state", jsonrpc2.ErrInvalidRequest, s.state)
	}
	s.stateMu.Lock()
	s.state = serverInitializing
	s.stateMu.Unlock()

	s.pendingFolders = params.WorkspaceFolders
	if len(s.pendingFolders) == 0 && params.RootURI != "" {
		s.pendingFolders = []protocol.WorkspaceFolder{{
			URI:  string(params.RootURI),
			Name: path.Base(string(params.RootURI.SpanURI())),
		}}
	}

	setTrace(params.Trace)

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CodeActionProvider:         false,
			CompletionProvider:         protocol.CompletionOptions{TriggerCharacters: []string{"."}},
			DefinitionProvider:         true,
			TypeDefinitionProvider:     false,
			ImplementationProvider:     false,
			DocumentFormattingProvider: false,
			DocumentSymbolProvider:     true,
			WorkspaceSymbolProvider:    false,
			FoldingRangeProvider:       false,
			HoverProvider:              true,
			DocumentHighlightProvider:  false,
			DocumentLinkProvider:       protocol.DocumentLinkOptions{},
			ReferencesProvider:         true,
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				Change:    protocol.Full,
				OpenClose: true,
				Save: protocol.SaveOptions{
					IncludeText: false,
				},
			},
			SemanticTokensProvider: registerSemanticTokens(),
			Workspace: protocol.Workspace5Gn{
				WorkspaceFolders: protocol.WorkspaceFolders4Gn{
					Supported:           true,
					ChangeNotifications: "workspace/didChangeWorkspaceFolders",
				},
			},
		},
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version,omitempty"`
		}{Name: "ntt", Version: "0.12.0"},
	}, nil
}

func (s *Server) initialized(ctx context.Context, params *protocol.InitializedParams) error {
	s.stateMu.Lock()
	s.state = serverInitialized
	s.stateMu.Unlock()

	for _, folder := range s.pendingFolders {
		log.Printf("Scanning %q for possible TTCN-3 suites\n", folder.URI)
		for _, root := range project.Discover(folder.URI) {
			s.AddSuite(root)
		}
	}

	s.testCtrl = &TestController{}
	s.testCtrl.Start(s.client, s, &s.Suites)
	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state < serverInitialized {
		log.Verbose("language server shutdown without initialization")
	}
	s.state = serverShutDown
	s.testCtrl.Shutdown()
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
