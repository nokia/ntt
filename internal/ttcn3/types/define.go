package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (info *Info) Define(n ast.Node) {
	info.currScope = nil
	if info.Scopes == nil {
		info.Scopes = make(map[*ast.Ident]Scope)
	}
	if info.Modules == nil {
		info.Modules = make(map[string]*Module)
	}
	info.descent(n, nil)
}

// descent inserts all declarations into their enclosing scope. It also tracks
// the scopes of referencing identifiers.
func (info *Info) descent(n ast.Node, scp Scope) {
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case ast.NodeList:
			for _, n := range n {
				info.descent(n, info.currScope)
			}
			return false

		case *ast.Ident:
			info.Scopes[n] = info.currScope
			return true

		case *ast.BlockStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			for i := range n.Stmts {
				info.descent(n.Stmts[i], info.currScope)
			}
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.ForStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			info.descent(n.Init, info.currScope)
			info.descent(n.Cond, info.currScope)
			info.descent(n.Post, info.currScope)
			info.descent(n.Body, info.currScope)
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.Module:
			info.currMod = NewModule(n, n.Name.String())
			info.Modules[n.Name.String()] = info.currMod
			info.currScope = info.currMod
			for i := range n.Defs {
				info.descent(n.Defs[i], info.currScope)
			}
			if n.With == nil {
				info.descent(n.With, info.currScope)
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

			if n.Module != nil {
				info.descent(n.Module, info.currScope)
			}
			for i := range n.List {
				info.descent(n.List[i], info.currScope)
			}
			return false

		case *ast.Declarator:
			v := NewVar(n, ast.Name(n.Name))
			info.insert(v)

			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i], info.currScope)
			}
			info.descent(n.Value, info.currScope)
			return false

		case *ast.TemplateDecl:
			sym := NewVar(n, n.Name.String())
			info.insert(sym)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.descent(n.TypePars, info.currScope)
			}
			info.descent(n.Type, info.currScope)
			info.descent(n.Base, info.currScope)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.descent(n.Params, info.currScope)
			}
			info.descent(n.Value, info.currScope)
			info.descent(n.With, info.currScope)

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
				info.descent(n.TypePars, info.currScope)
			}
			info.descent(n.RunsOn, info.currScope)
			info.descent(n.Mtc, info.currScope)
			info.descent(n.System, info.currScope)
			info.descent(n.Return, info.currScope)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.descent(n.Params, info.currScope)
			}
			info.descent(n.Body, info.currScope)
			info.descent(n.With, info.currScope)

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
				info.descent(n.Fields[i], info.currScope)
			}

			info.currScope = name.Parent()
			return false

		case *ast.EnumTypeDecl:
			name := NewEnumeratedType(n.Name, n.Name.String(), nil)
			info.insert(name)

			// enumerated labels are in the global scope too
			for _, l := range n.Enums {
				switch l := l.(type) {
				case *ast.CallExpr:
					id := l.Fun
					switch id := id.(type) {
					case *ast.Ident:
						enumLabel := NewVar(id, id.String())
						enumLabel.typ = name
						info.insert(enumLabel)
					}
				case *ast.Ident:
					enumLabel := NewVar(l, l.String())
					enumLabel.typ = name
					info.insert(enumLabel)
				default:
					continue
				}
			}

			if n.With != nil {
				info.descent(n.With, info.currScope)
			}
			return false

		case *ast.ComponentTypeDecl:
			c := NewComponentType(n.Name, n.Name.String())
			info.insert(c)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.descent(n.TypePars, info.currScope)
			}
			if n.Extends != nil {
				for i := range n.Extends {
					info.descent(n.Extends[i], info.currScope)
				}
			}

			info.currScope = c
			info.descent(n.Body, info.currScope)
			info.currScope = c.parent
			return false

		case *ast.Field:
			v := NewVar(n.Name, ast.Name(n.Name))
			info.insert(v)
			info.descent(n.TypePars, info.currScope)

			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i], info.currScope)
			}
			info.descent(n.Type, info.currScope)
			info.descent(n.ValueConstraint, info.currScope)
			info.descent(n.LengthConstraint, info.currScope)
			return false

		case *ast.FormalPar:
			info.descent(n.Type, info.currScope)
			v := NewVar(n.Name, ast.Name(n.Name))

			info.insert(v)
			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i], info.currScope)
			}
			info.descent(n.Value, info.currScope)

			return false

		case *ast.PortMapAttribute:
			info.currScope = NewLocalScope(n, info.currScope)
			info.descent(n.Params, info.currScope)
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		default:
			return true
		}
	})
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
