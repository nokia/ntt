package ast

import (
	"fmt"
)

// WalkModuleDefs calls fun for every module definition.
func WalkModuleDefs(fun func(def *ModuleDef) bool, nodes ...Node) {
	for _, n := range nodes {
		switch x := n.(type) {
		case *Module:
			walkModuleDefs(fun, x.Defs...)
		case *GroupDecl:
			walkModuleDefs(fun, x.Defs...)
		case *ModuleDef:
			if g, ok := x.Def.(*GroupDecl); ok {
				WalkModuleDefs(fun, g)
				return
			}
			if !fun(x) {
				return
			}
		}
	}
}

func walkModuleDefs(fun func(def *ModuleDef) bool, defs ...*ModuleDef) {
	for _, d := range defs {
		WalkModuleDefs(fun, d)
	}
}

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(v Visitor, node Node) {
	if node == nil {
		return
	}

	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *ErrorNode:

	case *Ident:

	case *ParametrizedIdent:
		if n.Ident != nil {
			Walk(v, n.Ident)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}

	case *ValueLiteral:

	case *CompositeLiteral:
		walkExprList(v, n.List)

	case *UnaryExpr:
		Walk(v, n.X)

	case *BinaryExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	case *ParenExpr:
		walkExprList(v, n.List)

	case *SelectorExpr:
		Walk(v, n.X)
		Walk(v, n.Sel)

	case *IndexExpr:
		Walk(v, n.X)
		Walk(v, n.Index)

	case *CallExpr:
		Walk(v, n.Fun)
		if n.Args != nil {
			Walk(v, n.Args)
		}

	case *LengthExpr:
		Walk(v, n.X)
		if n.Size != nil {
			Walk(v, n.Size)
		}

	case *RedirectExpr:
		Walk(v, n.X)
		walkExprList(v, n.Value)
		walkExprList(v, n.Param)
		Walk(v, n.Sender)
		Walk(v, n.Index)
		Walk(v, n.Timestamp)

	case *ValueExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	case *ParamExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	case *FromExpr:
		Walk(v, n.X)

	case *ModifiesExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	case *RegexpExpr:
		Walk(v, n.X)

	case *PatternExpr:
		Walk(v, n.X)

	case *DecmatchExpr:
		Walk(v, n.Params)
		Walk(v, n.X)

	case *DecodedExpr:
		Walk(v, n.Params)
		Walk(v, n.X)

	case *DefKindExpr:
		walkExprList(v, n.List)

	case *ExceptExpr:
		Walk(v, n.X)
		walkExprList(v, n.List)

	// Statements
	// -------------------------

	case *BlockStmt:
		walkStmtList(v, n.Stmts)

	case *DeclStmt:
		Walk(v, n.Decl)

	case *ExprStmt:
		Walk(v, n.Expr)

	case *BranchStmt:
		if n.Label != nil {
			Walk(v, n.Label)
		}

	case *ReturnStmt:
		Walk(v, n.Result)

	case *AltStmt:
		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *CallStmt:
		Walk(v, n.Stmt)
		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *ForStmt:
		Walk(v, n.Init)
		Walk(v, n.Cond)
		Walk(v, n.Post)
		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *WhileStmt:
		if n.Cond != nil {
			Walk(v, n.Cond)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *DoWhileStmt:
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.Cond != nil {
			Walk(v, n.Cond)
		}

	case *IfStmt:
		Walk(v, n.Cond)
		if n.Then != nil {
			Walk(v, n.Then)
		}
		Walk(v, n.Else)

	case *SelectStmt:
		if n.Tag != nil {
			Walk(v, n.Tag)
		}
		walkCaseClauseList(v, n.Body)

	case *CaseClause:
		if n.Case != nil {
			Walk(v, n.Case)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}

	case *CommClause:
		Walk(v, n.X)
		Walk(v, n.Comm)
		if n.Body != nil {
			Walk(v, n.Body)
		}

	// TypeSpecs
	// ----------------

	case *Field:
		Walk(v, n.Type)
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkParenExprList(v, n.ArrayDef)
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		if n.ValueConstraint != nil {
			Walk(v, n.ValueConstraint)
		}
		if n.LengthConstraint != nil {
			Walk(v, n.LengthConstraint)
		}

	case *RefSpec:
		Walk(v, n.X)

	case *StructSpec:
		walkFieldList(v, n.Fields)

	case *ListSpec:
		if n.Length != nil {
			Walk(v, n.Length)
		}
		Walk(v, n.ElemType)

	case *EnumSpec:
		walkExprList(v, n.Enums)

	case *BehaviourSpec:
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.RunsOn != nil {
			Walk(v, n.RunsOn)
		}
		if n.System != nil {
			Walk(v, n.System)
		}
		if n.Return != nil {
			Walk(v, n.Return)
		}

	// Declarations
	// ----------------------
	case *ValueDecl:
		if n.TemplateRestriction != nil {
			Walk(v, n.TemplateRestriction)
		}
		Walk(v, n.Type)
		walkDeclList(v, n.Decls)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *Declarator:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkParenExprList(v, n.ArrayDef)
		Walk(v, n.Value)

	case *TemplateDecl:
		Walk(v, n.Type)
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}
		Walk(v, n.Base)
		Walk(v, n.Value)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *ModuleParameterGroup:
		walkValueDeclList(v, n.Decls)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *FuncDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.RunsOn != nil {
			Walk(v, n.RunsOn)
		}
		if n.Mtc != nil {
			Walk(v, n.Mtc)
		}
		if n.System != nil {
			Walk(v, n.System)
		}
		if n.Return != nil {
			Walk(v, n.Return)
		}
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *SignatureDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.Return != nil {
			Walk(v, n.Return)
		}
		if n.Exception != nil {
			Walk(v, n.Exception)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *SubTypeDecl:
		if n.Field != nil {
			Walk(v, n.Field)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *StructTypeDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		walkFieldList(v, n.Fields)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *EnumTypeDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		walkExprList(v, n.Enums)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *BehaviourTypeDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.RunsOn != nil {
			Walk(v, n.RunsOn)
		}
		if n.System != nil {
			Walk(v, n.System)
		}
		if n.Return != nil {
			Walk(v, n.Return)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *PortTypeDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		walkNodeList(v, n.Attrs)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *PortAttribute:
		walkExprList(v, n.Types)

	case *PortMapAttribute:
		if n.Params != nil {
			Walk(v, n.Params)
		}

	case *ComponentTypeDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.TypePars != nil {
			Walk(v, n.TypePars)
		}
		walkExprList(v, n.Extends)
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *Module:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Language != nil {
			Walk(v, n.Language)
		}
		walkModuleDefList(v, n.Defs)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *ModuleDef:
		Walk(v, n.Def)

	case *ControlPart:
		if n.Body != nil {
			Walk(v, n.Body)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	case *ImportDecl:
		if n.Module != nil {
			Walk(v, n.Module)
		}
		if n.Language != nil {
			Walk(v, n.Language)
		}
		walkDefKindExprList(v, n.List)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *GroupDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkModuleDefList(v, n.Defs)
		if n.With != nil {
			Walk(v, n.With)
		}

	case *FriendDecl:
		if n.Module != nil {
			Walk(v, n.Module)
		}
		if n.With != nil {
			Walk(v, n.With)
		}

	// Misc
	// -------

	case *LanguageSpec:

	case *RestrictionSpec:

	case *RunsOnSpec:
		Walk(v, n.Comp)

	case *SystemSpec:
		Walk(v, n.Comp)

	case *MtcSpec:
		Walk(v, n.Comp)

	case *ReturnSpec:
		if n.Restriction != nil {
			Walk(v, n.Restriction)
		}
		Walk(v, n.Type)

	case *FormalPars:
		walkFormalParList(v, n.List)

	case *FormalPar:
		if n.TemplateRestriction != nil {
			Walk(v, n.TemplateRestriction)
		}
		Walk(v, n.Type)
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkParenExprList(v, n.ArrayDef)
		Walk(v, n.Value)

	case *WithSpec:
		walkWithStmtList(v, n.List)

	case *WithStmt:
		walkExprList(v, n.List)
		Walk(v, n.Value)

	case NodeList:
		for _, n := range n {
			Walk(v, n)
		}
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func walkCaseClauseList(v Visitor, list []*CaseClause) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkDefKindExprList(v Visitor, list []*DefKindExpr) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkExprList(v Visitor, list []Expr) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkFieldList(v Visitor, list []*Field) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkParenExprList(v Visitor, list []*ParenExpr) {
	for _, n := range list {
		Walk(v, n)
	}
}
func walkFormalParList(v Visitor, list []*FormalPar) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkModuleDefList(v Visitor, list []*ModuleDef) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkNodeList(v Visitor, list []Node) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkDeclList(v Visitor, list []*Declarator) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkStmtList(v Visitor, list []Stmt) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkValueDeclList(v Visitor, list []*ValueDecl) {
	for _, n := range list {
		Walk(v, n)
	}
}

func walkWithStmtList(v Visitor, list []*WithStmt) {
	for _, n := range list {
		Walk(v, n)
	}
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
//
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
