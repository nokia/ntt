package lsp

import (
	"bytes"
	"context"
	"os"
	"runtime/debug"
	"text/template"

	"github.com/nokia/ntt/internal/ntt"
)

const statusTemplate = `

=== Language Server Status ===

Executable : {{ .Executable }}
Version    : {{ .Version }} {{ .Sum }}
Process ID : {{ .PID }}


=== Session ===

Root Folder: {{ .Suite.Root }}
Known Files: {{- range .Suite.Files}}
	{{.}}{{end}}

`

type Status struct {
	Executable string
	Version    string
	Sum        string
	PID        int
	Suite      struct {
		Root  string
		Files []string
	}
}

func NewStatus(suite *ntt.Suite) *Status {
	s := Status{
		PID: os.Getpid(),
	}

	if path, err := os.Executable(); err == nil {
		s.Executable = path
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		s.Version = info.Main.Version
		s.Sum = info.Main.Sum
	}

	if root := suite.Root(); root != nil {
		s.Suite.Root = root.Path()
	}

	s.Suite.Files, _ = suite.Files()
	s.Suite.Files = append(s.Suite.Files, ntt.FindAuxiliaryTTCN3Files()...)

	return &s
}

func (s *Server) status(ctx context.Context) (interface{}, error) {

	var buf bytes.Buffer

	t := template.Must(template.New("ntt.status").Parse(statusTemplate))
	if err := t.Execute(&buf, NewStatus(s.suite)); err != nil {
		panic(err.Error())
	}
	s.Log(ctx, buf.String())
	return nil, nil
}
