// Package ast provides TTCN-3 syntax tree nodes and functions for tree
// traversal.
package ast

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/token"
)

func (x ErrorNode) Pos() loc.Pos    { return x.From.Pos() }
func (x ErrorNode) End() loc.Pos    { return x.To.End() }
func (x ErrorNode) LastTok() *Token { return x.To.LastTok() }

func (n *NodeList) Pos() loc.Pos {
	if len(n.Nodes) == 0 {
		return loc.NoPos
	}
	return n.Nodes[0].Pos()
}

func (n *NodeList) End() loc.Pos {
	if tok := n.LastTok(); tok != nil {
		return tok.End()
	}
	return loc.NoPos
}

func (n *NodeList) LastTok() *Token {
	if len(n.Nodes) == 0 {
		return nil
	}
	return n.Nodes[len(n.Nodes)-1].LastTok()
}

func (x *Ident) LastTok() *Token {
	if x.Tok2.LastTok().IsValid() {
		return x.Tok2.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ParametrizedIdent) LastTok() *Token { return x.Params.LastTok() }
func (x *ValueLiteral) LastTok() *Token      { return x.Tok.LastTok() }
func (x *CompositeLiteral) LastTok() *Token  { return x.RBrace.LastTok() }
func (x *UnaryExpr) LastTok() *Token         { return x.X.LastTok() }
func (x *BinaryExpr) LastTok() *Token        { return x.Y.LastTok() }

func (x *ParenExpr) LastTok() *Token { return x.RParen.LastTok() }
func (x *SelectorExpr) LastTok() *Token {
	if x.Sel != nil {
		return x.Sel.LastTok()
	}
	return x.Dot.LastTok()
}

func (x *IndexExpr) LastTok() *Token  { return x.RBrack.LastTok() }
func (x *CallExpr) LastTok() *Token   { return x.Args.LastTok() }
func (x *LengthExpr) LastTok() *Token { return x.Size.LastTok() }
func (x *RedirectExpr) LastTok() *Token {
	if x.Timestamp != nil {
		return x.Timestamp.LastTok()
	}
	if x.Index != nil {
		return x.Index.LastTok()
	}
	if x.Sender != nil {
		return x.Sender.LastTok()
	}
	if x.Param != nil {
		return x.Param[len(x.Param)-1].LastTok()
	}
	if x.Value != nil {
		return x.Value[len(x.Value)-1].LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ValueExpr) LastTok() *Token    { return x.Y.LastTok() }
func (x *ParamExpr) LastTok() *Token    { return x.Y.LastTok() }
func (x *FromExpr) LastTok() *Token     { return x.X.LastTok() }
func (x *ModifiesExpr) LastTok() *Token { return x.Y.LastTok() }
func (x *RegexpExpr) LastTok() *Token   { return x.X.LastTok() }
func (x *PatternExpr) LastTok() *Token  { return x.X.LastTok() }
func (x *DecmatchExpr) LastTok() *Token { return x.X.LastTok() }
func (x *DecodedExpr) LastTok() *Token  { return x.X.LastTok() }

func (x *DefKindExpr) LastTok() *Token {
	if len(x.List) > 0 {
		return x.List[len(x.List)-1].LastTok()
	}
	return &x.Kind
}

func (x *ExceptExpr) LastTok() *Token {
	if x.RBrace.LastTok().IsValid() {
		return x.RBrace.LastTok()
	}
	return x.List[len(x.List)-1].LastTok()
}

func (x *Ident) Pos() loc.Pos             { return x.Tok.Pos() }
func (x *ParametrizedIdent) Pos() loc.Pos { return x.Ident.Pos() }
func (x *ValueLiteral) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *CompositeLiteral) Pos() loc.Pos  { return x.LBrace.Pos() }
func (x *UnaryExpr) Pos() loc.Pos         { return x.Op.Pos() }
func (x *BinaryExpr) Pos() loc.Pos {
	if x.X != nil {
		return x.X.Pos()
	}
	return x.Op.Pos()
}

func (x *ParenExpr) Pos() loc.Pos    { return x.LParen.Pos() }
func (x *SelectorExpr) Pos() loc.Pos { return x.X.Pos() }
func (x *IndexExpr) Pos() loc.Pos {
	if x.X != nil {
		return x.X.Pos()
	}
	return x.LBrack.Pos()
}
func (x *CallExpr) Pos() loc.Pos { return x.Fun.Pos() }

func (x *LengthExpr) Pos() loc.Pos {
	if x.X != nil {
		return x.X.Pos()
	}
	return x.Len.Pos()
}

func (x *RedirectExpr) Pos() loc.Pos { return x.X.Pos() }
func (x *ValueExpr) Pos() loc.Pos    { return x.X.Pos() }
func (x *ParamExpr) Pos() loc.Pos    { return x.X.Pos() }
func (x *FromExpr) Pos() loc.Pos     { return x.Kind.Pos() }
func (x *ModifiesExpr) Pos() loc.Pos { return x.Tok.Pos() }
func (x *RegexpExpr) Pos() loc.Pos   { return x.Tok.Pos() }
func (x *PatternExpr) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *DecmatchExpr) Pos() loc.Pos { return x.Tok.Pos() }
func (x *DecodedExpr) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *DefKindExpr) Pos() loc.Pos {
	if x.Kind.IsValid() {
		return x.Kind.Pos()
	}
	if len(x.List) > 0 {
		return x.List[0].Pos()
	}
	return loc.NoPos
}
func (x *ExceptExpr) Pos() loc.Pos { return x.X.Pos() }

func (x *Ident) End() loc.Pos             { return x.LastTok().End() }
func (x *ParametrizedIdent) End() loc.Pos { return x.LastTok().End() }
func (x *ValueLiteral) End() loc.Pos      { return x.LastTok().End() }
func (x *CompositeLiteral) End() loc.Pos  { return x.LastTok().End() }
func (x *UnaryExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *BinaryExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *ParenExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *SelectorExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *IndexExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *CallExpr) End() loc.Pos          { return x.LastTok().End() }
func (x *LengthExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *RedirectExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *ValueExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *ParamExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *FromExpr) End() loc.Pos          { return x.LastTok().End() }
func (x *ModifiesExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *RegexpExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *PatternExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *DecmatchExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *DecodedExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *DefKindExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *ExceptExpr) End() loc.Pos        { return x.LastTok().End() }

func (x *BlockStmt) LastTok() *Token { return x.RBrace.LastTok() }
func (x *DeclStmt) LastTok() *Token  { return x.Decl.LastTok() }
func (x *ExprStmt) LastTok() *Token  { return x.Expr.LastTok() }

func (x *BranchStmt) LastTok() *Token {
	if x.Label != nil {
		return x.Label.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ReturnStmt) LastTok() *Token {
	if x.Result != nil {
		return x.Result.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *CallStmt) LastTok() *Token    { return x.Body.LastTok() }
func (x *AltStmt) LastTok() *Token     { return x.Body.LastTok() }
func (x *ForStmt) LastTok() *Token     { return x.Body.LastTok() }
func (x *WhileStmt) LastTok() *Token   { return x.Body.LastTok() }
func (x *DoWhileStmt) LastTok() *Token { return x.Cond.LastTok() }

func (x *IfStmt) LastTok() *Token {
	if x.Else != nil {
		return x.Else.LastTok()
	}
	return x.Then.LastTok()
}

func (x *SelectStmt) LastTok() *Token { return x.RBrace.LastTok() }
func (x *CaseClause) LastTok() *Token { return x.Body.LastTok() }

func (x *CommClause) LastTok() *Token {
	if x.Body != nil {
		return x.Body.LastTok()
	}
	return x.Comm.LastTok()
}

func (x *BlockStmt) Pos() loc.Pos   { return x.LBrace.Pos() }
func (x *DeclStmt) Pos() loc.Pos    { return x.Decl.Pos() }
func (x *ExprStmt) Pos() loc.Pos    { return x.Expr.Pos() }
func (x *BranchStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *ReturnStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CallStmt) Pos() loc.Pos    { return x.Stmt.Pos() }
func (x *AltStmt) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *ForStmt) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *WhileStmt) Pos() loc.Pos   { return x.Tok.Pos() }
func (x *DoWhileStmt) Pos() loc.Pos { return x.DoTok.Pos() }
func (x *IfStmt) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *SelectStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CaseClause) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CommClause) Pos() loc.Pos  { return x.LBrack.Pos() }

func (x *BlockStmt) End() loc.Pos   { return x.LastTok().End() }
func (x *DeclStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *ExprStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *BranchStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *ReturnStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *CallStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *AltStmt) End() loc.Pos     { return x.LastTok().End() }
func (x *ForStmt) End() loc.Pos     { return x.LastTok().End() }
func (x *WhileStmt) End() loc.Pos   { return x.LastTok().End() }
func (x *DoWhileStmt) End() loc.Pos { return x.LastTok().End() }
func (x *IfStmt) End() loc.Pos      { return x.LastTok().End() }
func (x *SelectStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *CaseClause) End() loc.Pos  { return x.LastTok().End() }
func (x *CommClause) End() loc.Pos  { return x.LastTok().End() }

func (x *Field) LastTok() *Token {
	if x.Optional.LastTok().IsValid() {
		return x.Optional.LastTok()
	}
	if x.LengthConstraint != nil {
		return x.LengthConstraint.LastTok()
	}
	if x.ValueConstraint != nil {
		return x.ValueConstraint.LastTok()
	}
	if x.TypePars != nil {
		return x.TypePars.LastTok()
	}
	if l := len(x.ArrayDef); l != 0 {
		return x.ArrayDef[l-1].LastTok()
	}
	if x.Name != nil {
		return x.Name.LastTok()
	}
	return x.Type.LastTok()
}

func (x *RefSpec) LastTok() *Token    { return x.X.LastTok() }
func (x *StructSpec) LastTok() *Token { return x.RBrace.LastTok() }
func (x *ListSpec) LastTok() *Token   { return x.ElemType.LastTok() }
func (x *EnumSpec) LastTok() *Token   { return x.RBrace.LastTok() }

func (x *BehaviourSpec) LastTok() *Token {
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.System != nil {
		return x.System.LastTok()
	}
	if x.RunsOn != nil {
		return x.RunsOn.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Kind.LastTok()
}

func (x *Field) Pos() loc.Pos {
	if x.DefaultTok.Pos().IsValid() {
		return x.DefaultTok.Pos()
	}
	return x.Type.Pos()
}

func (x *RefSpec) Pos() loc.Pos       { return x.X.Pos() }
func (x *StructSpec) Pos() loc.Pos    { return x.Kind.Pos() }
func (x *ListSpec) Pos() loc.Pos      { return x.Kind.Pos() }
func (x *EnumSpec) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *BehaviourSpec) Pos() loc.Pos { return x.Kind.Pos() }

func (x *Field) End() loc.Pos         { return x.LastTok().End() }
func (x *RefSpec) End() loc.Pos       { return x.LastTok().End() }
func (x *StructSpec) End() loc.Pos    { return x.LastTok().End() }
func (x *ListSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *EnumSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *BehaviourSpec) End() loc.Pos { return x.LastTok().End() }

func (x *ValueDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if len(x.Decls) > 0 {
		return x.Decls[len(x.Decls)-1].LastTok()
	}
	if x.Type != nil {
		return x.Type.LastTok()
	}
	if x.Modif.IsValid() {
		return x.Modif.LastTok()
	}
	if x.TemplateRestriction != nil {
		return x.Type.LastTok()
	}
	return x.Kind.LastTok()
}

func (x *Declarator) LastTok() *Token {
	if x.Value != nil {
		return x.Value.LastTok()
	}
	if x.AssignTok.IsValid() {
		return x.AssignTok.LastTok()
	}
	if l := len(x.ArrayDef); l > 0 {
		return x.ArrayDef[l-1].LastTok()
	}
	if x.Name != nil {
		return x.Name.LastTok()
	}
	return &Token{}
}

func (x *TemplateDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Value != nil {
		return x.Value.LastTok()
	}
	if x.AssignTok.IsValid() {
		return x.AssignTok.LastTok()
	}
	if x.Base != nil {
		return x.Base.LastTok()
	}
	if x.ModifiesTok.IsValid() {
		x.ModifiesTok.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	if x.TypePars != nil {
		return x.TypePars.LastTok()
	}
	if x.Name != nil {
		return x.Name.LastTok()
	}
	if x.Type != nil {
		return x.Type.LastTok()
	}
	return x.TemplateTok.LastTok()
}

func (x *ModuleParameterGroup) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *FuncDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Body != nil {
		return x.Body.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Kind.LastTok()
}

func (x *SignatureDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Exception != nil {
		return x.Exception.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.NoBlock.LastTok().IsValid() {
		return x.NoBlock.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Name.LastTok()
}

func (x *SubTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Field.LastTok()
}

func (x *StructTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *EnumTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *BehaviourTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.System != nil {
		return x.System.LastTok()
	}
	if x.RunsOn != nil {
		return x.RunsOn.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Name.LastTok()
}

func (x *PortTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *PortAttribute) LastTok() *Token {
	return x.Types[len(x.Types)-1].LastTok()
}

func (x *PortMapAttribute) LastTok() *Token {
	return x.Params.LastTok()
}

func (x *ComponentTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Body.LastTok()
}

func (x *ValueDecl) Pos() loc.Pos {
	if x.Kind.Pos().IsValid() {
		return x.Kind.Pos()
	}
	return x.Type.Pos()
}

func (x *Declarator) Pos() loc.Pos {
	if x.Name != nil {
		return x.Name.Pos()
	}
	if len(x.ArrayDef) > 0 {
		return x.ArrayDef[0].Pos()
	}
	if x.AssignTok.IsValid() {
		return x.AssignTok.Pos()
	}
	if x.Value != nil {
		return x.Value.Pos()
	}
	return loc.NoPos
}

func (x *ModuleParameterGroup) Pos() loc.Pos {
	return x.Tok.Pos()
}

func (x *FuncDecl) Pos() loc.Pos {
	if x.External.Kind == token.EXTERNAL {
		return x.External.Pos()
	}
	return x.Kind.Pos()
}

func (x *SignatureDecl) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *SubTypeDecl) Pos() loc.Pos       { return x.TypeTok.Pos() }
func (x *StructTypeDecl) Pos() loc.Pos    { return x.TypeTok.Pos() }
func (x *EnumTypeDecl) Pos() loc.Pos      { return x.TypeTok.Pos() }
func (x *BehaviourTypeDecl) Pos() loc.Pos { return x.TypeTok.Pos() }
func (x *PortTypeDecl) Pos() loc.Pos      { return x.TypeTok.Pos() }
func (x *PortAttribute) Pos() loc.Pos     { return x.Kind.Pos() }
func (x *PortMapAttribute) Pos() loc.Pos  { return x.MapTok.Pos() }
func (x *ComponentTypeDecl) Pos() loc.Pos { return x.TypeTok.Pos() }

func (x *ValueDecl) End() loc.Pos            { return x.LastTok().End() }
func (x *Declarator) End() loc.Pos           { return x.LastTok().End() }
func (x *TemplateDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *ModuleParameterGroup) End() loc.Pos { return x.LastTok().End() }
func (x *FuncDecl) End() loc.Pos             { return x.LastTok().End() }
func (x *SignatureDecl) End() loc.Pos        { return x.LastTok().End() }
func (x *SubTypeDecl) End() loc.Pos          { return x.LastTok().End() }
func (x *StructTypeDecl) End() loc.Pos       { return x.LastTok().End() }
func (x *EnumTypeDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *BehaviourTypeDecl) End() loc.Pos    { return x.LastTok().End() }
func (x *PortTypeDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *PortAttribute) End() loc.Pos        { return x.LastTok().End() }
func (x *PortMapAttribute) End() loc.Pos     { return x.LastTok().End() }
func (x *ComponentTypeDecl) End() loc.Pos    { return x.LastTok().End() }

func (x *Module) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.RBrace.IsValid() {
		return x.RBrace.LastTok()
	}
	if l := len(x.Defs); l > 0 {
		return x.Defs[l-1].LastTok()
	}
	if x.Language != nil {
		return x.Language.LastTok()
	}

	if x.Name != nil {
		return x.Name.LastTok()
	}

	return x.Tok.LastTok()
}

func (x *ModuleDef) LastTok() *Token {
	return x.Def.LastTok()
}

func (x *ControlPart) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Body.LastTok()
}

func (x *ImportDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.RBrace.IsValid() {
		return x.RBrace.LastTok()
	}
	if l := len(x.List); l != 0 {
		return x.List[l-1].LastTok()
	}
	if x.Language != nil {
		return x.Language.LastTok()
	}
	if x.LBrace.IsValid() {
		return x.LBrace.LastTok()
	}
	if x.Module != nil {
		return x.Module.LastTok()
	}
	if x.FromTok.IsValid() {
		return x.FromTok.LastTok()
	}
	return x.ImportTok.LastTok()
}

func (x *GroupDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *FriendDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Module.LastTok()
}

func (x *Module) Pos() loc.Pos { return x.Tok.Pos() }

func (x *ModuleDef) Pos() loc.Pos {
	if x.Visibility.Pos().IsValid() {
		return x.Visibility.Pos()
	}
	return x.Def.Pos()
}

func (x *ControlPart) Pos() loc.Pos { return x.Name.Pos() }
func (x *ImportDecl) Pos() loc.Pos  { return x.ImportTok.Pos() }
func (x *GroupDecl) Pos() loc.Pos   { return x.Tok.Pos() }
func (x *FriendDecl) Pos() loc.Pos  { return x.FriendTok.Pos() }

func (x *Module) End() loc.Pos      { return x.LastTok().End() }
func (x *ModuleDef) End() loc.Pos   { return x.LastTok().End() }
func (x *ControlPart) End() loc.Pos { return x.LastTok().End() }
func (x *ImportDecl) End() loc.Pos  { return x.LastTok().End() }
func (x *GroupDecl) End() loc.Pos   { return x.LastTok().End() }
func (x *FriendDecl) End() loc.Pos  { return x.LastTok().End() }

func (x *LanguageSpec) LastTok() *Token { return x.List[len(x.List)-1].LastTok() }
func (x *RestrictionSpec) LastTok() *Token {
	if x.RParen.LastTok().IsValid() {
		return x.RParen.LastTok()
	}
	return x.Tok.LastTok()
}
func (x *RunsOnSpec) LastTok() *Token { return x.Comp.LastTok() }
func (x *SystemSpec) LastTok() *Token { return x.Comp.LastTok() }
func (x *MtcSpec) LastTok() *Token    { return x.Comp.LastTok() }
func (x *ReturnSpec) LastTok() *Token { return x.Type.LastTok() }
func (x *FormalPars) LastTok() *Token { return x.RParen.LastTok() }

func (x *FormalPar) LastTok() *Token {
	if x.Value != nil {
		return x.Value.LastTok()
	}
	if x.AssignTok.IsValid() {
		return x.AssignTok.LastTok()
	}
	if l := len(x.ArrayDef); l != 0 {
		return x.ArrayDef[l-1].LastTok()
	}
	if x.Name != nil {
		return x.Name.LastTok()
	}
	if x.Type != nil {
		return x.Type.LastTok()
	}
	if x.Modif.IsValid() {
		return x.Modif.LastTok()
	}
	if x.TemplateRestriction != nil {
		return x.TemplateRestriction.LastTok()
	}
	return x.Direction.LastTok()
}

func (x *WithSpec) LastTok() *Token { return x.RBrace.LastTok() }
func (x *WithStmt) LastTok() *Token { return x.Value.LastTok() }

func (x *LanguageSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *RestrictionSpec) Pos() loc.Pos {
	if x.TemplateTok.Pos().IsValid() {
		return x.TemplateTok.Pos()
	}
	return x.Tok.Pos()
}
func (x *RunsOnSpec) Pos() loc.Pos { return x.RunsTok.Pos() }
func (x *SystemSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *MtcSpec) Pos() loc.Pos    { return x.Tok.Pos() }
func (x *ReturnSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *FormalPars) Pos() loc.Pos { return x.LParen.Pos() }

func (x *FormalPar) Pos() loc.Pos {
	if x.Direction.Pos().IsValid() {
		return x.Direction.Pos()
	}
	if x.TemplateRestriction != nil && x.TemplateRestriction.Pos().IsValid() {
		return x.TemplateRestriction.Pos()
	}
	if x.Modif.Pos().IsValid() {
		return x.Modif.Pos()
	}
	return x.Type.Pos()
}

func (x *WithSpec) Pos() loc.Pos        { return x.Tok.Pos() }
func (x *WithStmt) Pos() loc.Pos        { return x.Kind.Pos() }
func (x *LanguageSpec) End() loc.Pos    { return x.LastTok().End() }
func (x *RestrictionSpec) End() loc.Pos { return x.LastTok().End() }
func (x *RunsOnSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *SystemSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *MtcSpec) End() loc.Pos         { return x.LastTok().End() }
func (x *ReturnSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *FormalPars) End() loc.Pos      { return x.LastTok().End() }
func (x *FormalPar) End() loc.Pos       { return x.LastTok().End() }
func (x *WithSpec) End() loc.Pos        { return x.LastTok().End() }
func (x *WithStmt) End() loc.Pos        { return x.LastTok().End() }
