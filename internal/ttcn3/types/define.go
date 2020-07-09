package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (info *Info) Define(m *ast.Module) {
	info.currScope = nil
	if info.Scopes == nil {
		info.Scopes = make(map[*ast.Ident]Scope)
	}
	if info.Modules == nil {
		info.Modules = make(map[string]*Module)
	}
	info.define(m)
}

func (info *Info) define(n ast.Node) {
	ast.Apply(n, func(c *ast.Cursor) bool {
		switch n := c.Node().(type) {

		case *ast.Ident:
			info.Scopes[n] = info.currScope
			return true

		case *ast.BlockStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			for i := range n.Stmts {
				info.define(n.Stmts[i])
			}
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.ForStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			info.define(n.Init)
			info.define(n.Cond)
			info.define(n.Post)
			info.define(n.Body)
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.Module:
			info.currMod = NewModule(n, n.Name.String())
			info.Modules[n.Name.String()] = info.currMod
			info.currScope = info.currMod
			for i := range n.Defs {
				info.define(n.Defs[i])
			}
			if n.With == nil {
				info.define(n.With)
			}

			info.currScope = nil
			return false

		case *ast.ImportDecl:
			name := n.Module.String()

			if info.Import != nil {
				if err := info.Import(name); err != nil {
					info.error(err)
				}
			}

			info.currMod.Imports = append(info.currMod.Imports, n.Module.String())
			return false

		case *ast.ValueDecl:
			info.define(n.Type)
			err := ast.Declarators(n.Decls, info.Fset, func(decl ast.Expr, name ast.Node, arrays []ast.Expr, value ast.Expr) {
				v := NewVar(decl, identName(name))
				info.insert(v)
				for i := range arrays {
					info.define(arrays[i])
				}
				info.define(value)
			})

			// Add syntax errors to the error list
			if err != nil {
				for _, e := range err.List() {
					info.error(e)
				}
			}

			info.define(n.With)
			return false

		case *ast.TemplateDecl:
			sym := NewVar(n, n.Name.String())
			info.insert(sym)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.define(n.TypePars)
			}
			info.define(n.Type)
			info.define(n.Base)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.define(n.Params)
			}
			info.define(n.Value)
			info.define(n.With)

			if n.Params != nil {
				info.currScope = info.currScope.(*LocalScope).parent
			}

			if n.TypePars != nil {
				info.currScope = info.currScope.(*LocalScope).parent
			}
			return false

		case *ast.FuncDecl:
			sym := NewFunc(n, n.Name.String())
			info.insert(sym)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.define(n.TypePars)
			}
			info.define(n.RunsOn)
			info.define(n.Mtc)
			info.define(n.System)
			info.define(n.Return)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.define(n.Params)
			}
			info.define(n.Body)
			info.define(n.With)

			if n.Params != nil {
				info.currScope = info.currScope.(*LocalScope).parent
			}

			if n.TypePars != nil {
				info.currScope = info.currScope.(*LocalScope).parent
			}
			return false

		case *ast.StructTypeDecl:
			name := NewTypeName(n.Name, n.Name.String(), nil)
			info.insert(name)

			s := NewStruct(n, name.Parent())
			name.typ = s
			info.currScope = s

			for i := range n.Fields {
				info.define(n.Fields[i])
			}

			info.currScope = name.Parent()
			return false

		case *ast.ComponentTypeDecl:
			c := NewComponentType(n, n.Name.String())
			info.insert(c)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.define(n.TypePars)
			}
			if n.Extends != nil {
				for i := range n.Extends {
					info.define(n.Extends[i])
				}
			}

			info.currScope = c
			info.define(n.Body)
			info.currScope = c.parent
			return false

		case *ast.Field:
			err := ast.Declarators([]ast.Expr{n.Name}, info.Fset, func(decl ast.Expr, name ast.Node, arrays []ast.Expr, value ast.Expr) {
				v := NewVar(decl, identName(name))
				info.insert(v)
				for i := range arrays {
					info.define(arrays[i])
				}
				info.define(value)
			})

			// Add syntax errors to the error list
			if err != nil {
				for _, e := range err.List() {
					info.error(e)
				}
			}
			info.define(n.Type)
			return false

		case *ast.FormalPar:
			info.define(n.Type)
			err := ast.Declarators([]ast.Expr{n.Name}, info.Fset, func(decl ast.Expr, name ast.Node, arrays []ast.Expr, value ast.Expr) {
				v := NewVar(decl, identName(name))
				info.insert(v)
				for i := range arrays {
					info.define(arrays[i])
				}
				info.define(value)
			})

			// Add syntax errors to the error list
			if err != nil {
				for _, e := range err.List() {
					info.error(e)
				}
			}
			return false

		case *ast.PortMapAttribute:
			info.currScope = NewLocalScope(n, info.currScope)
			info.define(n.Params)
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		default:
			return true
		}
	}, nil)
}

// insert object into current scope.
func (info *Info) insert(obj Object) {
	if alt := info.currScope.Insert(obj); alt != nil {
		oldPos := info.Fset.Position(alt.Pos())
		newPos := info.Fset.Position(obj.Pos())
		info.error(&RedefinitionError{Name: obj.Name(), OldPos: oldPos, NewPos: newPos})
		return
	}
	if obj.Parent() == nil {
		obj.setParent(info.currScope)
	}
}
