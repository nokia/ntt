package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
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
	if format == "t3xf" {
		return writeT3xf(Project)
	}

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
		b, err := json.MarshalIndent(src, "", "  ")
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
	var visit func(n syntax.Node)
	visit = func(n syntax.Node) {
		if n == nil {
			return
		}

		k := strings.TrimPrefix(strings.TrimPrefix(fmt.Sprintf("%T", n), "*"), "syntax.")
		begin := int(n.Pos())
		end := int(n.End())

		switch n := n.(type) {
		case syntax.Token:
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
			for _, c := range n.Children() {
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

func writeT3xf(conf *project.Config) error {
	c := NewCompiler()

	srcs, err := fs.TTCN3Files(Project.Sources...)
	if err != nil {
		return err
	}

	for _, src := range srcs {
		root := ttcn3.ParseFile(src).Root
		if root.Err() != nil {
			return root.Err()
		}
		c.Compile(root)
	}

	// TODO: Add support for imports
	b, err := c.Assemble()
	if err != nil {
		return err
	}

	return os.WriteFile("ntt.t3xf", b, 0644)
}

type Compiler struct {
	err error
	e   *t3xf.Encoder
}

func NewCompiler() *Compiler {
	c := &Compiler{
		e: t3xf.NewEncoder(),
	}
	c.emit(opcode.NOP, 0)
	c.emit(opcode.NATLONG, 2)
	c.emit(opcode.VERSION, 0)
	return c
}

func (c *Compiler) Err() error {
	return c.err
}

func (c *Compiler) Assemble() ([]byte, error) {
	if c.err == nil {
		return c.e.Assemble()
	}
	return nil, c.err
}

func (c *Compiler) Compile(n syntax.Node) error {
	switch n := n.(type) {
	case *syntax.Root:
		c.emit(opcode.SCAN, 0)
		for _, child := range n.Nodes {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.NAME, n.Filename)
		c.emit(opcode.SOURCE, 0)

	case *syntax.Module:
		if attrs := n.With; attrs != nil {
			c.errorf("module attributes not supported")
		}
		c.emit(opcode.SCAN, 0)
		for _, child := range n.Defs {
			c.Compile(child.Def)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.NAME, syntax.Name(n))
		c.emit(opcode.MODULE, 0)

	case *syntax.ControlPart:
		if attrs := n.With; attrs != nil {
			c.errorf("control part attributes not supported")
		}
		c.Compile(n.Body)
		c.emit(opcode.CONTROL, 0)

	case *syntax.BlockStmt:
		if len(n.Stmts) == 0 {
			c.emit(opcode.SKIP, 0)
			break
		}
		c.emit(opcode.SCAN, 0)
		for _, child := range n.Stmts {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)

	case *syntax.IfStmt:
		op := opcode.IF
		c.Compile(n.Cond)
		c.Compile(n.Then)
		if n.Else != nil {
			c.Compile(n.Else)
			op = opcode.IFELSE
		}
		c.emit(op, 0)

	case *syntax.WhileStmt:

	case *syntax.ExprStmt:
		c.Compile(n.Expr)

	case *syntax.BinaryExpr:
		c.Compile(n.X)
		c.Compile(n.Y)
		switch n.Op.Kind() {
		case syntax.ADD:
			c.emit(opcode.ADD, 0)
		case syntax.SUB:
			c.emit(opcode.SUB, 0)
		case syntax.MUL:
			c.emit(opcode.MUL, 0)
		case syntax.DIV:
			c.emit(opcode.DIV, 0)
		case syntax.MOD:
			c.emit(opcode.MOD, 0)
		case syntax.REM:
			c.emit(opcode.REM, 0)
		case syntax.EQ:
			c.emit(opcode.EQ, 0)
		case syntax.NE:
			c.emit(opcode.NE, 0)
		default:
			c.errorf("unsupported binary operator %s", n.Op)
		}

	case *syntax.ValueLiteral:
		switch n.Tok.Kind() {
		case syntax.INT:
			s := n.Tok.String()
			bi, ok := big.NewInt(0).SetString(s, 10)
			if !ok {
				c.errorf("invalid integer %s", s)
			}
			if i := bi.Int64(); bi.IsInt64() && math.MinInt32 <= i && i <= math.MaxInt32 {
				c.emit(opcode.NATLONG, int(i))
			} else {
				c.emit(opcode.ISTR, s)
			}

		case syntax.STRING:
			s, err := syntax.Unquote(n.Tok.String())
			if err != nil {
				c.errorf("invalid string %s", n.Tok)
			}
			c.emit(opcode.UTF8, s)

		case syntax.TRUE:
			c.emit(opcode.TRUE, 0)
		case syntax.FALSE:
			c.emit(opcode.FALSE, 0)

		case syntax.NULL:
			c.emit(opcode.NULL, 0)
		case syntax.ANY:
			c.emit(opcode.ANY, 0)
		case syntax.MUL:
			c.emit(opcode.ANYN, 0)
		case syntax.OMIT:
			c.emit(opcode.OMIT, 0)

		case syntax.ERROR:
			c.emit(opcode.ERROR, 0)
		case syntax.FAIL:
			c.emit(opcode.FAIL, 0)
		case syntax.INCONC:
			c.emit(opcode.INCONC, 0)
		case syntax.PASS:
			c.emit(opcode.PASS, 0)
		case syntax.NONE:
			c.emit(opcode.NONE, 0)
		}

	case *syntax.WithSpec:
		c.emit(opcode.SCAN, 0)
		for _, child := range n.List {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)

	default:
		c.errorf("unexpected node type %T", n)
	}

	return nil
}

func (c *Compiler) emit(op opcode.Opcode, arg any) int {
	pos := c.e.Len()
	if err := c.e.Encode(op, arg); err != nil {
		c.errorf("%w", err)
	}
	return pos

}

func (c *Compiler) errorf(format string, args ...interface{}) {
	c.err = errors.Join(c.err, fmt.Errorf(format, args...))
}
