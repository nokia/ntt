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
	case token.IDENT, token.UNIVERSAL, token.CHARSTRING, token.ADDRESS:
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
	default:
		p.errorExpected(p.pos, "type definition")
	}
	return nil
}

func (p *parser) parseNestedType() {
	if p.trace {
		defer un(trace(p, "NestedType"))
	}
	switch p.tok {
	case token.IDENT, token.ADDRESS, token.NULL, token.CHARSTRING, token.UNIVERSAL:
		p.parseTypeRef()
	case token.UNION:
		p.next()
		p.parseNestedStructType()
	case token.SET, token.RECORD:
		p.next()
		if p.tok == token.LBRACE {
			p.parseNestedStructType()
			break
		}
		p.parseNestedListType()
	case token.ENUMERATED:
		p.parseNestedEnumType()
	default:
		p.errorExpected(p.pos, "type definition")
	}
}

func (p *parser) parseNestedStructType() {
	if p.trace {
		defer un(trace(p, "NestedStructType"))
	}
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
	p.parseWith()
}

func (p *parser) parseStructField() {
	if p.trace {
		defer un(trace(p, "StructField"))
	}
	if p.tok == token.MODIF {
		p.next() // @default
	}
	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()

	if p.tok == token.OPTIONAL {
		p.next()
	}
}

func (p *parser) parseNestedListType() {
	if p.trace {
		defer un(trace(p, "NestedListType"))
	}
	p.parseLength()
	p.expect(token.OF)
	p.parseNestedType()
}

func (p *parser) parseListType() {
	if p.trace {
		defer un(trace(p, "ListType"))
	}
	p.parseLength()

	p.expect(token.OF)
	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()
	p.parseWith()
}

func (p *parser) parseNestedEnumType() {
	if p.trace {
		defer un(trace(p, "NestedEnumType"))
	}
	p.next()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE {
		p.parseExpr()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseEnumType() {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}
	p.next()
	p.parseIdent()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE {
		p.parseExpr()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
	p.parseWith()
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
	p.parseWith()
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
	p.parseWith()
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
	p.parseWith()

}

func (p *parser) parseSubType() *ast.SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}

	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()

	p.parseWith()
	return nil
}

func (p *parser) parseConstraint() {
	// TODO(mef) fix constraints consumed by previous PrimaryExpr

	if p.tok == token.LPAREN {
		p.next()
		p.parseExprList()
		p.expect(token.RPAREN)
	}

	p.parseLength()
}

func (p *parser) parseLength() {
	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}
}

func (p *parser) parseTypeRef() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	x := p.parsePrimaryExpr()
	return x
}
