package ttcn3

import (
	"fmt"
	"reflect"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

type Scope struct {
	ast.Node
	Tree  *Tree
	Names map[string]*Definition
}

type Definition struct {
	*ast.Ident
	Tree *Tree
	Next *Definition
}

func (scp *Scope) Insert(id *ast.Ident) {
	if scp.Names == nil {
		scp.Names = make(map[string]*Definition)
	}

	name := id.String()
	scp.Names[name] = &Definition{
		Ident: id,
		Tree:  scp.Tree,
		Next:  scp.Names[name],
	}
}

func NewScope(n ast.Node, tree *Tree) *Scope {
	scp := &Scope{
		Node: n,
		Tree: tree,
	}

	switch n := n.(type) {
	case *ast.TemplateDecl:
		scp.add(n.TypePars)
		scp.add(n.Params)

	case *ast.FuncDecl:
		scp.add(n.TypePars)
		scp.add(n.Params)
		scp.addBody(n.Body)

	case *ast.SignatureDecl:
		scp.add(n.TypePars)
		scp.add(n.Params)

	case *ast.SubTypeDecl:
		scp.add(n.Field)

	case *ast.StructTypeDecl:
		scp.add(n.TypePars)
		for _, n := range n.Fields {
			scp.add(n)
		}

	case *ast.EnumTypeDecl:
		scp.add(n.TypePars)
		for _, n := range n.Enums {
			scp.addEnum(n)
		}

	case *ast.BehaviourTypeDecl:
		scp.add(n.TypePars)
		scp.add(n.Params)

	case *ast.PortTypeDecl:
		scp.add(n.TypePars)

	case *ast.PortMapAttribute:
		scp.add(n.Params)

	case *ast.ComponentTypeDecl:
		scp.add(n.TypePars)
		scp.addBody(n.Body)

	case *ast.BlockStmt:
		scp.addBody(n)

	case *ast.AltStmt:
		scp.addBody(n.Body)

	case *ast.CallStmt:
		scp.addBody(n.Body)

	case *ast.ForStmt:
		scp.add(n.Init)
		scp.addBody(n.Body)

	case *ast.WhileStmt:
		scp.addBody(n.Body)

	case *ast.DoWhileStmt:
		scp.addBody(n.Body)

	case *ast.IfStmt:
		scp.add(n.Then)
		scp.add(n.Else)

	case *ast.CaseClause:
		scp.addBody(n.Body)

	case *ast.CommClause:
		scp.addBody(n.Body)

	case *ast.Field:
		scp.add(n.TypePars)

	case *ast.StructSpec:
		for _, n := range n.Fields {
			scp.add(n)
		}

	case *ast.EnumSpec:
		for _, n := range n.Enums {
			scp.addEnum(n)
		}

	case *ast.BehaviourSpec:
		scp.add(n.Params)

	case *ast.CompositeLiteral:
		for _, n := range n.List {
			scp.add(n)
		}

	case *ast.SelectorExpr:
		// TODO(5nord) Don't forget
		//n.X   Expr  // Preceding expression (might be nil)
		//n.Sel Expr  // Literal, identifier or reference.

	case *ast.Module:
		ast.Inspect(n, func(n ast.Node) bool {
			switch n := n.(type) {
			// Groups are not visible in the global scope.
			case *ast.GroupDecl:

			case *ast.ModuleDef:
				scp.add(n.Def)
			case *ast.EnumTypeDecl:
				for _, n := range n.Enums {
					scp.addEnum(n)
				}
			case *ast.EnumSpec:
				for _, n := range n.Enums {
					scp.addEnum(n)
				}

			}
			return true
		})

	case *ast.ControlPart:
		scp.addBody(n.Body)

	default:
		return nil
	}
	return scp
}

func (scp *Scope) addEnum(n ast.Node) {
	switch n := n.(type) {
	case *ast.CallExpr:
		if n, ok := n.Fun.(*ast.Ident); ok {
			scp.Insert(n)
		}
	case *ast.Ident:
		scp.Insert(n)
	default:
		log.Debugf("scopes.go: unknown enumeration syntax: %T", n)
	}
}

func (scp *Scope) addBody(n *ast.BlockStmt) {
	for _, stmt := range n.Stmts {
		scp.add(stmt)
	}
}

// add adds definitions to the scope;
func (scp *Scope) add(n ast.Node) error {
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() || n == nil {
		return nil
	}
	switch n := n.(type) {

	case *ast.ModuleDef:
		scp.add(n.Def)

	case *ast.TemplateDecl:
		scp.Insert(n.Name)

	case *ast.ValueDecl:
		for _, n := range n.Decls {
			scp.add(n)
		}

	case *ast.Declarator:
		scp.Insert(n.Name)

	case *ast.FuncDecl:
		scp.Insert(n.Name)

	case *ast.SignatureDecl:
		scp.Insert(n.Name)

	case *ast.SubTypeDecl:
		scp.add(n.Field)

	case *ast.StructTypeDecl:
		scp.Insert(n.Name)

	case *ast.EnumTypeDecl:
		scp.Insert(n.Name)

	case *ast.BehaviourTypeDecl:
		scp.Insert(n.Name)

	case *ast.PortTypeDecl:
		scp.Insert(n.Name)

	case *ast.ComponentTypeDecl:
		scp.Insert(n.Name)

	case *ast.DeclStmt:
		scp.add(n.Decl)

	case *ast.BranchStmt:
		if n.Tok.Kind == token.LABEL {
			scp.Insert(n.Label)
		}

	case *ast.Field:
		scp.Insert(n.Name)

	case *ast.Module:
		scp.Insert(n.Name)

	case *ast.ControlPart:
		// TODO(5nord) Add control part names to scope

	case *ast.ImportDecl:
		scp.Insert(n.Module)

	case *ast.GroupDecl:
		// GroupDecl are not added to the scope, but their members are.
		for _, n := range n.Defs {
			scp.add(n)
		}

	case *ast.FormalPars:
		for _, n := range n.List {
			scp.add(n)
		}

	case ast.NodeList:
		for _, n := range n {
			scp.add(n)
		}

	case *ast.FormalPar:
		scp.Insert(n.Name)
	}

	return fmt.Errorf("%T has not declaration", n)
}
