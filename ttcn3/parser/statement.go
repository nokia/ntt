package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	x := &ast.BlockStmt{LBrace: p.pos}
	p.expect(token.LBRACE)
	p.openScope()
	for p.tok != token.EOF && p.tok != token.RBRACE {
		x.Stmts = append(x.Stmts, p.parseStmt())
	}
	p.closeScope()
	p.expect(token.RBRACE)
	return x
}

func (p *parser) parseSimpleStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	p.parseExpr()
	return nil
}

func (p *parser) parseStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok {
	case token.VAR, token.CONST:
		p.parseDecl()
	case token.REPEAT, token.BREAK, token.CONTINUE:
		p.next()
	case token.LABEL:
		p.next()
		p.expect(token.IDENT)
	case token.GOTO:
		p.next()
		p.expect(token.IDENT)
	case token.RETURN:
		p.next()
		if p.tok != token.SEMICOLON && p.tok != token.RBRACE {
			p.parseExpr()
		}
	case token.SELECT:
		p.parseSelect()
	case token.ALT, token.INTERLEAVE:
		p.next()
		p.parseBlockStmt()
	case token.LBRACK:
		p.parseAltGuard()
	case token.FOR:
		p.parseForLoop()
	case token.WHILE:
		p.parseWhileLoop()
	case token.DO:
		p.parseDoWhileLoop()
	case token.IF:
		p.parseIfStmt()
	default:
		p.parseSimpleStmt()
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseForLoop() {
	p.next()
	p.expect(token.LPAREN)
	if p.tok == token.VAR {
		p.parseValueDecl()
	} else {
		p.parseExpr()
	}
	p.expect(token.SEMICOLON)
	p.parseExpr()
	p.expect(token.SEMICOLON)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
}

func (p *parser) parseWhileLoop() {
	p.next()
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
}

func (p *parser) parseDoWhileLoop() {
	p.next()
	p.parseBlockStmt()
	p.expect(token.WHILE)
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
}

func (p *parser) parseIfStmt() {
	p.next()
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
	if p.tok == token.ELSE {
		p.next()
		if p.tok == token.IF {
			p.parseIfStmt()
		} else {
			p.parseBlockStmt()
		}
	}
}

func (p *parser) parseSelect() {
	p.next()
	if p.tok == token.UNION {
		p.next()
	}

	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)
	for p.tok == token.CASE {
		p.parseCaseStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseCaseStmt() {
	p.expect(token.CASE)
	if p.tok == token.ELSE {
		p.next()
	} else {
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}
	p.parseBlockStmt()
}

func (p *parser) parseAltGuard() {
	p.next()
	if p.tok == token.ELSE {
		p.next()
		p.expect(token.RBRACK)
		p.parseBlockStmt()
		return
	}

	if p.tok != token.RBRACK {
		p.parseExpr()
	}
	p.expect(token.RBRACK)
	p.parseSimpleStmt()
	if p.tok == token.LBRACE {
		p.parseBlockStmt()
	}
}
