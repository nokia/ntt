package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) setTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	setTrace(params.Value)
	return nil
}

func setTrace(s string) {
	switch s {
	case "verbose":
		log.SetGlobalLevel(log.DebugLevel)
	case "off":
		log.SetGlobalLevel(log.PrintLevel)
	default:
	}
	return
}
