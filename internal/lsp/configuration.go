package lsp

import (
	"context"

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
		s.serverConfig.FormatEnabled = format
	}
	return nil
}
