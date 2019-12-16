package main

import (
	"context"
	"os"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/jsonrpc2"
	"github.com/spf13/cobra"
)

var (
	languageServerCmd = &cobra.Command{
		Hidden: true,
		Use:    "langserver",
		Short:  "Start TTCN-3 language server",
		Run:    langserver,
	}
)

func init() {
	rootCmd.AddCommand(languageServerCmd)
}

func langserver(cmd *cobra.Command, args []string) {
	ctx, srv := lsp.NewServer(context.Background(), jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout))
	srv.Run(ctx)
}
