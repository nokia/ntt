package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ttcn3c",
		Short: "ttcn3c parses TTCN-3 files and generates output based on the options given",
		RunE:  run,
	}

	format string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&format, "generator", "G", "stdout", "generator to use (default stdout)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ttcn3c: ", err.Error())
	os.Exit(1)
}

func run(cmd *cobra.Command, args []string) error {

	conf, err := project.Open(args...)
	if err != nil {
		fatal(err)
	}

	srcs, err := fs.TTCN3Files(conf.Sources...)
	if err != nil {
		fatal(err)
	}

	imports, err := fs.TTCN3Files(conf.Imports...)
	if err != nil {
		fatal(err)
	}

	files := append(srcs, imports...)

	name := fmt.Sprintf("ttcn3c-gen-%s", format)
	generator, err := exec.LookPath(name)
	if err != nil {
		fatal(fmt.Errorf("could not find generator %q", name))
	}

	proc := exec.Command(generator)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	stdin, err := proc.StdinPipe()
	if err != nil {
		fatal(err)
	}

	go func() {
		defer stdin.Close()
		for _, file := range files {
			src := buildSource(file)
			b, err := json.Marshal(src)
			if err != nil {
				fatal(err)
			}
			stdin.Write(b)
		}
	}()

	if err := proc.Run(); err != nil {
		fatal(err)
	}
	return nil
}

func buildSource(file string) ttcn3.Source {
	src := ttcn3.Source{
		Filename: file,
	}
	var visit func(n ast.Node)
	visit = func(n ast.Node) {
		if n == nil {
			return
		}

		k := strings.TrimPrefix(strings.TrimPrefix(fmt.Sprintf("%T", n), "*"), "ast.")
		begin := int(n.Pos())
		end := int(n.End())

		switch n := n.(type) {
		case *ast.Token:
			visit(*n)
		case ast.Token:
			if !n.IsValid() {
				break
			}
			src.Events = append(src.Events, ttcn3.NodeEvent{
				Kind: "AddToken",
				Text: n.String(),
				Offs: begin,
				Len:  end - begin,
			})
		default:
			src.Events = append(src.Events, ttcn3.NodeEvent{
				Kind: "Open" + k,
				Offs: begin,
				Len:  end - begin,
			})
			idx := len(src.Events) - 1
			for _, c := range ast.Children(n) {
				visit(c)
			}
			src.Events = append(src.Events, ttcn3.NodeEvent{
				Kind:  "Close" + k,
				Offs:  begin,
				Len:   end - begin,
				Other: idx,
			})
			src.Events[idx].Other = len(src.Events) - 1
		}
	}
	visit(ttcn3.ParseFile(file).Root)
	return src
}
