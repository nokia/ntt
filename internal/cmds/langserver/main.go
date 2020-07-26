package langserver

import (
	"context"
	"os"

	"github.com/nokia/ntt/internal/lsp"
	"github.com/nokia/ntt/internal/lsp/jsonrpc2"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Hidden: true,
		Use:    "langserver",
		Short:  "Start TTCN-3 language server",
		RunE:   langserver,
	}
)

func langserver(cmd *cobra.Command, args []string) error {
	ctx, srv := lsp.NewServer(context.Background(), jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout), false)
	return srv.Run(ctx)
}
