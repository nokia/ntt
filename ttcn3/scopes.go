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
	Node ast.Node
	Tree *Tree
	Next *Definition
}

func (scp *Scope) Insert(n ast.Node, id *ast.Ident) {
	if scp.Names == nil {
		scp.Names = make(map[string]*Definition)
	}

	if id != nil {
		name := id.String()
		scp.Names[name] = &Definition{
			Ident: id,
			Node:  n,
			Tree:  scp.Tree,
			Next:  scp.Names[name],
		}
	}
}

// Lookup returns a list of defintions for the given identifier.
// Lookup may be called with nil as receiver.
func (scp *Scope) Lookup(name string) []*Definition {
	if scp == nil {
		return nil
	}
	var defs []*Definition
	def := scp.Names[name]
	for def != nil {
		defs = append(defs, def)
		def = def.Next
	}
	return defs
}

// NewScope builts and populares a new scope from the given syntax node.
// NewScope returns nil if no valid scope could be built.
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
		if n.Field != nil {
			scp.addField(n.Field)
		}

	case *ast.Field:
		scp.addField(n)

	case *ast.StructTypeDecl:
		scp.add(n.TypePars)
		for _, n := range n.Fields {
			scp.add(n)
		}

	case *ast.EnumTypeDecl:
		scp.add(n.TypePars)
		for _, e := range n.Enums {
			scp.addEnum(n, e)
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

	case *ast.StructSpec:
		for _, n := range n.Fields {
			scp.add(n)
		}

	case *ast.EnumSpec:
		for _, e := range n.Enums {
			scp.addEnum(n, e)
		}

	case *ast.BehaviourSpec:
		scp.add(n.Params)

	case *ast.Module:
		ast.Inspect(n, func(n ast.Node) bool {
			switch n := n.(type) {
			// Groups are not visible in the global scope.
			case *ast.GroupDecl:

			case *ast.ModuleDef:
				scp.add(n.Def)
			case *ast.EnumTypeDecl:
				for _, e := range n.Enums {
					scp.addEnum(n, e)
				}
			case *ast.EnumSpec:
				for _, e := range n.Enums {
					scp.addEnum(n, e)
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

func (scp *Scope) addEnum(n ast.Node, e ast.Expr) {
	switch e := e.(type) {
	case *ast.CallExpr:
		if e, ok := e.Fun.(*ast.Ident); ok {
			scp.Insert(n, e)
		}
	case *ast.Ident:
		scp.Insert(n, e)
	default:
		log.Debugf("scopes.go: unknown enumeration syntax: %T", n)
	}
}

func (scp *Scope) addField(n *ast.Field) {
	scp.add(n.Type)
	scp.add(n.TypePars)
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
		scp.Insert(n, n.Name)

	case *ast.ValueDecl:
		for _, d := range n.Decls {
			scp.Insert(n, d.Name)
		}

	case *ast.FuncDecl:
		scp.Insert(n, n.Name)

	case *ast.SignatureDecl:
		scp.Insert(n, n.Name)

	case *ast.SubTypeDecl:
		scp.add(n.Field)

	case *ast.StructTypeDecl:
		scp.Insert(n, n.Name)

	case *ast.EnumTypeDecl:
		scp.Insert(n, n.Name)

	case *ast.BehaviourTypeDecl:
		scp.Insert(n, n.Name)

	case *ast.PortTypeDecl:
		scp.Insert(n, n.Name)

	case *ast.ComponentTypeDecl:
		scp.Insert(n, n.Name)

	case *ast.DeclStmt:
		scp.add(n.Decl)

	case *ast.BranchStmt:
		if n.Tok.Kind == token.LABEL {
			scp.Insert(n, n.Label)
		}

	case *ast.Field:
		scp.Insert(n, n.Name)

	case *ast.Module:
		scp.Insert(n, n.Name)

	case *ast.ControlPart:
		// TODO(5nord) Add control part names to scope

	case *ast.ImportDecl:
		scp.Insert(n, n.Module)

	case *ast.GroupDecl:
		// GroupDecl are not added to the scope, but their members are.
		for _, n := range n.Defs {
			scp.add(n)
		}

	case *ast.StructSpec:
		for _, n := range n.Fields {
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
		scp.Insert(n, n.Name)
	}

	return fmt.Errorf("%T is not a declaration", n)
}
