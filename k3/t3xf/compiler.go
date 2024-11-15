package t3xf

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/nokia/ntt/ttcn3/syntax"
)

var universe = &scope{
	store: map[string]symbol{
		"log":         {op: opcode.LOG},
		"integer":     {op: opcode.INTEGER},
		"float":       {op: opcode.FLOAT},
		"bitstring":   {op: opcode.BITSTRING},
		"hexstring":   {op: opcode.HEXSTRING},
		"octetstring": {op: opcode.OCTETSTRING},
		"boolean":     {op: opcode.BOOLEAN},
		"charstring":  {op: opcode.CHARSTRING},
		"timer":       {op: opcode.TIMER},
	}}

type scope struct {
	parent *scope
	store  map[string]symbol
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, store: make(map[string]symbol)}
}

func (s *scope) lookup(name string) (symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.parent != nil {
		return s.parent.lookup(name)
	}
	return obj, ok
}

func (s *scope) define(name string, sym symbol) {
	s.store[name] = sym
}

type symbol struct {
	n   syntax.Node
	op  opcode.Opcode
	arg any
}

func (s symbol) IsCompiled() bool {
	return !(s.op == 0 && s.arg == nil)
}

type Compiler struct {
	scope    *scope
	err      error
	e        *Encoder
	lastLine int
	fields   map[string]int
}

func NewCompiler() *Compiler {
	return &Compiler{
		e:      NewEncoder(),
		fields: make(map[string]int),
		scope:  newScope(universe),
	}
}

func (c *Compiler) Err() error {
	return c.err
}

func (c *Compiler) Assemble() ([]byte, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.e.Assemble()
}

func (c *Compiler) Compile(n syntax.Node) error {
	if line := syntax.Begin(n).Line; line != c.lastLine {
		c.emit(opcode.LINE, line)
		c.lastLine = line
	}

	switch n := n.(type) {
	case *syntax.Root:
		if n.Filename != "" {
			c.emit(opcode.SCAN, 0)
		}
		for _, child := range n.Nodes {
			c.Compile(child)
		}
		if n.Filename != "" {
			c.emit(opcode.BLOCK, 0)
			c.emit(opcode.NAME, n.Filename)
			c.emit(opcode.SOURCE, 0)
		}

	case *syntax.Module:
		c.scope = newScope(c.scope)
		n.Inspect(func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.GroupDecl:
				return true
			case *syntax.ModuleDef:
				return true
			case *syntax.ValueDecl:
				return true
			default:
				if name := syntax.Name(n); name != "" {
					c.scope.define(name, symbol{n: n})
				}
				return false
			}
		})

		op := opcode.MODULE
		if attrs := n.With; attrs != nil {
			op = opcode.MODULEW
			c.Compile(attrs)
		}
		c.emit(opcode.SCAN, 0)
		for _, child := range n.Defs {
			c.Compile(child.Def)
		}
		c.emit(opcode.BLOCK, 0)

		c.emit(opcode.NAME, n.Name.String())
		addr := c.emit(op, 0)
		c.scope = c.scope.parent
		c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})

	case *syntax.ModuleDef:
		c.Compile(n.Def)

	case *syntax.FuncDecl:
		switch k := n.KindTok.Kind(); {
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
		if n.KindTok != nil {
			k = n.KindTok.Kind()
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
		op := opcode.CONTROL
		if attrs := n.With; attrs != nil {
			op = opcode.CONTROLW
			c.Compile(attrs)
		}
		c.Compile(n.Body)

		addr := c.emit(op, 0)
		c.scope.define("control", symbol{n: n, op: opcode.REF, arg: Reference(addr)})

	case *syntax.BlockStmt:
		if len(n.Stmts) == 0 {
			c.emit(opcode.SKIP, 0)
			break
		}

		c.scope = newScope(c.scope)
		c.emit(opcode.SCAN, 0)
		for _, child := range n.Stmts {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)
		c.scope = c.scope.parent

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
		c.scope = newScope(c.scope)
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
		c.scope = c.scope.parent

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
		if n.Op.Kind() == syntax.ASSIGN {
			c.Compile(n.Y)
			c.Compile(n.X)
			c.emit(opcode.ASSIGN, 0)
			break
		}

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

	case *syntax.CompositeLiteral:
		c.emit(opcode.MARK, 0)
		for _, elem := range n.List {
			c.Compile(elem)
		}
		c.emit(opcode.VLIST, 0)

	case *syntax.RefSpec:
		c.Compile(n.X)

	case *syntax.Ident:
		name := n.String()
		sym, ok := c.scope.lookup(name)
		if !ok {
			c.errorf("undefined identifier %s", name)
		}

		if sym.IsCompiled() {
			c.emit(sym.op, sym.arg)
			break
		}

		// If the referenced node is not compiled yet, we write a
		// placeholder and patch it later.
		c.emit(opcode.REF, Reference(0))

	case *syntax.WithSpec:
		c.emit(opcode.SCAN, 0)
		for _, child := range n.List {
			c.Compile(child)
		}
		c.emit(opcode.BLOCK, 0)

	case *syntax.WithStmt:
		c.Compile(n.Value)
		switch n.KindTok.Kind() {
		case syntax.EXTENSION:
			c.emit(opcode.EXTENSION, 0)
		case syntax.ENCODE:
			c.emit(opcode.ENCODE, 0)
		case syntax.VARIANT:
			c.emit(opcode.VARIANT, 0)
		default:
			c.errorf("unsupported attribute %s", n.KindTok)
		}

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
		addr := c.emit(dir, 0)
		c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})

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
		c.scope = newScope(c.scope)

		op := opcode.TYPE
		if n.With != nil {
			op = opcode.TYPEW
			c.Compile(n.With)
		}
		c.compileStruct(n.KindTok.Kind(), n.Fields)
		c.emit(opcode.NAME, n.Name.String())
		addr := c.emit(op, 0)
		c.scope = c.scope.parent
		c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})

	case *syntax.SubTypeDecl:
		c.scope = newScope(c.scope)

		op := opcode.TYPE
		if n.With != nil {
			op = opcode.TYPEW
			c.Compile(n.With)
		}
		c.Compile(n.Field)
		c.emit(opcode.NAME, n.Field.Name.String())
		addr := c.emit(op, 0)
		c.scope = c.scope.parent
		c.scope.define(n.Field.Name.String(), symbol{n: n.Field, op: opcode.REF, arg: Reference(addr)})

	case *syntax.Field:
		c.compileNestedType(n.Type, n.ValueConstraint, n.LengthConstraint)
		c.compileArrayDef(n.ArrayDef)

	case *syntax.ListSpec:
		c.compileNestedType(n, nil, nil)

	case *syntax.StructSpec:
		c.scope = newScope(c.scope)
		c.compileStruct(n.KindTok.Kind(), n.Fields)
		c.scope = c.scope.parent

	default:
		c.errorf("unexpected node type %T", n)
	}

	return nil
}

func (c *Compiler) compileStruct(k syntax.Kind, fields []*syntax.Field) error {
	if len(fields) == 0 {
		c.emit(opcode.SKIP, 0)
	} else {
		c.emit(opcode.SCAN, 0)
		for _, field := range fields {
			c.Compile(field)
			c.emit(opcode.NAME, field.Name.String())
			switch {
			case field.Optional != nil:
				c.emit(opcode.FIELDO, 0)
			case k == syntax.UNION:
				c.emit(opcode.IFIELD, c.fieldIndex(field.Name.String()))
			default:
				c.emit(opcode.FIELD, 0)
			}

		}
		c.emit(opcode.BLOCK, 0)
	}
	switch k {
	case syntax.RECORD:
		c.emit(opcode.RECORD, 0)
	case syntax.SET:
		c.emit(opcode.SET, 0)
	case syntax.UNION:
		c.emit(opcode.UNION, 0)
	default:
		c.errorf("unsupported struct type %s", k)
	}
	return nil
}

func (c *Compiler) compileNestedType(ty syntax.TypeSpec, vc *syntax.ParenExpr, le *syntax.LengthExpr) error {
	if ls, ok := ty.(*syntax.ListSpec); ok {
		c.compileNestedType(ls.ElemType, vc, le)
		switch ls.KindTok.Kind() {
		case syntax.RECORD:
			c.emit(opcode.RECORDOF, 0)
		case syntax.SET:
			c.emit(opcode.SETOF, 0)
		default:
			c.errorf("unsupported list type %s", ls.KindTok)

		}
		if ls.Length != nil {
			c.emit(opcode.ANY, 0)
			for _, x := range ls.Length.Size.List {
				c.Compile(x)
			}
			c.emit(opcode.LENGTH, 0)
			c.emit(opcode.SUBTYPE, 0)
		}
		return nil
	}

	c.Compile(ty)

	if vc != nil || le != nil {
		switch {
		case vc == nil:
			c.emit(opcode.ANY, 0)
		case len(vc.List) == 1:
			c.Compile(vc.List[0])
		case len(vc.List) > 1:
			c.emit(opcode.MARK, 0)
			for _, x := range vc.List {
				c.Compile(x)
			}
			c.emit(opcode.COLLECT, 0)
		}
		if le != nil {
			for _, x := range le.Size.List {
				c.Compile(x)
			}
			c.emit(opcode.LENGTH, 0)
		}
		c.emit(opcode.SUBTYPE, 0)
	}

	return nil
}

func (c *Compiler) compileFunction(n *syntax.FuncDecl) error {
	c.scope = newScope(c.scope)

	if n.With != nil {
		c.errorf("function attributes not supported")
	}
	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	op := opcode.FUNCTION
	switch {
	case n.Return == nil && n.RunsOn != nil:
		op = opcode.FUNCTIONB
		c.Compile(n.RunsOn.Comp)

	case n.Return != nil && n.RunsOn == nil:
		op = opcode.FUNCTIONV
		c.Compile(n.Return)

	case n.Return != nil && n.RunsOn != nil:
		op = opcode.FUNCTIONVB
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.Return)
	}

	c.Compile(n.Params)
	c.Compile(n.Body)
	c.emit(opcode.NAME, n.Name.String())
	addr := c.emit(op, 0)

	c.scope = c.scope.parent
	c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})
	return nil
}

func (c *Compiler) compileExtFunc(n *syntax.FuncDecl) error {
	c.scope = newScope(c.scope)

	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	if n.RunsOn != nil {
		c.errorf("runs on clause not supported")
	}

	op := opcode.FUNCTIONX
	switch {
	case n.Return == nil && n.With != nil:
		op = opcode.FUNCTIONXW
		c.Compile(n.With)

	case n.Return != nil && n.With == nil:
		op = opcode.FUNCTIONXV
		c.Compile(n.Return)

	case n.Return != nil && n.With != nil:
		op = opcode.FUNCTIONXVW
		c.Compile(n.Return)
		c.Compile(n.With)
	}
	c.Compile(n.Params)
	c.emit(opcode.NAME, n.Name.String())
	addr := c.emit(op, 0)

	c.scope = c.scope.parent
	c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})
	return nil
}

func (c *Compiler) compileTestcase(n *syntax.FuncDecl) error {
	c.scope = newScope(c.scope)

	op := opcode.TESTCASE
	if n.System != nil {
		op = opcode.TESTCASES
		c.Compile(n.System.Comp)
	}
	c.Compile(n.RunsOn.Comp)
	c.Compile(n.Params)
	c.Compile(n.Body)
	c.emit(opcode.NAME, n.Name.String())
	addr := c.emit(op, 0)

	c.scope = c.scope.parent
	c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})
	return nil
}

func (c *Compiler) compileAltstep(n *syntax.FuncDecl) error {
	c.scope = newScope(c.scope)

	if n.Mtc != nil {
		c.errorf("MTC clause not supported")
	}

	op := opcode.ALTSTEP
	switch {
	case n.RunsOn == nil && n.With != nil:
		op = opcode.ALTSTEPW
		c.Compile(n.With)

	case n.RunsOn != nil && n.With == nil:
		op = opcode.ALTSTEPB
		c.Compile(n.RunsOn.Comp)

	case n.RunsOn != nil && n.With != nil:
		op = opcode.ALTSTEPBW
		c.Compile(n.RunsOn.Comp)
		c.Compile(n.With)
	}

	c.Compile(n.Params)
	c.Compile(n.Body)
	c.emit(opcode.NAME, n.Name.String())
	addr := c.emit(op, 0)
	c.scope = c.scope.parent
	c.scope.define(n.Name.String(), symbol{n: n, op: opcode.REF, arg: Reference(addr)})

	return nil
}

func (c *Compiler) compileVar(kind syntax.Kind, restr *syntax.RestrictionSpec, modif syntax.Token, typ syntax.Expr, decl *syntax.Declarator, attrs *syntax.WithSpec) error {
	if attrs != nil {
		c.errorf("attributes not supported")
	}
	c.Compile(typ)
	c.compileArrayDef(decl.ArrayDef)

	c.emit(opcode.NAME, decl.Name.String())
	if modif != nil {
		c.errorf("modifiers not supported")
	}
	addr := c.emit(opcode.VAR, 0)
	if decl.Value != nil {
		c.Compile(decl.Value)
		c.emit(opcode.REF, Reference(addr))
		c.emit(opcode.ASSIGN, 0)
	}
	c.scope.define(decl.Name.String(), symbol{n: decl, op: opcode.REF, arg: Reference(addr)})
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
	c.compileArrayDef(decl.ArrayDef)
	c.emit(opcode.NAME, decl.Name.String())
	if modif != nil {
		c.errorf("modifiers not supported")
	}
	addr := c.emit(op, 0)
	c.scope.define(decl.Name.String(), symbol{n: decl, op: opcode.REF, arg: Reference(addr)})
	return nil
}

func (c *Compiler) compileArrayDef(a []*syntax.ParenExpr) {
	for i := len(a) - 1; i >= 0; i-- {
		c.emit(opcode.SCAN, 0)
		for _, x := range a[i].List {
			c.Compile(x)
		}
		c.emit(opcode.BLOCK, 0)
		c.emit(opcode.ARRAY, 0)
	}
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
	c.compileArrayDef(decl.ArrayDef)

	c.emit(opcode.NAME, decl.Name.String())
	addr := c.emit(op, 0)
	c.scope.define(decl.Name.String(), symbol{n: decl, op: opcode.REF, arg: Reference(addr)})
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
