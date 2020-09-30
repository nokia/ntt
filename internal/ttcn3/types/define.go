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
	info.descent(m)
}

// descent inserts all declarations into their enclosing scope. It also tracks
// the scopes of referencing identifiers.
func (info *Info) descent(n ast.Node) {
	ast.Apply(n, func(c *ast.Cursor) bool {
		switch n := c.Node().(type) {

		case *ast.Ident:
			info.Scopes[n] = info.currScope
			return true

		case *ast.BlockStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			for i := range n.Stmts {
				info.descent(n.Stmts[i])
			}
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.ForStmt:
			info.currScope = NewLocalScope(n, info.currScope)
			info.descent(n.Init)
			info.descent(n.Cond)
			info.descent(n.Post)
			info.descent(n.Body)
			info.currScope = info.currScope.(*LocalScope).parent
			return false

		case *ast.Module:
			info.currMod = NewModule(n, n.Name.String())
			info.Modules[n.Name.String()] = info.currMod
			info.currScope = info.currMod
			for i := range n.Defs {
				info.descent(n.Defs[i])
			}
			if n.With == nil {
				info.descent(n.With)
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
				info.descent(n.Module)
			}
			for i := range n.List {
				info.descent(n.List[i])
			}
			return false

		case *ast.Declarator:
			v := NewVar(n, identName(n.Name))
			info.insert(v)

			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i])
			}
			info.descent(n.Value)
			return false

		case *ast.TemplateDecl:
			sym := NewVar(n, n.Name.String())
			info.insert(sym)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.descent(n.TypePars)
			}
			info.descent(n.Type)
			info.descent(n.Base)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.descent(n.Params)
			}
			info.descent(n.Value)
			info.descent(n.With)

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
				info.descent(n.TypePars)
			}
			info.descent(n.RunsOn)
			info.descent(n.Mtc)
			info.descent(n.System)
			info.descent(n.Return)
			if n.Params != nil {
				info.currScope = NewLocalScope(n.Params, info.currScope)
				info.descent(n.Params)
			}
			info.descent(n.Body)
			info.descent(n.With)

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
				info.descent(n.Fields[i])
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
						//log.Debug(fmt.Sprintf("Enumlabels with value:%s[%d]=%s", name.Name(), i, enumLabel.Name()))
					}
				case *ast.Ident:
					enumLabel := NewVar(l, l.String())
					enumLabel.typ = name
					info.insert(enumLabel)
				default:
					//info.error(errors.Error{Pos: l.Pos()., Msg: "Expected label or label(integer value)"})
					continue

				}
			}

			if n.With == nil {
				info.descent(n.With)
			}
			return false

		case *ast.ComponentTypeDecl:
			c := NewComponentType(n.Name, n.Name.String())
			info.insert(c)
			if n.TypePars != nil {
				info.currScope = NewLocalScope(n.TypePars, info.currScope)
				info.descent(n.TypePars)
			}
			if n.Extends != nil {
				for i := range n.Extends {
					info.descent(n.Extends[i])
				}
			}

			info.currScope = c
			info.descent(n.Body)
			info.currScope = c.parent
			return false

		case *ast.Field:
			v := NewVar(n.Name, identName(n.Name))
			info.insert(v)
			info.descent(n.TypePars)

			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i])
			}
			info.descent(n.Type)
			info.descent(n.ValueConstraint)
			info.descent(n.LengthConstraint)
			return false

		case *ast.FormalPar:
			info.descent(n.Type)
			v := NewVar(n.Name, identName(n.Name))

			info.insert(v)
			for i := range n.ArrayDef {
				info.descent(n.ArrayDef[i])
			}
			info.descent(n.Value)

			return false

		case *ast.PortMapAttribute:
			info.currScope = NewLocalScope(n, info.currScope)
			info.descent(n.Params)
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
