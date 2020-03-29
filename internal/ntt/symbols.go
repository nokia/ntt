package ntt

import (
	"fmt"
	"sort"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
	"github.com/y0ssar1an/q"
)

// Symbols
func (suite *Suite) Symbols(file string) (*Module, error) {
	syntax, fset, err := suite.Parse(suite.File(file))

	// If we don't a have a syntax tree, we don't need to
	// process any further.
	if syntax == nil {
		return nil, err
	}

	b := newBuilder(fset)

	// Add syntax errors to the error list
	if err, ok := err.(*errors.ErrorList); ok {
		for _, e := range err.List() {
			b.errs = append(b.errs, e)
		}
	}

	b.define(syntax)
	b.resolve(syntax)

	return b.mods[0], &b.errs
}

type Scope interface {

	// Insert attemps to insert an object obj into the Scope. If the scope
	// already contains an alternative object alt with the same name, Insert
	// leaves the scope unchanged and returns altnative object. Otherwise it
	// inserts obj, sets the object's parent scope, if not already set, and
	// returns nil.
	Insert(obj Object) Object

	// Lookup returns the object for a given name. Lookup may follow scope chains.
	Lookup(name string) Object

	// Names lists all names defined in this scope.
	Names() []string
}

// Object describes a named language entity, such as a function or const.
type Object interface {
	Name() string // Object name.

	// Parent returns the (lexical) scope the object is defined in.
	Parent() Scope

	// Type returns the type of the object.
	Type() Type

	// setParent sets the scope the object is defined in.
	setParent(s Scope)

	Range
}

// Range interface is identical to ast.Node interface and helps handling source
// code locations.
type Range interface {
	Pos() loc.Pos
	End() loc.Pos
}

// Module describes a Module.
type Module struct {
	object
	scope
}

func NewModule(rng Range, name string) *Module {
	return &Module{
		object: object{
			pos:  rng.Pos(),
			end:  rng.End(),
			name: name,
		},
	}
}

func (m *Module) Lookup(name string) Object {
	// m.scope.Lookup does not climb up scope chains. When obj != nil we know
	// the scope is m.scope.
	// However we must return m to make sure clients can a type assertions, like
	// 		scp.(*ntt.Module).Name()
	if obj := m.scope.Lookup(name); obj != nil {
		return obj
	}
	return Universe.Lookup(name)
}

// Func describes testcases, altsteps, functions and external functions.
type Func struct {
	object
	external bool
}

func NewFunc(rng Range, name string) *Func {
	return &Func{
		object: object{
			pos:  rng.Pos(),
			end:  rng.End(),
			name: name,
		},
	}
}

// Import describes the view to an imported module.
type Import struct {
	object
	module string
}

func NewImport(rng Range, name string, module string) *Import {
	return &Import{
		object: object{
			pos:  rng.Pos(),
			end:  rng.End(),
			name: name,
		},
		module: module,
	}
}

// Var describes an object, which can hold an value. This could be a local
// variable, a constant, a module parameter or a template.
type Var struct {
	object
}

func NewVar(rng Range, name string) *Var {
	return &Var{
		object: object{
			pos:  rng.Pos(),
			end:  rng.End(),
			name: name,
		},
	}
}

type TypeName struct {
	object
}

func NewTypeName(rng Range, name string, typ Type) *TypeName {
	return &TypeName{
		object: object{
			pos:  rng.Pos(),
			end:  rng.End(),
			name: name,
			typ:  typ,
		},
	}
}

// object implements the common parts of an Object
type object struct {
	pos, end loc.Pos
	name     string
	parent   Scope
	typ      Type
}

// Object interface

func (obj *object) Name() string      { return obj.name }
func (obj *object) Parent() Scope     { return obj.parent }
func (obj *object) Type() Type        { return obj.typ }
func (obj *object) setParent(s Scope) { obj.parent = s }

// Range interface

func (obj *object) Pos() loc.Pos { return obj.pos }
func (obj *object) End() loc.Pos { return obj.end }

type LocalScope struct {
	pos, end loc.Pos
	parent   Scope
	scope
}

func NewLocalScope(rng Range, parent Scope) *LocalScope {
	return &LocalScope{
		pos:    rng.Pos(),
		end:    rng.End(),
		parent: parent,
	}
}

func (ls *LocalScope) Lookup(name string) Object {
	if obj := ls.scope.Lookup(name); obj != nil {
		return obj
	}

	// Ascend into parent scopes.
	if ls.parent != nil {
		return ls.parent.Lookup(name)
	}

	return nil
}

// scope implements the common parts of Scope
type scope struct {
	elems map[string]Object
}

func (s *scope) Insert(obj Object) Object {
	name := obj.Name()
	if alt := s.elems[name]; alt != nil {
		return alt
	}
	if s.elems == nil {
		s.elems = make(map[string]Object)
	}

	s.elems[name] = obj
	return nil
}

func (s *scope) Lookup(name string) Object {
	if obj := s.elems[name]; obj != nil {
		return obj
	}

	return nil
}

func (s *scope) Names() []string {
	names := make([]string, len(s.elems))
	i := 0
	for name := range s.elems {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

type builder struct {
	fset  *loc.FileSet
	errs  errors.ErrorList
	stack []Scope
	mods  []*Module

	scope
	currScope Scope

	// scopes associates every referencing identifier with its current scope.
	scopes map[*ast.Ident]Scope

	// types assiciates every expression with its (calculated) type.
	types map[ast.Expr]Type
}

func newBuilder(fset *loc.FileSet) *builder {
	b := &builder{
		fset:   fset,
		scopes: make(map[*ast.Ident]Scope),
		types:  make(map[ast.Expr]Type),
	}
	b.currScope = &b.scope
	return b
}

// define builds a scope tree in which all identifiers part of a declaration are
// defined. All referencing identifiers will be associated with current scope.
func (b *builder) define(n ast.Node) {
	ast.Apply(n, b.defineEnter, b.defineExit)
}

func (b *builder) defineEnter(c *ast.Cursor) bool {
	switch n := c.Node().(type) {

	case *ast.Ident:
		b.scopes[n] = b.currScope

	case *ast.BlockStmt, *ast.ForStmt:
		b.currScope = NewLocalScope(n, b.currScope)

	case *ast.Module:
		b.defineModule(n)
		return false

	case *ast.ValueDecl:
		b.defineVar(n)
		return false

	case *ast.StructTypeDecl:
		b.defineStruct(n)
		return false

	case *ast.Field:
		b.defineField(n)
		return false
	}
	return true
}

func (b *builder) defineExit(c *ast.Cursor) bool {
	switch c.Node().(type) {
	case *ast.BlockStmt, *ast.ForStmt:
		b.currScope = b.currScope.(*LocalScope).parent
	case *ast.StructTypeDecl:
		b.currScope = b.currScope.(*Struct).parent

	}
	return true
}

func (b *builder) defineModule(n *ast.Module) {
	mod := NewModule(n, n.Name.String())
	b.mods = append(b.mods, mod)
	b.insert(mod)
	b.currScope = mod
	for i := range n.Defs {
		b.define(n.Defs[i])
	}
	b.define(n.With)
	b.currScope = mod.Parent()
}

func (b *builder) defineVar(n *ast.ValueDecl) {
	b.define(n.Type)
	err := ast.Declarators(n, b.fset, func(decl ast.Expr, name ast.Node, arrays []ast.Expr, value ast.Expr) {
		v := NewVar(decl, identName(name))
		b.insert(v)
		for i := range arrays {
			b.define(arrays[i])
		}
		b.define(value)
	})

	// Add syntax errors to the error list
	if err != nil {
		for _, e := range err.List() {
			b.errs = append(b.errs, e)
		}
	}

	b.define(n.With)
}

func (b *builder) defineStruct(n *ast.StructTypeDecl) {
	name := NewTypeName(n, n.Name.String(), nil)
	b.insert(name)

	s := NewStruct(n, name.Parent())
	name.typ = s
	b.currScope = s

	for i := range n.Fields {
		b.define(n.Fields[i])
	}
}

func (b *builder) defineField(n *ast.Field) {
	v := NewVar(n, n.Name.String())
	b.insert(v)
	b.define(n.Type)
}

// resolve resolves all references and types.
func (b *builder) resolve(n ast.Node) {
	ast.Apply(n, b.resolveEnter, b.resolveExit)
}

func (b *builder) resolveEnter(c *ast.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ValueDecl:
		b.resolveVar(n)
		return false
	}
	return true
}

func (b *builder) resolveVar(n *ast.ValueDecl) {
	b.resolve(n.Type)
	ast.Declarators(n, b.fset, func(decl ast.Expr, name ast.Node, arrays []ast.Expr, value ast.Expr) {
		//v := NewVar(decl, identName(name))
		//b.insert(v)
		//for i := range arrays {
		//	b.define(arrays[i])
		//}
		b.resolve(value)
	})
	b.resolve(n.With)
}

func (b *builder) resolveExit(c *ast.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ValueLiteral:
		b.types[n] = literalType(n.Tok.Kind)

	case *ast.Ident:
		scp := b.scopes[n]

		// Identifier which to not have a scope are part of declarations and can
		// be skipped.
		if scp == nil {
			break
		}

		def := scp.Lookup(n.String())
		if def == nil {
			b.errorf(n, "unknown identifier %q", n.String())
			break
		}

		// In local scopes, check if declaration comes after
		if _, ok := def.Parent().(*LocalScope); ok {
			if def.End() >= n.Pos() {
				b.errorf(n, "unknown identifier %q", n.String())
			}
		}
		q.Q("IDENT", n, def, def.Type())
		b.types[n] = def.Type()

	case *ast.ParametrizedIdent:
		b.types[n] = b.types[n.Ident]

	case *ast.SelectorExpr:
		typ := b.types[n.X]
		if typ == nil {
			q.Q("NOT FOUND", n.X)
			break
		}

		id := identName(n.Sel)
		if scp, ok := typ.(Scope); ok {
			if obj := scp.Lookup(id); obj != nil {
				b.types[n] = obj.Type()
			}
		}
		b.errorf(n.Sel, "unknown identifer %q", id)

	case *ast.UnaryExpr:
		b.types[n] = b.types[n.X]

	case *ast.BinaryExpr:
		switch n.Op.Kind {
		case token.ASSIGN:
			b.types[n] = b.types[n.X]
		case token.COLON:
			b.types[n] = b.types[n.X]
		case token.RANGE:
			b.types[n] = Typ[Integer]
		case token.OR:
			b.types[n] = Typ[Boolean]
		case token.XOR:
			b.types[n] = Typ[Boolean]
		case token.AND:
			b.types[n] = Typ[Boolean]
		case token.NOT:
			b.types[n] = Typ[Boolean]
		case token.EQ, token.NE:
			b.types[n] = Typ[Boolean]
		case token.LT, token.LE, token.GT, token.GE:
			b.types[n] = Typ[Boolean]
		case token.SHR, token.SHL, token.ROR, token.ROL:
			b.types[n] = Typ[String]
		case token.OR4B:
			b.types[n] = Typ[String]
		case token.XOR4B:
			b.types[n] = Typ[String]
		case token.AND4B:
			b.types[n] = Typ[String]
		case token.NOT4B:
			b.types[n] = Typ[String]
		case token.CONCAT:
			b.types[n] = Typ[String]
		case token.ADD, token.SUB:
			b.types[n] = Typ[Numerical]
		case token.MUL, token.DIV, token.REM, token.MOD:
			b.types[n] = Typ[Numerical]
		}
	}
	return true
}

// error records errors during definition phase, such like ErrRedefined, ...
func (b *builder) errorf(rng Range, format string, args ...interface{}) {
	b.errs.Add(b.fset.Position(rng.Pos()), fmt.Sprintf(format, args...))
}

// insert object into current scope.
func (b *builder) insert(obj Object) {
	if alt := b.currScope.Insert(obj); alt != nil {
		// TODO(5nord) Make nicer errors: On what location is which object
		// defined.
		b.errorf(obj, "redefinition of %q", obj.Name())
		return
	}
	if obj.Parent() == nil {
		obj.setParent(b.currScope)
	}
}

func identName(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Ident:
		return n.String()
	case *ast.ParametrizedIdent:
		return n.Ident.String()
	}
	return "_"
}

func literalType(tok token.Kind) Type {
	switch tok {
	case token.INT:
		return Typ[Integer]

	case token.FLOAT:
		return Typ[Float]

	case token.NAN:
		return Typ[Float]

	case token.ANY, token.MUL:
		return Typ[Template]

	case token.NULL:
		return Typ[Component]

	case token.OMIT:
		return Typ[Omit]

	case token.FALSE, token.TRUE:
		return Typ[Boolean]

	case token.NONE, token.INCONC, token.PASS, token.FAIL, token.ERROR:
		return Typ[Verdict]

	case token.BSTRING:
		// TODO(5nord) Implement hexstring, octetstring, ...
		return Typ[Bitstring]

	case token.STRING:
		// TODO(5nord) Implement universal charstring
		return Typ[Charstring]

	default:
		return Typ[Invalid]
	}
}
