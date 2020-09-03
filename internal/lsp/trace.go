package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) setTraceNotification(ctx context.Context, params *protocol.SetTraceParams) error {
	switch params.Value {
	case "verbose":
		log.SetGlobalLevel(log.DebugLevel)
	case "off":
		log.SetGlobalLevel(log.PrintLevel)
	default:
	}
	return nil
}
