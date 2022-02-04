package ast

import (
	"fmt"

	"github.com/nokia/ntt/ttcn3/token"
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

// Name returns the name of a Node. If the node has no name (like statements)
// Name will return an empty string.
func Name(n Node) string {
	switch n := n.(type) {
	case *Ident:
		if n == nil {
			return ""
		}
		return n.String()
	case *SelectorExpr:
		name := Name(n.X)
		if n.Sel != nil {
			name += "." + Name(n.Sel)
		}
		return name
	case *BranchStmt:
		if n.Tok.Kind == token.LABEL {
			return Name(n.Label)
		}
	case *CallExpr:
		return Name(n.Fun)
	case *LengthExpr:
		return Name(n.X)
	case *ParametrizedIdent:
		return Name(n.Ident)
	case *Module:
		return Name(n.Name)
	case *Field:
		return Name(n.Name)
	case *PortTypeDecl:
		return Name(n.Name)
	case *ComponentTypeDecl:
		return Name(n.Name)
	case *StructTypeDecl:
		return Name(n.Name)
	case *EnumTypeDecl:
		return Name(n.Name)
	case *BehaviourTypeDecl:
		return Name(n.Name)
	case *Declarator:
		return Name(n.Name)
	case *FormalPar:
		return Name(n.Name)
	case *TemplateDecl:
		return Name(n.Name)
	case *FuncDecl:
		return Name(n.Name)
	case *RefSpec:
		return Name(n.X)
	case *SignatureDecl:
		return Name(n.Name)
	}
	return ""
}

func Children(n Node) []Node {
	var children []Node
	switch n := n.(type) {
	case Token:
	case *ErrorNode:
		children = append(children, n.From)
		children = append(children, n.To)

	case *Ident:
		children = append(children, n.Tok)
		children = append(children, n.Tok2)

	case *ParametrizedIdent:
		children = append(children, n.Ident)
		children = append(children, n.Params)

	case *ValueLiteral:
		children = append(children, n.Tok)

	case *CompositeLiteral:
		children = append(children, n.LBrace)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *UnaryExpr:
		children = append(children, n.Op)
		children = append(children, n.X)

	case *BinaryExpr:
		children = append(children, n.X)
		children = append(children, n.Op)
		children = append(children, n.Y)

	case *ParenExpr:
		children = append(children, n.LParen)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RParen)

	case *SelectorExpr:
		children = append(children, n.X)
		children = append(children, n.Dot)
		children = append(children, n.Sel)

	case *IndexExpr:
		children = append(children, n.X)
		children = append(children, n.LBrack)
		children = append(children, n.Index)
		children = append(children, n.RBrack)

	case *CallExpr:
		children = append(children, n.Fun)
		children = append(children, n.Args)

	case *LengthExpr:
		children = append(children, n.X)
		children = append(children, n.Len)
		children = append(children, n.Size)

	case *RedirectExpr:
		children = append(children, n.X)
		children = append(children, n.Tok)
		children = append(children, n.ValueTok)
		for _, n := range n.Value {
			children = append(children, n)
		}
		children = append(children, n.ParamTok)
		for _, n := range n.Param {
			children = append(children, n)
		}
		children = append(children, n.SenderTok)
		children = append(children, n.Sender)
		children = append(children, n.IndexTok)
		children = append(children, n.IndexValueTok)
		children = append(children, n.Index)
		children = append(children, n.TimestampTok)
		children = append(children, n.Timestamp)

	case *ValueExpr:
		children = append(children, n.X)
		children = append(children, n.Tok)
		children = append(children, n.Y)

	case *ParamExpr:
		children = append(children, n.X)
		children = append(children, n.Tok)
		children = append(children, n.Y)

	case *FromExpr:
		children = append(children, n.Kind)
		children = append(children, n.FromTok)
		children = append(children, n.X)

	case *ModifiesExpr:
		children = append(children, n.Tok)
		children = append(children, n.X)
		children = append(children, n.Assign)
		children = append(children, n.Y)

	case *RegexpExpr:
		children = append(children, n.Tok)
		children = append(children, n.NoCase)
		children = append(children, n.X)

	case *PatternExpr:
		children = append(children, n.Tok)
		children = append(children, n.NoCase)
		children = append(children, n.X)

	case *DecmatchExpr:
		children = append(children, n.Tok)
		children = append(children, n.Params)
		children = append(children, n.X)

	case *DecodedExpr:
		children = append(children, n.Tok)
		children = append(children, n.Params)
		children = append(children, n.X)

	case *DefKindExpr:
		children = append(children, n.Kind)
		for _, n := range n.List {
			children = append(children, n)
		}

	case *ExceptExpr:
		children = append(children, n.X)
		children = append(children, n.ExceptTok)
		children = append(children, n.LBrace)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *BlockStmt:
		children = append(children, n.LBrace)
		for _, n := range n.Stmts {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *DeclStmt:
		children = append(children, n.Decl)

	case *ExprStmt:
		children = append(children, n.Expr)

	case *BranchStmt:
		children = append(children, n.Tok)
		children = append(children, n.Label)

	case *ReturnStmt:
		children = append(children, n.Tok)
		children = append(children, n.Result)

	case *AltStmt:
		children = append(children, n.Tok)
		children = append(children, n.Body)

	case *CallStmt:
		children = append(children, n.Stmt)
		children = append(children, n.Body)

	case *ForStmt:
		children = append(children, n.Tok)
		children = append(children, n.LParen)
		children = append(children, n.Init)
		children = append(children, n.InitSemi)
		children = append(children, n.Cond)
		children = append(children, n.CondSemi)
		children = append(children, n.Post)
		children = append(children, n.RParen)
		children = append(children, n.Body)

	case *WhileStmt:
		children = append(children, n.Tok)
		children = append(children, n.Cond)
		children = append(children, n.Body)

	case *DoWhileStmt:
		children = append(children, n.DoTok)
		children = append(children, n.Body)
		children = append(children, n.WhileTok)
		children = append(children, n.Cond)

	case *IfStmt:
		children = append(children, n.Tok)
		children = append(children, n.Cond)
		children = append(children, n.Then)
		children = append(children, n.ElseTok)
		children = append(children, n.Else)

	case *SelectStmt:
		children = append(children, n.Tok)
		children = append(children, n.Union)
		children = append(children, n.Tag)
		children = append(children, n.LBrace)
		for _, n := range n.Body {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *CaseClause:
		children = append(children, n.Tok)
		children = append(children, n.Case)
		children = append(children, n.Body)

	case *CommClause:
		children = append(children, n.LBrack)
		children = append(children, n.X)
		children = append(children, n.Else)
		children = append(children, n.RBrack)
		children = append(children, n.Comm)
		children = append(children, n.Body)

	case *Field:
		children = append(children, n.DefaultTok)
		children = append(children, n.Type)
		children = append(children, n.Name)
		for _, n := range n.ArrayDef {
			children = append(children, n)
		}
		children = append(children, n.TypePars)
		children = append(children, n.ValueConstraint)
		children = append(children, n.LengthConstraint)
		children = append(children, n.Optional)

	case *RefSpec:
		children = append(children, n.X)

	case *StructSpec:
		children = append(children, n.Kind)
		children = append(children, n.LBrace)
		for _, n := range n.Fields {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *ListSpec:
		children = append(children, n.Kind)
		children = append(children, n.Length)
		children = append(children, n.OfTok)
		children = append(children, n.ElemType)

	case *EnumSpec:
		children = append(children, n.Tok)
		children = append(children, n.LBrace)
		for _, n := range n.Enums {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *BehaviourSpec:
		children = append(children, n.Kind)
		children = append(children, n.Params)
		children = append(children, n.RunsOn)
		children = append(children, n.System)
		children = append(children, n.Return)

	case *ValueDecl:
		children = append(children, n.Kind)
		children = append(children, n.TemplateRestriction)
		children = append(children, n.Modif)
		children = append(children, n.Type)
		for _, n := range n.Decls {
			children = append(children, n)
		}
		children = append(children, n.With)

	case *Declarator:
		children = append(children, n.Name)
		for _, n := range n.ArrayDef {
			children = append(children, n)
		}
		children = append(children, n.AssignTok)
		children = append(children, n.Value)

	case *TemplateDecl:
		children = append(children, &n.RestrictionSpec)
		children = append(children, n.Modif)
		children = append(children, n.Type)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.Params)
		children = append(children, n.ModifiesTok)
		children = append(children, n.Base)
		children = append(children, n.AssignTok)
		children = append(children, n.Value)
		children = append(children, n.With)

	case *ModuleParameterGroup:
		children = append(children, n.Tok)
		children = append(children, n.LBrace)
		for _, n := range n.Decls {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *FuncDecl:
		children = append(children, n.External)
		children = append(children, n.Kind)
		children = append(children, n.Name)
		children = append(children, n.Modif)
		children = append(children, n.TypePars)
		children = append(children, n.Params)
		children = append(children, n.RunsOn)
		children = append(children, n.Mtc)
		children = append(children, n.System)
		children = append(children, n.Return)
		children = append(children, n.Body)
		children = append(children, n.With)

	case *SignatureDecl:
		children = append(children, n.Tok)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.Params)
		children = append(children, n.NoBlock)
		children = append(children, n.Return)
		children = append(children, n.ExceptionTok)
		children = append(children, n.Exception)
		children = append(children, n.With)

	case *SubTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.Field)
		children = append(children, n.With)

	case *StructTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.Kind)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.LBrace)
		for _, n := range n.Fields {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *EnumTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.EnumTok)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.LBrace)
		for _, n := range n.Enums {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *BehaviourTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.Kind)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.Params)
		children = append(children, n.RunsOn)
		children = append(children, n.System)
		children = append(children, n.Return)
		children = append(children, n.With)

	case *PortTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.PortTok)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.Kind)
		children = append(children, n.Realtime)
		children = append(children, n.LBrace)
		for _, n := range n.Attrs {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *PortAttribute:
		children = append(children, n.Kind)
		for _, n := range n.Types {
			children = append(children, n)
		}

	case *PortMapAttribute:
		children = append(children, n.MapTok)
		children = append(children, n.ParamTok)
		children = append(children, n.Params)

	case *ComponentTypeDecl:
		children = append(children, n.TypeTok)
		children = append(children, n.CompTok)
		children = append(children, n.Name)
		children = append(children, n.TypePars)
		children = append(children, n.ExtendsTok)
		for _, n := range n.Extends {
			children = append(children, n)
		}
		children = append(children, n.Body)
		children = append(children, n.With)

	case *Module:
		children = append(children, n.Tok)
		children = append(children, n.Name)
		children = append(children, n.Language)
		children = append(children, n.LBrace)
		for _, n := range n.Defs {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *ModuleDef:
		children = append(children, n.Visibility)
		children = append(children, n.Def)

	case *ControlPart:
		children = append(children, n.Tok)
		children = append(children, n.Body)
		children = append(children, n.With)

	case *ImportDecl:
		children = append(children, n.ImportTok)
		children = append(children, n.FromTok)
		children = append(children, n.Module)
		children = append(children, n.Language)
		children = append(children, n.LBrace)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *GroupDecl:
		children = append(children, n.Tok)
		children = append(children, n.Name)
		children = append(children, n.LBrace)
		for _, n := range n.Defs {
			children = append(children, n)
		}
		children = append(children, n.RBrace)
		children = append(children, n.With)

	case *FriendDecl:
		children = append(children, n.FriendTok)
		children = append(children, n.ModuleTok)
		children = append(children, n.Module)
		children = append(children, n.With)

	case *LanguageSpec:
		children = append(children, n.Tok)
		for _, n := range n.List {
			children = append(children, n)
		}

	case *RestrictionSpec:
		children = append(children, n.TemplateTok)
		children = append(children, n.LParen)
		children = append(children, n.Tok)
		children = append(children, n.RParen)

	case *RunsOnSpec:
		children = append(children, n.RunsTok)
		children = append(children, n.OnTok)
		children = append(children, n.Comp)

	case *SystemSpec:
		children = append(children, n.Tok)
		children = append(children, n.Comp)

	case *MtcSpec:
		children = append(children, n.Tok)
		children = append(children, n.Comp)

	case *ReturnSpec:
		children = append(children, n.Tok)
		children = append(children, n.Restriction)
		children = append(children, n.Modif)
		children = append(children, n.Type)

	case *FormalPars:
		children = append(children, n.LParen)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RParen)

	case *FormalPar:
		children = append(children, n.Direction)
		children = append(children, n.TemplateRestriction)
		children = append(children, n.Modif)
		children = append(children, n.Type)
		children = append(children, n.Name)
		for _, n := range n.ArrayDef {
			children = append(children, n)
		}
		children = append(children, n.AssignTok)
		children = append(children, n.Value)

	case *WithSpec:
		children = append(children, n.Tok)
		children = append(children, n.LBrace)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RBrace)

	case *WithStmt:
		children = append(children, n.Kind)
		children = append(children, n.Override)
		children = append(children, n.LParen)
		for _, n := range n.List {
			children = append(children, n)
		}
		children = append(children, n.RParen)
		children = append(children, n.Value)
	}
	return children
}
