package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	x := &ast.BlockStmt{LBrace: p.pos}
	p.expect(token.LBRACE)
	p.openScope()
	for p.tok != token.RBRACE {
		x.Stmts = append(x.Stmts, p.parseStmt())
	}
	p.closeScope()
	p.expect(token.RBRACE)
	return x
}

func (p *parser) parseStmt() ast.Stmt {
	switch p.tok {
	case token.IDENT:
		return p.parseSimpleStmt()
	case token.VAR, token.CONST:
		return p.parseDeclStmt()
	default:
		p.errorExpected(p.pos, "statement")
		return nil
	}
}

func (p *parser) parseDeclStmt() *ast.DeclStmt {
	x := &ast.DeclStmt{Decl: p.parseDecl()}
	p.expect(token.SEMICOLON)
	return x
}

func (p *parser) parseSimpleStmt() *ast.ExprStmt {
	x := &ast.ExprStmt{Expr: p.parseExpr()}
	p.expect(token.SEMICOLON)
	return x
}
