package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseWithStmt() ast.Node {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}
	switch p.tok {
	case token.ENCODE,
		token.VARIANT,
		token.DISPLAY,
		token.EXTENSION,
		token.OPTIONAL,
		token.STEPSIZE,
		token.OVERRIDE:
		p.next()
	default:
		p.errorExpected(p.pos, "with-attribute")
		p.next()
	}

	switch p.tok {
	case token.OVERRIDE:
		p.next()
	case token.MODIF:
		p.next() // consume '@local'
	}

	if p.tok == token.LPAREN {
		p.next()
		for {
			p.parseWithQualifier()
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
		p.expect(token.RPAREN)
	}

	p.expect(token.STRING)

	if p.tok == token.DOT {
		p.next()
		p.expect(token.STRING)
	}

	p.expectSemi()
	return nil
}

func (p *parser) parseWithQualifier() {
	switch p.tok {
	case token.IDENT:
		p.parseTypeRef()
	case token.LBRACK:
		p.parseIndexExpr(nil)
	case token.TYPE, token.TEMPLATE, token.CONST, token.ALTSTEP, token.TESTCASE, token.FUNCTION, token.SIGNATURE, token.MODULEPAR, token.GROUP:
		p.next()
		p.expect(token.ALL)
		if p.tok == token.EXCEPT {
			p.next()
			p.expect(token.LBRACE)
			p.parseRefList()
			p.expect(token.RBRACE)
		}
	default:
		p.errorExpected(p.pos, "with-qualifier")
	}
}
