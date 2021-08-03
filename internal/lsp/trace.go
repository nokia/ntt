package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) setTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	setTrace(params.Value)
	return nil
}

func setTrace(s string) {
	// Ignore client settings when NTT_DEBUG is enabled
	if env := env.Getenv("NTT_DEBUG"); env != "" {
		log.SetGlobalLevel(log.DebugLevel)
		return
	}

	switch s {
	case "verbose":
		log.SetGlobalLevel(log.DebugLevel)
	case "off":
		log.SetGlobalLevel(log.PrintLevel)
	}
	return
}
