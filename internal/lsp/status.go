package lsp

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/nokia/ntt/internal/ntt"
)

const statusTemplate = `

=== Language Server Status ===

Executable : {{ .Executable }}
Version    : {{ .Version }}
Process ID : {{ .PID }}


=== Session ===

{{range .Suites}}
Root Folder: {{ .Root }}
Known Files: {{ range .Files}}
	- {{.}}{{end}}

{{end}}
`

type Status struct {
	Executable string
	Version    string
	PID        int
	Suites     []*ntt.Suite
}

func NewStatus(suites []*ntt.Suite) *Status {
	s := Status{
		PID:    os.Getpid(),
		Suites: suites,
	}

	if path, err := os.Executable(); err == nil {
		s.Executable = path
		if out, err := exec.Command(path, "version").Output(); err == nil {
			s.Version = strings.TrimSpace(string(out))
		}
	}

	return &s
}

func (s *Server) status(ctx context.Context) (interface{}, error) {
	t := template.Must(template.New("ntt.status").Parse(statusTemplate))
	var suites []*ntt.Suite
	for _, s := range s.roots {
		suites = append(suites, s)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, NewStatus(suites)); err != nil {
		panic(err.Error())
	}
	s.Log(ctx, buf.String())
	return nil, nil
}
