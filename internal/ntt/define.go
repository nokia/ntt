package ntt

import (
	"fmt"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (suite *Suite) Define(fset *loc.FileSet, mod *ast.Module) *Module {
	v := &visitor{
		Module: Module{
			syntax: mod,
			scope:  scope{},
			object: object{
				name: mod.Name.String(),
			},
			Imports:   make([]*Import, 0, 10), // 10 is the rough average we experience so far.
			Tests:     make([]*Function, 0, len(mod.Defs)),
			Functions: make([]*Function, 0, len(mod.Defs)),
			Altsteps:  make([]*Function, 0, len(mod.Defs)),
			links:     make(map[*ast.Ident]Scope),
		},
		scopes: make([]Scope, 0, 25),
		fset:   fset,
	}

	v.pushScope(v)
	for _, d := range mod.Defs {
		ast.Walk(v, d)
	}
	v.popScope()

	return &v.Module
}

type visitor struct {
	Module

	fset   *loc.FileSet
	scopes []Scope
	errs   errors.ErrorList
}

// error records errors during definition phase, such like ErrRedefined, ...
func (v *visitor) error(pos loc.Pos, msg string) {
	v.errs.Add(v.fset.Position(pos), msg)
}

// pushScope pushes Scope s on a stack and makes it the active scope.
func (v *visitor) pushScope(s Scope) {
	v.scopes = append(v.scopes, s)
}

// popScope pops the active scope.
func (v *visitor) popScope() {
	v.scopes = v.scopes[:len(v.scopes)-1]
}

// currScope returns the active scope or nil if there isn't any.
func (v *visitor) currScope() Scope {
	if len(v.scopes) > 0 {
		return v.scopes[len(v.scopes)-1]
	}
	return nil
}

// insert object into current scope.
func (v *visitor) insert(obj Object) {
	if alt := v.currScope().Insert(obj); alt != nil {
		// TODO(5nord) Make nicer errors: On what location is which object
		// defined.
		v.error(obj.Pos(), fmt.Sprintf("redefinition of %q", obj.Name()))
	}
}

// link current scope with identifier for later resolve.
func (v *visitor) linkScope(id *ast.Ident) {
	v.links[id] = v.currScope()
}

func (v *visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}

	switch n := n.(type) {

	// Various Declarations
	// --------------------
	//

	case *ast.FuncDecl:
		v.addFunction(n)

	case *ast.ValueDecl:
		v.addDecl(n)

	// Parameter handling
	// ------------------
	// (type parameters, template parameters, ...
	//
	//

	case *ast.FormalPars:
		// TODO(5nord) before we push the parameter scope we should register the
		// scope of default parameters, to prevent this error:
		//
		//         function foo(integer x:= 1 + x) { ... }
		v.pushScope(newScope(n.Pos(), n.End(), v.currScope()))
		return v

	// Handle Statements with anonymous scope.
	//
	//
	case *ast.BlockStmt, *ast.ForStmt:
		v.pushScope(newScope(n.Pos(), n.End(), v.currScope()))
		return v

	// Expressions
	// -----------
	//
	// Traverse Expressions and link (non-definition) identifiers with active
	// scopes. This link provides the context for a later resolve stage.
	case *ast.Ident:
		v.linkScope(n) // link scopes and identifiers for later resolve stage.
		return v

	case ast.Expr:
		return v

	// Import Handling Related Nodes
	// -----------------------------
	//
	// TODO(5nord) Full import support (with exclude).
	case *ast.ImportDecl:
		v.addImport(n)

	// TODO(5nord) Add support for import groups.
	case *ast.GroupDecl:
		return v

	// TODO(5nord) Add support for visibility.
	case *ast.ModuleDef:
		return v

	// TODOs
	case *ast.ReturnSpec,
		*ast.RestrictionSpec,
		*ast.MtcSpec,
		*ast.RunsOnSpec:
		return v

	default:
		return v
		//panic(fmt.Sprintf("unexpected ast.Node %#v", n))

	}
	return nil
}

func (v *visitor) addImport(n *ast.ImportDecl) {
	imp := newImport(n, v.currScope())
	v.Imports = append(v.Imports, imp)
}

func (v *visitor) addFunction(n *ast.FuncDecl) {
	fun := newFunction(n, v.currScope())

	v.insert(fun)
	switch {
	case fun.IsTest():
		v.Tests = append(v.Tests, fun)
	case fun.IsAltstep():
		v.Altsteps = append(v.Altsteps, fun)
	default:
		v.Functions = append(v.Functions, fun)
	}

	// Traverse type parameters first, because they are in their own scope
	// and must be available for function parameters, return type and
	// component references.
	if n.TypePars != nil {
		ast.Walk(v, n.TypePars)
	}

	if n.RunsOn != nil {
		ast.Walk(v, n.RunsOn)
	}

	if n.Mtc != nil {
		ast.Walk(v, n.Mtc)
	}

	if n.System != nil {
		ast.Walk(v, n.System)
	}

	if n.Return != nil {
		ast.Walk(v, n.Return)
	}

	if n.Params != nil {
		ast.Walk(v, n.Params)
	}

	if n.Body != nil {
		ast.Walk(v, n.Body)
	}

	if n.With != nil {
		ast.Walk(v, n.With)
	}
}

func (v *visitor) addDecl(n *ast.ValueDecl) {

	ast.Walk(v, n.Type)

	err := ast.Declarators(n, v.fset, func(name ast.Node, arrays []ast.Expr, value ast.Expr) {
		var s string
		switch name := name.(type) {
		case *ast.Ident:
			s = name.String()
		case *ast.ParametrizedIdent:
			s = name.Ident.String()
		}

		decl := newDecl(s)
		v.insert(decl)

		for i := range arrays {
			ast.Walk(v, arrays[i])
		}

		ast.Walk(v, value)

	})

	if err != nil {
		v.errs = append(v.errs, err.(errors.ErrorList)...)
	}

	if n.With != nil {
		ast.Walk(v, n.With)
	}
}
