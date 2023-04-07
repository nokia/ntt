package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	CompileCommand = &cobra.Command{
		Use:   "compile",
		Short: "Compile TTCN-3 sources and generate output for other tools",
		Long:  `Compile TTCN-3 sources and generate output for other tools.`,

		RunE: compile,
	}

	format string
)

func init() {
	CompileCommand.Flags().StringVarP(&format, "generator", "G", "stdout", "generator to use (default stdout)")
}

func compile(cmd *cobra.Command, args []string) error {
	srcs, err := fs.TTCN3Files(Project.Sources...)
	if err != nil {
		return err
	}

	imports, err := fs.TTCN3Files(Project.Imports...)
	if err != nil {
		return err
	}

	files := append(srcs, imports...)

	if format == "stdout" {
		writeSource(os.Stdout, files...)
		return nil
	}

	generator, err := exec.LookPath(fmt.Sprintf("ntt-gen-%s", format))
	if err != nil {
		return fmt.Errorf("could not find generator %q", format)
	}
	proc := exec.Command(generator)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	stdin, err := proc.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		writeSource(stdin, files...)
	}()

	if err := proc.Run(); err != nil {
		return err
	}
	return nil
}

func writeSource(w io.Writer, files ...string) {
	for _, file := range files {
		src := buildSource(file)
		b, err := json.Marshal(src)
		if err != nil {
			fatal(err)
		}
		w.Write(b)
	}
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
		case ast.Token:
			if n == nil {
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
