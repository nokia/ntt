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
	p := printer{w: w, fset: fset, lineStart: true}
	p.print(n)
	return p.err
}

type printer struct {
	w         io.Writer
	fset      *loc.FileSet
	indent    int
	lineStart bool
	err       error
}

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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.List)
			p.print("\n", unindent, n.RBrace, "\n")

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
			p.print(n.List)
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
			p.print(n.Value)
			p.print(n.ParamTok)
			p.print(n.Param)
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
			p.print(n.List)

		case *ast.ExceptExpr:
			if n == nil {
				return
			}
			p.print(n.X)
			p.print(n.ExceptTok)
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.List)
			p.print("\n", unindent, n.RBrace, "\n")

		case *ast.BlockStmt:
			if n == nil {
				return
			}
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Stmts)
			p.print("\n", unindent, n.RBrace, "\n")

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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Body)
			p.print("\n", unindent, n.RBrace, "\n")

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
			p.print(n.ArrayDef)
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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Fields)
			p.print("\n", unindent, n.RBrace, "\n")

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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Enums)
			p.print("\n", unindent, n.RBrace, "\n")

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
			p.print(n.Decls)
			p.print(n.With)

		case *ast.Declarator:
			if n == nil {
				return
			}
			p.print(n.Name)
			p.print(n.ArrayDef)
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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Decls)
			p.print("\n", unindent, n.RBrace, "\n")
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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Fields)
			p.print("\n", unindent, n.RBrace, "\n")
			p.print(n.With)

		case *ast.EnumTypeDecl:
			if n == nil {
				return
			}
			p.print(n.TypeTok)
			p.print(n.EnumTok)
			p.print(n.Name)
			p.print(n.TypePars)
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Enums)
			p.print("\n", unindent, n.RBrace, "\n")
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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Attrs)
			p.print("\n", unindent, n.RBrace, "\n")
			p.print(n.With)

		case *ast.PortAttribute:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.Types)

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
			p.print(n.Extends)
			p.print(n.Body)
			p.print(n.With)

		case *ast.Module:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print(n.Name)
			p.print(n.Language)
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Defs)
			p.print("\n", unindent, n.RBrace, "\n")
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
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.List)
			p.print("\n", unindent, n.RBrace, "\n")
			p.print(n.With)

		case *ast.GroupDecl:
			if n == nil {
				return
			}

			p.print(n.Tok, " ")
			p.print(n.Name, " ")
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.Defs)
			p.print("\n", unindent, n.RBrace, "\n")
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
			p.print(n.List)

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
			p.print(n.List)
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
			p.print(n.ArrayDef)
			p.print(n.AssignTok)
			p.print(n.Value)

		case *ast.WithSpec:
			if n == nil {
				return
			}
			p.print(n.Tok)
			p.print("\n", n.LBrace, "\n", indent)
			p.print(n.List)
			p.print("\n", unindent, n.RBrace, "\n")

		case *ast.WithStmt:
			if n == nil {
				return
			}
			p.print(n.Kind)
			p.print(n.Override)
			p.print(n.LParen)
			p.print(n.List)
			p.print(n.RParen)
			p.print(n.Value)

		case []*ast.CaseClause:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.DefKindExpr:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.Field:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.FormalPar:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.ModuleDef:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.ParenExpr:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.ValueDecl:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []*ast.WithStmt:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []ast.Decl:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []ast.Expr:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []ast.Node:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []ast.Stmt:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}
		case []ast.Token:
			sep := ","
			for i, item := range n {
				p.print(item)
				if sep != "," || i < len(n)-1 {
					p.print(sep)
				}
			}

		case ast.Token:
			if n.IsValid() {
				p.print(n.String())
			}

		default:
			switch {
			case p.lineStart && n != "\n":
				for i := 0; i < p.indent; i++ {
					fmt.Fprint(p.w, "	")
				}
				p.lineStart = false
			case n == "(", n == ")", n == ",", n == "\n", n == ";":

			default:
				fmt.Fprint(p.w, " ")
			}
			fmt.Fprint(p.w, n)
			if n == "\n" {
				p.lineStart = true
			}
		}
	}
}
