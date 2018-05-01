package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseExprList() (list []ast.Expr) {
	list = append(list, p.parseExpr())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseExpr())
	}
	return list
}

func (p *parser) parseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "Expr"))
	}

	x := p.parseBinaryExpr(token.LowestPrec + 1)

	if p.tok == token.ASSIGN {
		p.next()
		p.parseExpr()
	}
	return x
}

func (p *parser) parseBinaryExpr(prec1 int) ast.Expr {
	x := p.parseUnaryExpr()
	for {
		prec := p.tok.Precedence()
		if prec < prec1 {
			return x
		}
		pos := p.pos
		op := p.tok
		p.next()

		y := p.parseBinaryExpr(prec + 1)

		x = &ast.BinaryExpr{X: x, Op: op, OpPos: pos, Y: y}
	}
}

func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case token.SUB, token.ADD, token.NOT, token.NOT4B, token.EXCL:
		op, pos := p.tok, p.pos
		p.next()
		// handle unused expr '-'
		if op == token.SUB && (p.tok == token.COMMA || p.tok == token.SEMICOLON || p.tok == token.RBRACE || p.tok == token.RBRACK || p.tok == token.EOF) {
			return nil
		}
		return &ast.UnaryExpr{Op: op, OpPos: pos, X: p.parseUnaryExpr()}
	}
	return p.parsePrimaryExpr()
}

func (p *parser) parsePrimaryExpr() ast.Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok {
		case token.DOT:
			x = p.parseSelectorExpr(x)
		case token.LBRACK:
			x = p.parseIndexExpr(x)
		case token.LPAREN:
			x = p.parseCallExpr(x)
		default:
			break L
		}
	}

	if p.tok == token.TO || p.tok == token.FROM {
		p.next()
		p.parseExpr()
	}

	if p.tok == token.REDIRECT {
		p.parseRedirect()
	}

	return x
}

func (p *parser) parseOperand() ast.Expr {
	switch p.tok {
	case token.IDENT, token.TIMER, token.TESTCASE:
		id := &ast.Ident{NamePos: p.pos, Name: p.lit}
		p.next()
		return id
	case token.LBRACE:
		p.next()
		p.parseExprList()
		p.expect(token.RBRACE)
		return nil
	case token.LPAREN:
		p.next()
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		set := &ast.SetExpr{List: p.parseExprList()}
		p.expect(token.RPAREN)
		return set
	case token.INT, token.FLOAT, token.STRING, token.BSTRING,
		token.ANY, token.MUL:
		lit := &ast.ValueLiteral{Kind: p.tok, ValuePos: p.pos, Value: p.lit}
		p.next()
		return lit
	}

	p.errorExpected(p.pos, "operand")
	return nil
}

func (p *parser) parseSelectorExpr(x ast.Expr) ast.Expr {
	p.next()
	return &ast.SelectorExpr{X: x, Sel: p.parseIdent()}
}

func (p *parser) parseIndexExpr(x ast.Expr) ast.Expr {
	p.next()
	x = &ast.IndexExpr{X: x, Index: p.parseExpr()}
	p.expect(token.RBRACK)
	return x
}

func (p *parser) parseCallExpr(x ast.Expr) ast.Expr {
	p.next()

	var list []ast.Expr
	if p.tok != token.RPAREN {
		list = p.parseExprList()
	}
	p.expect(token.RPAREN)
	return &ast.CallExpr{Fun: x, Args: list}
}

func (p *parser) parseRedirect() ast.Expr {
	p.next()

	if p.tok == token.VALUE {
		p.next()
		p.parseExprList()
	}

	if p.tok == token.PARAM {
		p.next()
		p.parseExprList()
	}

	if p.tok == token.SENDER {
		p.next()
		p.parsePrimaryExpr()
	}

	if p.tok == token.MODIF {
		p.next()
		if p.tok == token.VALUE {
			p.next()
		}
		p.parsePrimaryExpr()
	}

	if p.tok == token.TIMESTAMP {
		p.next()
		p.parsePrimaryExpr()
	}

	return nil
}
