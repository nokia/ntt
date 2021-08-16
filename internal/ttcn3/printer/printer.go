package printer

import (
	"fmt"
	"io"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type whiteSpace byte

const (
	ignore   = whiteSpace(0)
	blank    = whiteSpace(' ')
	vtab     = whiteSpace('\v')
	newline  = whiteSpace('\n')
	formfeed = whiteSpace('\f')
	indent   = whiteSpace('>')
	unindent = whiteSpace('<')
)

func Print(w io.Writer, fset *loc.FileSet, n ast.Node) error {
	p := printer{w: w, fset: fset}
	p.print(n)
	return p.err
}

type printer struct {
	w      io.Writer
	fset   *loc.FileSet
	indent int
	err    error
}

var lastPrintNewLine = false

func (p *printer) print(values ...interface{}) {

	for _, v := range values {
		switch n := v.(type) {
		case whiteSpace:
			switch n {
			case indent:
				p.indent++
			case unindent:
				p.indent--
			default:
				fmt.Fprint(p.w, n)
			}

		case *ast.ErrorNode:
			if n == nil {
				return
			}

			p.print(n.From)
			p.print(n.To)

		case *ast.Ident:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Tok2)

		case *ast.ParametrizedIdent:
			if n == nil {
				return
			}
			p.print(n.Ident)
			p.print(n.Params)

		case *ast.ValueLiteral:
			if n == nil {
				return
			}
			p.print(n.Tok)

		case *ast.CompositeLiteral:
			if n == nil {
				return
			}
			p.print(n.LBrace)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.UnaryExpr:
			if n == nil {
				return
			}
			p.print(n.Op)
			p.print(n.X)

		case *ast.BinaryExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Op)
			p.print(n.Y)

		case *ast.ParenExpr:
			if n == nil {
				return
			}
			p.print(n.LParen)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RParen)

		case *ast.SelectorExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Dot)
			p.print(n.Sel)

		case *ast.IndexExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.LBrack)
			p.print(n.Index)
			p.print(n.RBrack)

		case *ast.CallExpr:
			if n == nil {
				return
			}
			p.print(n.Fun)
			p.print(n.Args)

		case *ast.LengthExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Len)
			p.print(n.Size)

		case *ast.RedirectExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Tok)
			p.print(n.ValueTok)
			for i := range n.Value {
				p.print(n.Value[i], ", ")
			}
			p.print(n.ParamTok)
			for i := range n.Param {
				p.print(n.Param[i], ", ")
			}
			p.print(n.SenderTok)
			p.print(n.Sender)
			p.print(n.IndexTok)
			p.print(n.IndexValueTok)
			p.print(n.Index)
			p.print(n.TimestampTok)
			p.print(n.Timestamp)

		case *ast.ValueExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Tok)
			p.print(n.Y)

		case *ast.ParamExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.Tok)
			p.print(n.Y)

		case *ast.FromExpr:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.FromTok)
			p.print(n.X)

		case *ast.ModifiesExpr:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.X)
			p.print(n.Assign)
			p.print(n.Y)

		case *ast.RegexpExpr:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.NoCase)
			p.print(n.X)

		case *ast.PatternExpr:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.NoCase)
			p.print(n.X)

		case *ast.DecmatchExpr:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Params)
			p.print(n.X)

		case *ast.DecodedExpr:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Params)
			p.print(n.X)

		case *ast.DefKindExpr:
			if n == nil {
				return
			}
			p.print(n.Kind)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}

		case *ast.ExceptExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.ExceptTok)
			p.print(n.LBrace)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.BlockStmt:
			if n == nil {
				return
			}
			p.print(n.LBrace)
			for i := range n.Stmts {
				p.print(n.Stmts[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.DeclStmt:
			if n == nil {
				return
			}
			p.print(n.Decl)

		case *ast.ExprStmt:
			if n == nil {
				return
			}
			p.print(n.Expr)

		case *ast.BranchStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Label)

		case *ast.ReturnStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Result)

		case *ast.AltStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Body)

		case *ast.CallStmt:
			if n == nil {
				return
			}
			p.print(n.Stmt)
			p.print(n.Body)

		case *ast.ForStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.LParen)
			p.print(n.Init)
			p.print(n.InitSemi)
			p.print(n.Cond)
			p.print(n.CondSemi)
			p.print(n.Post)
			p.print(n.RParen)
			p.print(n.Body)

		case *ast.WhileStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Cond)
			p.print(n.Body)

		case *ast.DoWhileStmt:
			if n == nil {
				return
			}
			p.print(n.DoTok)
			p.print(n.Body)
			p.print(n.WhileTok)
			p.print(n.Cond)

		case *ast.IfStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Cond)
			p.print(n.Then)
			p.print(n.ElseTok)
			p.print(n.Else)

		case *ast.SelectStmt:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Union)
			p.print(n.Tag)
			p.print(n.LBrace)
			for i := range n.Body {
				p.print(n.Body[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.CaseClause:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Case)
			p.print(n.Body)

		case *ast.CommClause:
			if n == nil {
				return
			}
			p.print(n.LBrack)
			p.print(n.X)
			p.print(n.Else)
			p.print(n.RBrack)
			p.print(n.Comm)
			p.print(n.Body)

		case *ast.Field:
			if n == nil {
				return
			}
			p.print(n.DefaultTok)
			p.print(n.Type)
			p.print(n.Name)
			for i := range n.ArrayDef {
				p.print(n.ArrayDef[i], ", ")
			}
			p.print(n.TypePars)
			p.print(n.ValueConstraint)
			p.print(n.LengthConstraint)
			p.print(n.Optional)

		case *ast.RefSpec:
			if n == nil {
				return
			}
			p.print(n.X)

		case *ast.StructSpec:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.LBrace)
			for i := range n.Fields {
				p.print(n.Fields[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.ListSpec:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.Length)
			p.print(n.OfTok)
			p.print(n.ElemType)

		case *ast.EnumSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.LBrace)
			for i := range n.Enums {
				p.print(n.Enums[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.BehaviourSpec:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.Params)
			p.print(n.Return)
			p.print(n.System)
			p.print(n.Return)

		case *ast.ValueDecl:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.TemplateRestriction)
			p.print(n.Modif)
			p.print(n.Type)
			for i := range n.Decls {
				p.print(n.Decls[i], ", ")
			}
			p.print(n.With)

		case *ast.Declarator:
			if n == nil {
				return
			}
			p.print(n.Name)
			for i := range n.ArrayDef {
				p.print(n.ArrayDef[i], ", ")
			}
			p.print(n.AssignTok)
			p.print(n.Value)

		case *ast.TemplateDecl:
			if n == nil {
				return
			}
			p.print(&n.RestrictionSpec)
			p.print(n.Modif)
			p.print(n.Type)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.Params)
			p.print(n.ModifiesTok)
			p.print(n.Base)
			p.print(n.AssignTok)
			p.print(n.Value)
			p.print(n.With)

		case *ast.ModuleParameterGroup:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.LBrace)
			for i := range n.Decls {
				p.print(n.Decls[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.FuncDecl:
			if n == nil {
				return
			}
			p.print(n.External)
			p.print(n.Kind)
			p.print(n.Name)
			p.print(n.Modif)
			p.print(n.TypePars)
			p.print(n.Params)
			p.print(n.RunsOn)
			p.print(n.Mtc)
			p.print(n.System)
			p.print(n.Return)
			p.print(n.Body)
			p.print(n.With)

		case *ast.SignatureDecl:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.Params)
			p.print(n.NoBlock)
			p.print(n.Return)
			p.print(n.ExceptionTok)
			p.print(n.Exception)
			p.print(n.With)

		case *ast.SubTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.Field)
			p.print(n.With)

		case *ast.StructTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.Kind)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.LBrace)
			for i := range n.Fields {
				p.print(n.Fields[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.EnumTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.EnumTok)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.LBrace)
			for i := range n.Enums {
				p.print(n.Enums[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.BehaviourTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.Kind)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.Params)
			p.print(n.RunsOn)
			p.print(n.System)
			p.print(n.Return)
			p.print(n.With)

		case *ast.PortTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.PortTok)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.Kind)
			p.print(n.Realtime)
			p.print(n.LBrace)
			for i := range n.Attrs {
				p.print(n.Attrs[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.PortAttribute:
			if n == nil {
				return
			}
			p.print(n.Kind)
			for i := range n.Types {
				p.print(n.Types[i], ", ")
			}

		case *ast.PortMapAttribute:
			if n == nil {
				return
			}
			p.print(n.MapTok)
			p.print(n.ParamTok)
			p.print(n.Params)

		case *ast.ComponentTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.CompTok)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print(n.ExtendsTok)
			for i := range n.Extends {
				p.print(n.Extends[i], ", ")
			}
			p.print(n.Body)
			p.print(n.With)

		case *ast.Module:
			if n == nil {
				return
			}

			p.print(n.Tok, " ")
			p.print(n.Name, " ")
			p.print(n.Language)
			p.print(n.LBrace)
			for i := range n.Defs {
				p.print(n.Defs[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.ModuleDef:
			if n == nil {
				return
			}
			p.print(n.Def)
			p.print(n.Visibility)

		case *ast.ControlPart:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Body)
			p.print(n.With)

		case *ast.ImportDecl:
			if n == nil {
				return
			}
			p.print(n.ImportTok)
			p.print(n.FromTok)
			p.print(n.Module)
			p.print(n.Language)
			p.print(n.LBrace)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.GroupDecl:
			if n == nil {
				return
			}

			p.print(n.Tok, " ")
			p.print(n.Name, " ")
			p.print(n.LBrace)
			for i := range n.Defs {
				p.print(n.Defs[i], ", ")
			}
			p.print(n.RBrace)
			p.print(n.With)

		case *ast.FriendDecl:
			if n == nil {
				return
			}
			p.print(n.FriendTok)
			p.print(n.ModuleTok)
			p.print(n.Module)
			p.print(n.With)

		case *ast.LanguageSpec:
			if n == nil {
				return
			}
			p.print(n.Tok, " ")
			for i := range n.List {
				p.print(n.List[i], ", ")
			}

		case *ast.RestrictionSpec:
			if n == nil {
				return
			}
			p.print(n.TemplateTok)
			p.print(n.LParen)
			p.print(n.Tok)
			p.print(n.RParen)

		case *ast.RunsOnSpec:
			if n == nil {
				return
			}
			p.print(n.RunsTok)
			p.print(n.OnTok)
			p.print(n.Comp)

		case *ast.SystemSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Comp)

		case *ast.MtcSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Comp)

		case *ast.ReturnSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Restriction)
			p.print(n.Modif)
			p.print(n.Type)

		case *ast.FormalPars:
			if n == nil {
				return
			}
			p.print(n.LParen)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RParen)

		case *ast.FormalPar:
			if n == nil {
				return
			}
			p.print(n.Direction)
			p.print(n.TemplateRestriction)
			p.print(n.Modif)
			p.print(n.Type)
			p.print(n.Name)
			for i := range n.ArrayDef {
				p.print(n.ArrayDef[i], ", ")
			}
			p.print(n.AssignTok)
			p.print(n.Value)

		case *ast.WithSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.LBrace)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RBrace)

		case *ast.WithStmt:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.Override)
			p.print(n.LParen)
			for i := range n.List {
				p.print(n.List[i], ", ")
			}
			p.print(n.RParen)
			p.print(n.Value)

		case ast.Token:
			if n.IsValid() {
				if n.Pos() != 1 {
					if n.String() == "{" || n.String() == "}" { // if Token is "{" or "}"
						p.print("\n")
						if n.String() == "}" {
							p.print(unindent)
						}
					} else {
						p.print(" ")
					}
				}
				p.print(n.String())
				if n.String() == "{" { // if Token is "{"
					p.print("\n")
					p.print(indent)
				}
				if n.String() == "}" { // if Token is "}"
					p.print("\n")
				}
			}

		default:

			if lastPrintNewLine {
				for i := 0; i < p.indent; i++ {
					fmt.Print("	")
				}
				lastPrintNewLine = false
			}
			fmt.Fprint(p.w, v)
			if n == "\n" {
				lastPrintNewLine = true
			}
		}
	}
}
