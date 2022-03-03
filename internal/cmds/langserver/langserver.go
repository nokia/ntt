package langserver

import (
	"context"
	"os"

	"github.com/nokia/ntt/internal/fakenet"
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
	stream := jsonrpc2.NewHeaderStream(fakenet.NewConn("stdio", os.Stdin, os.Stdout))
	return lsp.NewServer(stream).Serve(context.TODO())
}
