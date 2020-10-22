package ast

import (
	"fmt"

	"github.com/nokia/ntt/internal/ttcn3/token"
)

// First returns the first valid token of a syntax tree
func FirstToken(n Node) *Token {
	switch n := n.(type) {
	case Token:
		return &n
	case *ErrorNode:
		return &n.From
	case *Ident:
		return &n.Tok
	case *ParametrizedIdent:
		return FirstToken(n.Ident)
	case *ValueLiteral:
		return &n.Tok
	case *CompositeLiteral:
		return &n.LBrace
	case *UnaryExpr:
		return &n.Op
	case *BinaryExpr:
		if n.X != nil {
			return FirstToken(n.X)
		}
		return &n.Op

	case *ParenExpr:
		return &n.LParen
	case *SelectorExpr:
		return FirstToken(n.X)
	case *IndexExpr:
		if n.X != nil {
			return FirstToken(n.X)
		}
		return &n.LBrack
	case *CallExpr:
		return FirstToken(n.Fun)
	case *LengthExpr:
		if n.X != nil {
			return FirstToken(n.X)
		}
		return &n.Len
	case *RedirectExpr:
		return FirstToken(n.X)
	case *TemplateDecl:
		return FirstToken(&n.RestrictionSpec)
	case *ValueExpr:
		return FirstToken(n.X)
	case *ParamExpr:
		return FirstToken(n.X)
	case *FromExpr:
		return &n.Kind
	case *ModifiesExpr:
		return &n.Tok
	case *RegexpExpr:
		return &n.Tok
	case *PatternExpr:
		return &n.Tok
	case *DecmatchExpr:
		return &n.Tok
	case *DecodedExpr:
		return &n.Tok
	case *DefKindExpr:
		return &n.Kind
	case *ExceptExpr:
		return FirstToken(n.X)
	case *BlockStmt:
		return &n.LBrace
	case *DeclStmt:
		return FirstToken(n.Decl)
	case *ExprStmt:
		return FirstToken(n.Expr)
	case *BranchStmt:
		return &n.Tok
	case *ReturnStmt:
		return &n.Tok
	case *CallStmt:
		return FirstToken(n.Stmt)
	case *AltStmt:
		return &n.Tok
	case *ForStmt:
		return &n.Tok
	case *WhileStmt:
		return &n.Tok
	case *DoWhileStmt:
		return &n.DoTok
	case *IfStmt:
		return &n.Tok
	case *SelectStmt:
		return &n.Tok
	case *CaseClause:
		return &n.Tok
	case *CommClause:
		return &n.LBrack
	case *Field:
		if n.DefaultTok.IsValid() {
			return &n.DefaultTok
		}
		return FirstToken(n.Type)
	case *RefSpec:
		return FirstToken(n.X)
	case *StructSpec:
		return &n.Kind
	case *ListSpec:
		return &n.Kind
	case *EnumSpec:
		return &n.Tok
	case *BehaviourSpec:
		return &n.Kind
	case *ValueDecl:
		if n.Kind.IsValid() {
			return &n.Kind
		}
		return FirstToken(n.Type)

	case *Declarator:
		if n.Name != nil {
			return FirstToken(n.Name)
		}
		if len(n.ArrayDef) > 0 {
			return FirstToken(n.ArrayDef[0])
		}
		if n.AssignTok.IsValid() {
			return &n.AssignTok
		}
		if n.Value != nil {
			return FirstToken(n.Value)
		}
		panic("ast.Node contains no tokens. This is probably a parser error")

	case *ModuleParameterGroup:
		return &n.Tok

	case *FuncDecl:
		if n.External.Kind == token.EXTERNAL {
			return &n.External
		}
		return &n.Kind

	case *SignatureDecl:
		return &n.Tok
	case *SubTypeDecl:
		return &n.TypeTok
	case *StructTypeDecl:
		return &n.TypeTok
	case *EnumTypeDecl:
		return &n.TypeTok
	case *BehaviourTypeDecl:
		return &n.TypeTok
	case *PortTypeDecl:
		return &n.TypeTok
	case *PortAttribute:
		return &n.Kind
	case *PortMapAttribute:
		return &n.MapTok
	case *ComponentTypeDecl:
		return &n.TypeTok
	case *Module:
		return &n.Tok
	case *ModuleDef:
		if n.Visibility.IsValid() {
			return &n.Visibility
		}
		return FirstToken(n.Def)
	case *ControlPart:
		return &n.Tok
	case *ImportDecl:
		return &n.ImportTok
	case *GroupDecl:
		return &n.Tok
	case *FriendDecl:
		return &n.FriendTok
	case *LanguageSpec:
		return &n.Tok
	case *RestrictionSpec:
		if n.TemplateTok.IsValid() {
			return &n.TemplateTok
		}
		return &n.Tok
	case *RunsOnSpec:
		return &n.RunsTok
	case *SystemSpec:
		return &n.Tok
	case *MtcSpec:
		return &n.Tok
	case *ReturnSpec:
		return &n.Tok
	case *FormalPars:
		return &n.LParen
	case *FormalPar:
		if n.Direction.IsValid() {
			return &n.Direction
		}
		if n.TemplateRestriction != nil {
			return FirstToken(n.TemplateRestriction)
		}
		if n.Modif.IsValid() {
			return &n.Modif
		}
		return FirstToken(n.Type)
	case *WithSpec:
		return &n.Tok
	case *WithStmt:
		return &n.Kind
	default:
		panic(fmt.Sprintf("unknown ast.Node: %T", n))
	}
}
