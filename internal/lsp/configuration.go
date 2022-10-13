package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) Config(section string) interface{} {
	v, err := s.client.Configuration(context.TODO(), &protocol.ParamConfiguration{
		ConfigurationParams: protocol.ConfigurationParams{
			Items: []protocol.ConfigurationItem{
				{Section: section},
			},
		},
	})
	if err != nil {
		s.Log(context.TODO(), err.Error())
	}
	if len(v) == 1 {
		return v[0]
	}
	return v
}

func (s *Server) didChangeConfiguration(ctx context.Context, _ *protocol.DidChangeConfigurationParams) error {
	if format, ok := s.Config("ttcn3.format.enabled").(bool); ok {

		if format == s.serverConfig.FormatEnabled {
			return nil
		}
		if s.serverConfig.FormatEnabled {
			log.Debug("didChangeConfiguration: unregister formatter capability\n")
			s.client.UnregisterCapability(ctx, &protocol.UnregistrationParams{
				Unregisterations: []protocol.Unregistration{{ID: "ttcn3.formatting", Method: "textDocument/formatting"}},
			})
		} else {
			log.Debug("didChangeConfiguration: register formatter capability\n")
			s.client.RegisterCapability(ctx, &protocol.RegistrationParams{
				Registrations: []protocol.Registration{
					{ID: "ttcn3.formatting",
						Method:          "textDocument/formatting",
						RegisterOptions: protocol.TextDocumentRegistrationOptions{DocumentSelector: []protocol.DocumentFilter{{Language: "ttcn3"}}},
					}}})
		}
		s.serverConfig.FormatEnabled = format
	}
	return nil
}
