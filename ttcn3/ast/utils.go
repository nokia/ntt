package ast

import (
	"reflect"
	"strings"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/token"
)

// IsNil returns true if the node is nil.
func IsNil(n Node) bool {
	if n == nil {
		return true
	}
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	return false
}

// FindChildOfType returns the first direct child of the give node, enclosing
// given position.
func FindChildOf(n Node, pos loc.Pos) Node {
	if IsNil(n) || !pos.IsValid() {
		return nil
	}
	children := Children(n)
	for _, c := range children {
		if IsNil(c) {
			continue
		}

		// ErrorNodes may overlap with other nodes and mess up the search.
		if _, ok := c.(*ErrorNode); ok {
			continue
		}

		if c.Pos() <= pos && pos < c.End() {
			return c
		}
	}
	return nil
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
		if n.Tok.Kind() == token.LABEL {
			return Name(n.Label)
		}
	case *ControlPart:
		return Name(n.Name)
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
	case *SubTypeDecl:
		if n.Field != nil {
			return Name(n.Field)
		}
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
	case *ModuleDef:
		return Name(n.Def)

	}
	return ""
}

func add(s []Node, n Node) []Node {
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() || n == nil {
		return s
	}
	return append(s, n)
}

func Children(n Node) []Node {
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() || n == nil {
		return nil
	}
	var children []Node
	switch n := n.(type) {
	case *NodeList:
		for _, child := range n.Nodes {
			children = add(children, child)
		}
	case Token:
	case *ErrorNode:
		children = add(children, n.From)
		children = add(children, n.To)

	case *Ident:
		children = add(children, n.Tok)
		children = add(children, n.Tok2)

	case *ParametrizedIdent:
		children = add(children, n.Ident)
		children = add(children, n.Params)

	case *ValueLiteral:
		children = add(children, n.Tok)

	case *CompositeLiteral:
		children = add(children, n.LBrace)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *UnaryExpr:
		children = add(children, n.Op)
		children = add(children, n.X)

	case *BinaryExpr:
		children = add(children, n.X)
		children = add(children, n.Op)
		children = add(children, n.Y)

	case *ParenExpr:
		children = add(children, n.LParen)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RParen)

	case *SelectorExpr:
		children = add(children, n.X)
		children = add(children, n.Dot)
		children = add(children, n.Sel)

	case *IndexExpr:
		children = add(children, n.X)
		children = add(children, n.LBrack)
		children = add(children, n.Index)
		children = add(children, n.RBrack)

	case *CallExpr:
		children = add(children, n.Fun)
		children = add(children, n.Args)

	case *LengthExpr:
		children = add(children, n.X)
		children = add(children, n.Len)
		children = add(children, n.Size)

	case *RedirectExpr:
		children = add(children, n.X)
		children = add(children, n.Tok)
		children = add(children, n.ValueTok)
		for _, n := range n.Value {
			children = add(children, n)
		}
		children = add(children, n.ParamTok)
		for _, n := range n.Param {
			children = add(children, n)
		}
		children = add(children, n.SenderTok)
		children = add(children, n.Sender)
		children = add(children, n.IndexTok)
		children = add(children, n.IndexValueTok)
		children = add(children, n.Index)
		children = add(children, n.TimestampTok)
		children = add(children, n.Timestamp)

	case *ValueExpr:
		children = add(children, n.X)
		children = add(children, n.Tok)
		children = add(children, n.Y)

	case *ParamExpr:
		children = add(children, n.X)
		children = add(children, n.Tok)
		children = add(children, n.Y)

	case *FromExpr:
		children = add(children, n.Kind)
		children = add(children, n.FromTok)
		children = add(children, n.X)

	case *ModifiesExpr:
		children = add(children, n.Tok)
		children = add(children, n.X)
		children = add(children, n.Assign)
		children = add(children, n.Y)

	case *RegexpExpr:
		children = add(children, n.Tok)
		children = add(children, n.NoCase)
		children = add(children, n.X)

	case *PatternExpr:
		children = add(children, n.Tok)
		children = add(children, n.NoCase)
		children = add(children, n.X)

	case *DecmatchExpr:
		children = add(children, n.Tok)
		children = add(children, n.Params)
		children = add(children, n.X)

	case *DecodedExpr:
		children = add(children, n.Tok)
		children = add(children, n.Params)
		children = add(children, n.X)

	case *DefKindExpr:
		children = add(children, n.Kind)
		for _, n := range n.List {
			children = add(children, n)
		}

	case *ExceptExpr:
		children = add(children, n.X)
		children = add(children, n.ExceptTok)
		children = add(children, n.LBrace)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *BlockStmt:
		children = add(children, n.LBrace)
		for _, n := range n.Stmts {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *DeclStmt:
		children = add(children, n.Decl)

	case *ExprStmt:
		children = add(children, n.Expr)

	case *BranchStmt:
		children = add(children, n.Tok)
		children = add(children, n.Label)

	case *ReturnStmt:
		children = add(children, n.Tok)
		children = add(children, n.Result)

	case *AltStmt:
		children = add(children, n.Tok)
		children = add(children, n.NoDefault)
		children = add(children, n.Body)

	case *CallStmt:
		children = add(children, n.Stmt)
		children = add(children, n.Body)

	case *ForStmt:
		children = add(children, n.Tok)
		children = add(children, n.LParen)
		children = add(children, n.Init)
		children = add(children, n.InitSemi)
		children = add(children, n.Cond)
		children = add(children, n.CondSemi)
		children = add(children, n.Post)
		children = add(children, n.RParen)
		children = add(children, n.Body)

	case *WhileStmt:
		children = add(children, n.Tok)
		children = add(children, n.Cond)
		children = add(children, n.Body)

	case *DoWhileStmt:
		children = add(children, n.DoTok)
		children = add(children, n.Body)
		children = add(children, n.WhileTok)
		children = add(children, n.Cond)

	case *IfStmt:
		children = add(children, n.Tok)
		children = add(children, n.Cond)
		children = add(children, n.Then)
		children = add(children, n.ElseTok)
		children = add(children, n.Else)

	case *SelectStmt:
		children = add(children, n.Tok)
		children = add(children, n.Union)
		children = add(children, n.Tag)
		children = add(children, n.LBrace)
		for _, n := range n.Body {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *CaseClause:
		children = add(children, n.Tok)
		children = add(children, n.Case)
		children = add(children, n.Body)

	case *CommClause:
		children = add(children, n.LBrack)
		children = add(children, n.X)
		children = add(children, n.Else)
		children = add(children, n.RBrack)
		children = add(children, n.Comm)
		children = add(children, n.Body)

	case *Field:
		children = add(children, n.DefaultTok)
		children = add(children, n.Type)
		children = add(children, n.Name)
		for _, n := range n.ArrayDef {
			children = add(children, n)
		}
		children = add(children, n.TypePars)
		children = add(children, n.ValueConstraint)
		children = add(children, n.LengthConstraint)
		children = add(children, n.Optional)

	case *RefSpec:
		children = add(children, n.X)

	case *StructSpec:
		children = add(children, n.Kind)
		children = add(children, n.LBrace)
		for _, n := range n.Fields {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *ListSpec:
		children = add(children, n.Kind)
		children = add(children, n.Length)
		children = add(children, n.OfTok)
		children = add(children, n.ElemType)

	case *EnumSpec:
		children = add(children, n.Tok)
		children = add(children, n.LBrace)
		for _, n := range n.Enums {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *BehaviourSpec:
		children = add(children, n.Kind)
		children = add(children, n.Params)
		children = add(children, n.RunsOn)
		children = add(children, n.System)
		children = add(children, n.Return)

	case *ValueDecl:
		children = add(children, n.Kind)
		children = add(children, n.TemplateRestriction)
		children = add(children, n.Modif)
		children = add(children, n.Type)
		for _, n := range n.Decls {
			children = add(children, n)
		}
		children = add(children, n.With)

	case *Declarator:
		children = add(children, n.Name)
		for _, n := range n.ArrayDef {
			children = add(children, n)
		}
		children = add(children, n.AssignTok)
		children = add(children, n.Value)

	case *TemplateDecl:
		children = add(children, n.RestrictionSpec)
		children = add(children, n.Modif)
		children = add(children, n.Type)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.Params)
		children = add(children, n.ModifiesTok)
		children = add(children, n.Base)
		children = add(children, n.AssignTok)
		children = add(children, n.Value)
		children = add(children, n.With)

	case *ModuleParameterGroup:
		children = add(children, n.Tok)
		children = add(children, n.LBrace)
		for _, n := range n.Decls {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *FuncDecl:
		children = add(children, n.External)
		children = add(children, n.Kind)
		children = add(children, n.Name)
		children = add(children, n.Modif)
		children = add(children, n.TypePars)
		children = add(children, n.Params)
		children = add(children, n.RunsOn)
		children = add(children, n.Mtc)
		children = add(children, n.System)
		children = add(children, n.Return)
		children = add(children, n.Body)
		children = add(children, n.With)

	case *SignatureDecl:
		children = add(children, n.Tok)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.Params)
		children = add(children, n.NoBlock)
		children = add(children, n.Return)
		children = add(children, n.ExceptionTok)
		children = add(children, n.Exception)
		children = add(children, n.With)

	case *SubTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.Field)
		children = add(children, n.With)

	case *StructTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.Kind)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.LBrace)
		for _, n := range n.Fields {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *EnumTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.EnumTok)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.LBrace)
		for _, n := range n.Enums {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *BehaviourTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.Kind)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.Params)
		children = add(children, n.RunsOn)
		children = add(children, n.System)
		children = add(children, n.Return)
		children = add(children, n.With)

	case *PortTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.PortTok)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.Kind)
		children = add(children, n.Realtime)
		children = add(children, n.LBrace)
		for _, n := range n.Attrs {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *PortAttribute:
		children = add(children, n.Kind)
		for _, n := range n.Types {
			children = add(children, n)
		}

	case *PortMapAttribute:
		children = add(children, n.MapTok)
		children = add(children, n.ParamTok)
		children = add(children, n.Params)

	case *ComponentTypeDecl:
		children = add(children, n.TypeTok)
		children = add(children, n.CompTok)
		children = add(children, n.Name)
		children = add(children, n.TypePars)
		children = add(children, n.ExtendsTok)
		for _, n := range n.Extends {
			children = add(children, n)
		}
		children = add(children, n.Body)
		children = add(children, n.With)

	case *Module:
		children = add(children, n.Tok)
		children = add(children, n.Name)
		children = add(children, n.Language)
		children = add(children, n.LBrace)
		for _, n := range n.Defs {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *ModuleDef:
		children = add(children, n.Visibility)
		children = add(children, n.Def)

	case *ControlPart:
		children = add(children, n.Name)
		children = add(children, n.Body)
		children = add(children, n.With)

	case *ImportDecl:
		children = add(children, n.ImportTok)
		children = add(children, n.FromTok)
		children = add(children, n.Module)
		children = add(children, n.Language)
		children = add(children, n.LBrace)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *GroupDecl:
		children = add(children, n.Tok)
		children = add(children, n.Name)
		children = add(children, n.LBrace)
		for _, n := range n.Defs {
			children = add(children, n)
		}
		children = add(children, n.RBrace)
		children = add(children, n.With)

	case *FriendDecl:
		children = add(children, n.FriendTok)
		children = add(children, n.ModuleTok)
		children = add(children, n.Module)
		children = add(children, n.With)

	case *LanguageSpec:
		children = add(children, n.Tok)
		for _, n := range n.List {
			children = add(children, n)
		}

	case *RestrictionSpec:
		children = add(children, n.TemplateTok)
		children = add(children, n.LParen)
		children = add(children, n.Tok)
		children = add(children, n.RParen)

	case *RunsOnSpec:
		children = add(children, n.RunsTok)
		children = add(children, n.OnTok)
		children = add(children, n.Comp)

	case *SystemSpec:
		children = add(children, n.Tok)
		children = add(children, n.Comp)

	case *MtcSpec:
		children = add(children, n.Tok)
		children = add(children, n.Comp)

	case *ReturnSpec:
		children = add(children, n.Tok)
		children = add(children, n.Restriction)
		children = add(children, n.Modif)
		children = add(children, n.Type)

	case *FormalPars:
		children = add(children, n.LParen)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RParen)

	case *FormalPar:
		children = add(children, n.Direction)
		children = add(children, n.TemplateRestriction)
		children = add(children, n.Modif)
		children = add(children, n.Type)
		children = add(children, n.Name)
		for _, n := range n.ArrayDef {
			children = add(children, n)
		}
		children = add(children, n.AssignTok)
		children = add(children, n.Value)

	case *WithSpec:
		children = add(children, n.Tok)
		children = add(children, n.LBrace)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RBrace)

	case *WithStmt:
		children = add(children, n.Kind)
		children = add(children, n.Override)
		children = add(children, n.LParen)
		for _, n := range n.List {
			children = add(children, n)
		}
		children = add(children, n.RParen)
		children = add(children, n.Value)
	}
	return children
}

type span struct {
	Begin, End loc.Position
}

func newSpan(fset *loc.FileSet, n Node) span {
	return span{
		Begin: fset.Position(n.Pos()),
		End:   fset.Position(n.End()),
	}
}

// Doc returns the documentation string for the given node.
func Doc(fset *loc.FileSet, n Node) string {
	if n == nil {
		return ""
	}

	tok := n.FirstTok()
	if tok == nil {
		return ""
	}

	var ret string
	prev := newSpan(fset, tok)
L:
	for {
		tok = tok.PrevTok()
		if tok == nil {
			break
		}

		switch tok.Kind() {
		case token.COMMENT:
			curr := newSpan(fset, tok)
			dist := prev.Begin.Line - curr.End.Line
			if dist > 1 {
				break L
			}
			prev = curr
			text := tok.String()
			switch text[1] {
			case '/':
				text = text[2:]
				if len(text) > 0 && text[0] == ' ' {
					text = text[1:]
				}
				ret = text + "\n" + ret
			case '*':
				text = text[2 : len(text)-2]
				lines := strings.Split(text, "\n")
				for i, line := range lines {
					if len(line) > 0 && line[0] == ' ' {
						line = line[1:]
					}
					line := strings.TrimRight(line, " ")
					lines[i] = line
				}
				text = strings.Join(lines, "\n")
				if dist > 0 {
					text = text + "\n"
				} else {
					text = text + " "
				}
				ret = text + ret
			}

		case token.EXTERNAL, token.PRIVATE, token.PUBLIC, token.FRIEND:
			// Modifiers might not necessarily be part
			// of the Node and are just skipped over.
		default:
			break L
		}
	}
	return ret
}
