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
	err      error
	e        *t3xf.Encoder
	lastLine int
	fields   map[string]int
}

func NewCompiler() *Compiler {
	c := &Compiler{
		e:      t3xf.NewEncoder(),
		fields: make(map[string]int),
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
	if line := syntax.Begin(n).Line; line != c.lastLine {
		c.emit(opcode.LINE, line)
		c.lastLine = line
	}

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
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.MODULE, 0)

	case *syntax.FuncDecl:
		switch k := n.Kind.Kind(); {
		case k == syntax.FUNCTION && n.External == nil:
			c.compileFunction(n)
		case k == syntax.FUNCTION && n.External != nil:
			c.compileExtFunc(n)
		case k == syntax.TESTCASE:
			c.compileTestcase(n)
		case k == syntax.ALTSTEP:
			c.compileAltstep(n)
		default:
			c.errorf("unsupported behaviour %s", k)
		}

	case *syntax.ValueDecl:
		k := syntax.VAR
		if n.Kind != nil {
			k = n.Kind.Kind()
		}

		fn := c.compileVar
		switch k {
		case syntax.CONST:
			fn = c.compileConst
		case syntax.MODULEPAR:
			fn = c.compileModulePar
		}

		for _, decl := range n.Decls {
			fn(k, n.TemplateRestriction, n.Modif, n.Type, decl, n.With)
		}

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

	case *syntax.ReturnStmt:
		if n.Result != nil {
			c.Compile(n.Result)
		}
		c.emit(opcode.RETURN, 0)

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
		c.emit(opcode.SCAN, 0)
		c.Compile(n.Cond)
		c.emit(opcode.BLOCK, 0)
		c.Compile(n.Body)
		c.emit(opcode.WHILE, 0)

	case *syntax.DoWhileStmt:
		c.Compile(n.Body)
		c.emit(opcode.SCAN, 0)
		c.Compile(n.Cond)
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.DOWHILE, 0)

	case *syntax.ForStmt:
		c.emit(opcode.SCAN, 0)
		c.Compile(n.Init)
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.SCAN, 0)
		c.Compile(n.Cond)
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.SCAN, 0)
		c.Compile(n.Post)
		c.emit(opcode.BLOCK, 0)
		c.Compile(n.Body)
		c.emit(opcode.FOR, 0)

	case *syntax.ExprStmt:
		c.Compile(n.Expr)

	case *syntax.DeclStmt:
		c.Compile(n.Decl)

	case *syntax.CallExpr:
		for _, arg := range n.Args.List {
			c.Compile(arg)
		}
		c.Compile(n.Fun)

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
		case syntax.RANGE:
			c.emit(opcode.RANGE, 0)
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

	case *syntax.RefSpec:
		c.Compile(n.X)

	case *syntax.Ident:
		s := n.String()
		switch s {
		case "log":
			c.emit(opcode.LOG, 0)
		case "integer":
			c.emit(opcode.INTEGER, 0)
		case "float":
			c.emit(opcode.FLOAT, 0)
		case "bitstring":
			c.emit(opcode.BITSTRING, 0)
		case "hexstring":
			c.emit(opcode.HEXSTRING, 0)
		case "octetstring":
			c.emit(opcode.OCTETSTRING, 0)
		case "boolean":
			c.emit(opcode.BOOLEAN, 0)
		case "charstring":
			c.emit(opcode.CHARSTRING, 0)
		case "timer":
			c.emit(opcode.TIMER, 0)
		default:
			c.errorf("unknown identifier %s", s)
		}

	case *syntax.WithSpec:
		c.emit(opcode.SCAN, 0)
		for _, child := range n.List {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)

	case *syntax.FormalPars:
		if len(n.List) == 0 {
			c.emit(opcode.SKIP, 0)
			break
		}
		c.emit(opcode.SCAN, 0)
		for _, child := range n.List {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)

	case *syntax.FormalPar:
		c.Compile(n.Type)
		if n.TemplateRestriction != nil {
			c.Compile(n.TemplateRestriction)
		}
		c.emit(opcode.NAME, n.Name.String())
		dir := opcode.IN
		if n.Direction != nil {
			switch n.Direction.Kind() {
			case syntax.OUT:
				dir = opcode.OUT
			case syntax.INOUT:
				dir = opcode.INOUT
			}
		}
		c.emit(dir, 0)

	case *syntax.RestrictionSpec:
		op := opcode.PERMITT
		if n.Tok != nil {
			switch n.Tok.Kind() {
			case syntax.OMIT:
				op = opcode.PERMITO
			case syntax.PRESENT:
				op = opcode.PERMITP
			}
		}
		c.emit(op, 0)

	case *syntax.ReturnSpec:
		c.Compile(n.Type)
		if n.Restriction != nil {
			c.Compile(n.Restriction)
		}

	case *syntax.StructTypeDecl:
		op := opcode.TYPE
		if n.With != nil {
			op = opcode.TYPEW
			c.Compile(n.With)
		}
		c.emit(opcode.SCAN, 0)
		for _, field := range n.Fields {
			c.Compile(field)
			c.emit(opcode.NAME, field.Name.String())
			switch {
			case field.Optional != nil:
				c.emit(opcode.FIELDO, 0)
			case n.Kind.Kind() == syntax.UNION:
				c.emit(opcode.IFIELD, c.fieldIndex(field.Name.String()))
			default:
				c.emit(opcode.FIELD, 0)
			}
		}
		c.emit(opcode.BLOCK, 0)
		switch n.Kind.Kind() {
		case syntax.RECORD:
			c.emit(opcode.RECORD, 0)
		case syntax.SET:
			c.emit(opcode.SET, 0)
		case syntax.UNION:
			c.emit(opcode.UNION, 0)
		default:
			c.errorf("unsupported struct type %s", n.Kind)
		}
		c.emit(opcode.NAME, n.Name.String())
		c.emit(op, 0)

	case *syntax.SubTypeDecl:
		op := opcode.TYPE
		if n.With != nil {
			op = opcode.TYPEW
			c.Compile(n.With)
		}
		c.Compile(n.Field)
		c.emit(opcode.NAME, n.Field.Name.String())
		c.emit(op, 0)

	case *syntax.Field:
		c.Compile(n.Type)

		// NOTE: This length constraint must move to the element type. not sure how this works mit structs (inside fields)
		if n.ValueConstraint != nil || n.LengthConstraint != nil {
			c.emit(opcode.SCAN, 0)
			if n.ValueConstraint != nil {
				c.emit(opcode.MARK, 0)
				for _, x := range n.ValueConstraint.List {
					c.Compile(x)
				}
				c.emit(opcode.COLLECT, 0)
			} else {
				c.emit(opcode.ANY, 0)
			}
			if n.LengthConstraint != nil {
				for _, x := range n.LengthConstraint.Size.List {
					c.Compile(x)
				}
				c.emit(opcode.LENGTH, 0)
			}
			c.emit(opcode.BLOCK, 0)
			c.emit(opcode.SUBTYPE, 0)
		}

		for i := len(n.ArrayDef) - 1; i >= 0; i-- {
			c.emit(opcode.SCAN, 0)
			for _, x := range n.ArrayDef[i].List {
				c.Compile(x)
			}
			c.emit(opcode.BLOCK, 0)
			c.emit(opcode.ARRAY, 0)
		}

	case *syntax.ListSpec:
		c.Compile(n.ElemType)
		switch n.Kind.Kind() {
		case syntax.RECORD:
			c.emit(opcode.RECORDOF, 0)
		case syntax.SET:
			c.emit(opcode.SETOF, 0)
		default:
			c.errorf("unsupported list type %s", n.Kind)

		}
		if n.Length != nil {
			c.emit(opcode.SCAN, 0)
			c.emit(opcode.ANY, 0)
			for _, x := range n.Length.Size.List {
				c.Compile(x)
			}
			c.emit(opcode.LENGTH, 0)
			c.emit(opcode.BLOCK, 0)
			c.emit(opcode.SUBTYPE, 0)
		}

	default:
		c.errorf("unexpected node type %T", n)
	}

	return nil
}

func (c *Compiler) compileFunction(n *syntax.FuncDecl) error {
	if n.With != nil {
		c.errorf("function attributes not supported")
	}
	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	switch {
	case n.Return == nil && n.RunsOn == nil:
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTION, 0)

	case n.Return == nil && n.RunsOn != nil:
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONB, 0)

	case n.Return != nil && n.RunsOn == nil:
		c.Compile(n.Return)
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONV, 0)

	case n.Return != nil && n.RunsOn != nil:
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.Return)
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONVB, 0)
	}

	return nil
}

func (c *Compiler) compileExtFunc(n *syntax.FuncDecl) error {
	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	if n.RunsOn != nil {
		c.errorf("runs on clause not supported")
	}

	switch {
	case n.Return == nil && n.With == nil:
		c.Compile(n.Params)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONX, 0)

	case n.Return == nil && n.With != nil:
		c.Compile(n.With)
		c.Compile(n.Params)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONXW, 0)

	case n.Return != nil && n.With == nil:
		c.Compile(n.Return)
		c.Compile(n.Params)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONXV, 0)

	case n.Return != nil && n.With != nil:
		c.Compile(n.Return)
		c.Compile(n.With)
		c.Compile(n.Params)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.FUNCTIONXVW, 0)
	}
	return nil
}

func (c *Compiler) compileTestcase(n *syntax.FuncDecl) error {
	op := opcode.TESTCASE
	if n.System != nil {
		op = opcode.TESTCASES
		c.Compile(n.System.Comp)
	}
	c.Compile(n.RunsOn.Comp)
	c.Compile(n.Params)
	c.Compile(n.Body)
	c.emit(opcode.NAME, n.Name.String())
	c.emit(op, 0)
	return nil
}

func (c *Compiler) compileAltstep(n *syntax.FuncDecl) error {
	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	switch {
	case n.RunsOn == nil && n.With == nil:
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.ALTSTEP, 0)

	case n.RunsOn == nil && n.With != nil:
		c.Compile(n.With)
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.ALTSTEPW, 0)

	case n.RunsOn != nil && n.With == nil:
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.Params)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.ALTSTEPB, 0)

	case n.RunsOn != nil && n.With != nil:
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.With)
		c.Compile(n.Params)
		c.Compile(n.Body)
		c.emit(opcode.NAME, n.Name.String())
		c.emit(opcode.ALTSTEPBW, 0)
	}
	return nil
}

func (c *Compiler) compileVar(kind syntax.Kind, restr *syntax.RestrictionSpec, modif syntax.Token, typ syntax.Expr, decl *syntax.Declarator, attrs *syntax.WithSpec) error {
	if attrs != nil {
		c.errorf("attributes not supported")
	}
	c.Compile(typ)
	for i := len(decl.ArrayDef) - 1; i >= 0; i-- {
		c.emit(opcode.SCAN, 0)
		for _, x := range decl.ArrayDef[i].List {
			c.Compile(x)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.ARRAY, 0)
	}
	c.emit(opcode.NAME, decl.Name.String())
	if modif != nil {
		c.errorf("modifiers not supported")
	}
	addr := c.emit(opcode.VAR, 0)
	if decl.Value != nil {
		c.Compile(decl.Value)
		c.emit(opcode.REF, t3xf.Reference(addr))
		c.emit(opcode.ASSIGN, 0)
	}
	return nil
}

func (c *Compiler) compileConst(kind syntax.Kind, restr *syntax.RestrictionSpec,
	modif syntax.Token, typ syntax.Expr, decl *syntax.Declarator, attrs *syntax.WithSpec) error {
	op := opcode.CONST
	if attrs != nil {
		op = opcode.CONSTW
		c.Compile(attrs)
	}

	if decl.Value != nil {
		c.emit(opcode.SCAN, 0)
		c.Compile(decl.Value)
		c.emit(opcode.BLOCK, 0)
	} else {
		c.errorf("constant declaration without value")
	}

	c.Compile(typ)
	for i := len(decl.ArrayDef) - 1; i >= 0; i-- {
		c.emit(opcode.SCAN, 0)
		for _, x := range decl.ArrayDef[i].List {
			c.Compile(x)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.ARRAY, 0)
	}
	c.emit(opcode.NAME, decl.Name.String())
	if modif != nil {
		c.errorf("modifiers not supported")
	}
	c.emit(op, 0)
	return nil
}

func (c *Compiler) compileModulePar(kind syntax.Kind, restr *syntax.RestrictionSpec,
	modif syntax.Token, typ syntax.Expr, decl *syntax.Declarator, attrs *syntax.WithSpec) error {

	op := opcode.MPAR
	if decl.Value != nil {
		op = opcode.MPARD
		c.emit(opcode.SCAN, 0)
		c.Compile(decl.Value)
		c.emit(opcode.BLOCK, 0)
	}

	if attrs != nil {
		c.errorf("attributes not supported")
	}
	if modif != nil {
		c.errorf("modifiers not supported")
	}

	c.Compile(typ)
	for i := len(decl.ArrayDef) - 1; i >= 0; i-- {
		c.emit(opcode.SCAN, 0)
		for _, x := range decl.ArrayDef[i].List {
			c.Compile(x)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.ARRAY, 0)
	}

	c.emit(opcode.NAME, decl.Name.String())
	c.emit(op, 0)
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

func (c *Compiler) fieldIndex(name string) int {
	i, ok := c.fields[name]
	if !ok {
		i = len(c.fields)
		c.fields[name] = i
	}
	return i
}
