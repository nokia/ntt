// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lsp implements LSP for gopls.
package lsp

import (
	"context"
	"fmt"
	"sync"

	"github.com/nokia/ntt/internal/lsp/jsonrpc2"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
)

func NewServer(stream jsonrpc2.Stream) *Server {
	return &Server{
		conn: jsonrpc2.NewConn(stream),
	}
}

func (s *Server) Serve(ctx context.Context) error {
	s.client = protocol.ClientDispatcher(s.conn)
	ctx = protocol.WithClient(ctx, s.client)
	handler := protocol.ServerHandler(s, jsonrpc2.MethodNotFound)
	s.conn.Go(ctx, protocol.Handlers(handler))
	<-s.conn.Done()
	return s.conn.Err()
}

type serverState int

const (
	serverCreated      = serverState(iota)
	serverInitializing // set once the server has received "initialize" request
	serverInitialized  // set once the server has received "initialized" request
	serverShutDown
)

func (s serverState) String() string {
	switch s {
	case serverCreated:
		return "created"
	case serverInitializing:
		return "initializing"
	case serverInitialized:
		return "initialized"
	case serverShutDown:
		return "shutDown"
	}
	return fmt.Sprintf("(unknown state: %d)", int(s))
}

// Server implements the protocol.Server interface.
type Server struct {
	conn   jsonrpc2.Conn
	client protocol.Client

	stateMu sync.Mutex
	state   serverState

	// folders is only valid between initialize and initialized, and holds the
	// set of folders to build views for when we are ready
	pendingFolders []protocol.WorkspaceFolder

	suite *ntt.Suite

	diagsMu sync.Mutex
	diags   map[string][]protocol.Diagnostic
}

func (s *Server) Fatal(ctx context.Context, msg string) {
	s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    protocol.Error,
		Message: msg,
	})
}

func (s *Server) Info(ctx context.Context, msg string) {
	s.client.ShowMessage(ctx, &protocol.ShowMessageParams{
		Type:    protocol.Info,
		Message: msg,
	})
}

func (s *Server) Log(ctx context.Context, msg string) {
	s.client.LogMessage(ctx, &protocol.LogMessageParams{
		Type:    protocol.Log,
		Message: msg,
	})
}

func (s *Server) cancelRequest(ctx context.Context, params *protocol.CancelParams) error {
	return nil
}

func (s *Server) codeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	return nil, nil
}

func (s *Server) nonstandardRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	return nil, notImplemented(method)
}

func notImplemented(method string) error {
	return fmt.Errorf("%w: %q not yet implemented", jsonrpc2.ErrMethodNotFound, method)
}

//go:generate helper/helper -d protocol/tsserver.go -o server_gen.go -u .
