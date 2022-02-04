package ttcn3

import (
	"fmt"

	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

type Scope struct {
	Names map[string]*Definition
}

type Definition struct {
	*ast.Ident

	Next *Definition
}

func (scp *Scope) add(n ast.Node) error {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
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
		if n.Field != nil {
			scp.add(n.Field)
		}

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
		// TODO(5nord) Add group names to scope

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

func (scp *Scope) Insert(id *ast.Ident) {
	if scp.Names == nil {
		scp.Names = make(map[string]*Definition)
	}

	name := id.String()
	scp.Names[name] = &Definition{
		Ident: id,
		Next:  scp.Names[name],
	}
}

func NewScope(n ast.Node) *Scope {
	scp := &Scope{}

	switch n := n.(type) {
	case *ast.BlockStmt:
		for _, stmt := range n.Stmts {
			scp.add(stmt)
		}
	default:
		return nil
	}
	return scp
}
