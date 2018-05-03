package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseType() ast.Decl {
	if p.trace {
		defer un(trace(p, "Type"))
	}
	p.next()
	switch p.tok {
	case token.IDENT:
		p.parseSubType()
	case token.UNION:
		p.next()
		p.parseStructType()
	case token.SET, token.RECORD:
		p.next()
		if p.tok == token.IDENT {
			p.parseStructType()
			break
		}
		p.parseListType()
	case token.ENUMERATED:
		p.parseEnumType()
	case token.PORT:
		p.parsePortType()
	case token.COMPONENT:
		p.parseComponentType()
	case token.FUNCTION, token.ALTSTEP, token.TESTCASE:
		p.parseBehaviourType()
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseNestedType() {
	if p.trace {
		defer un(trace(p, "NestedType"))
	}
	p.next()
}

func (p *parser) parseStructType() {
	if p.trace {
		defer un(trace(p, "StructType"))
	}
	p.parseIdent()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE {
		p.parseStructField()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseStructField() {
	if p.trace {
		defer un(trace(p, "StructField"))
	}
	if p.tok == token.MODIF {
		p.next() // @default
	}
	p.parseNestedType()
	p.parseIdent()
	if p.tok == token.OPTIONAL {
		p.next()
	}
}

func (p *parser) parseListType() {
	if p.trace {
		defer un(trace(p, "ListType"))
	}
	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}

	p.expect(token.OF)
	p.parseNestedType()
	p.parseIdent()

	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}
}

func (p *parser) parseEnumType() {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}
	p.next()
}

func (p *parser) parsePortType() {
	if p.trace {
		defer un(trace(p, "PortType"))
	}
	p.next()
	p.parseIdent()
	switch p.tok {
	case token.MIXED, token.MESSAGE, token.PROCEDURE:
		p.next()
	default:
		p.errorExpected(p.pos, "'message' or 'procedure'")
	}

	p.expect(token.LBRACE)
	for p.tok != token.RBRACE {
		p.parsePortAttribute()
		p.expectSemi()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parsePortAttribute() {
	if p.trace {
		defer un(trace(p, "PortAttribute"))
	}
	switch p.tok {
	case token.IN, token.OUT, token.INOUT:
		p.next()
		p.parseRefList()
	case token.ADDRESS:
		p.next()
		p.parseRefList()
	case token.MAP, token.UNMAP:
		p.next()
		p.expect(token.PARAM)
		p.parseParameters()
	}
}

func (p *parser) parseComponentType() {
	if p.trace {
		defer un(trace(p, "ComponentType"))
	}
	p.next()
	p.parseIdent()
	if p.tok == token.EXTENDS {
		p.next()
		p.parseRefList()
	}
	p.parseBlockStmt()
}

func (p *parser) parseBehaviourType() {
	if p.trace {
		defer un(trace(p, "BehaviourType"))
	}
	p.next()
	p.next()
	p.parseParameters()
	if p.tok == token.RUNS {
		p.next()
		p.expect(token.ON)
		p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		p.next()
		p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		p.parseReturn()
	}

}

func (p *parser) parseSubType() *ast.SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}

	p.parseTypeRef()
	p.parseIdent()
	if p.tok == token.LPAREN {
		p.next()
		p.parseExprList()
		p.expect(token.RPAREN)
	}

	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}

	return nil
}

func (p *parser) parseTypeRef() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	x := p.parsePrimaryExpr()
	return x
}
