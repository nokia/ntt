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

var regFormatCapability = protocol.RegistrationParams{
	Registrations: []protocol.Registration{
		{ID: "FormCap",
			Method:          "textDocument/formatting",
			RegisterOptions: protocol.TextDocumentRegistrationOptions{DocumentSelector: []protocol.DocumentFilter{{Language: "ttcn3"}}},
		}}}
var unregFormatCapability = protocol.UnregistrationParams{
	Unregisterations: []protocol.Unregistration{{ID: "FormCap", Method: "textDocument/formatting"}},
}

func (s *Server) didChangeConfiguration(ctx context.Context, _ *protocol.DidChangeConfigurationParams) error {
	if format, ok := s.Config("ttcn3.server.useOwnFormatter").(bool); ok {

		if format == s.serverConfig.UseOwnFormatter {
			return nil
		}
		if s.serverConfig.UseOwnFormatter {
			log.Debug("didChangeConfiguration: unregister formatter capability\n")
			s.client.UnregisterCapability(ctx, &unregFormatCapability)
		} else {
			log.Debug("didChangeConfiguration: register formatter capability\n")
			s.client.RegisterCapability(ctx, &regFormatCapability)
		}
		s.serverConfig.UseOwnFormatter = format
	}
	return nil
}
