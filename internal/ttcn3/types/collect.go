package types

import (
	"fmt"

	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

func (info *Info) CollectInfo(n ast.Node) {
	if info.Types == nil {
		info.Types = make(map[ast.Expr]Type)
	}
	ast.Apply(n, info.enter, info.exit)
}

// enter walks the AST top-down, builds scopes and collects symbols.
func (info *Info) enter(c *ast.Cursor) bool {
	return true
}

// exit walks the AST bottom-up, resolves types and checks for errors.
func (info *Info) exit(c *ast.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ValueLiteral:
		t, ok := literalTypes[n.Tok.Kind]
		if !ok {
			panic("unhandled literal")
		}
		info.Types[n] = t

	case *ast.UnaryExpr:
		info.Types[n] = info.expressionType(n.Op.Kind, n.X)
	}

	return true
}

// expressionType returns the type of an expression, checking compatibility of
// the operater and its operands. expressionType promotes integers to floats
// if one operand is a float. It returns the Invalid type if the expression is
// not valid.
func (info *Info) expressionType(op token.Kind, operands ...ast.Expr) Type {

	// the operator denominates the expected type of the result. Abstract
	// types like "numerical" or "string" are possible.
	expected, ok := operatorTypes[op]
	if !ok {
		info.error(fmt.Errorf("unhandled operator %v", op))
		return Typ[Invalid]
	}

	hasFloats := false
	for _, operand := range operands {
		actual, ok := info.Types[operand]
		if !ok {
			panic(fmt.Sprintf("no type for operand %+v", operand))
		}

		if !Compatible(actual, expected) {
			info.invalidTypeError(operand, actual, expected)
			expected = Typ[Invalid]
		}

		if actual == Typ[Float] {
			hasFloats = true
		}
	}

	// Use the most specific type of the operands.
	if expected == Typ[Numerical] {
		expected = Typ[Integer]
		if hasFloats {
			expected = Typ[Float]
		}
	}

	return expected
}

var operatorTypes = map[token.Kind]Type{
	token.ADD: Typ[Numerical],
	token.SUB: Typ[Numerical],
	token.MUL: Typ[Numerical],
	token.DIV: Typ[Numerical],

	token.GT: Typ[Numerical],
	token.GE: Typ[Numerical],
	token.LT: Typ[Numerical],
	token.LE: Typ[Numerical],

	token.MOD:   Typ[Integer],
	token.REM:   Typ[Integer],
	token.RANGE: Typ[Integer],
	token.EXCL:  Typ[Integer],

	token.AND: Typ[Boolean],
	token.OR:  Typ[Boolean],
	token.NOT: Typ[Boolean],
	token.XOR: Typ[Boolean],

	token.SHL:   Typ[Bitstring],
	token.SHR:   Typ[Bitstring],
	token.ROL:   Typ[Bitstring],
	token.ROR:   Typ[Bitstring],
	token.AND4B: Typ[Bitstring],
	token.OR4B:  Typ[Bitstring],
	token.NOT4B: Typ[Bitstring],
	token.XOR4B: Typ[Bitstring],

	token.CONCAT: Typ[String],
}

var literalTypes = map[token.Kind]Type{
	token.INT:    Typ[Integer],
	token.FLOAT:  Typ[Float],
	token.NAN:    Typ[Float],
	token.ANY:    Typ[Template],
	token.NULL:   Typ[Component],
	token.OMIT:   Typ[Omit],
	token.FALSE:  Typ[Boolean],
	token.TRUE:   Typ[Boolean],
	token.NONE:   Typ[Verdict],
	token.INCONC: Typ[Verdict],
	token.PASS:   Typ[Verdict],
	token.FAIL:   Typ[Verdict],
	token.ERROR:  Typ[Verdict],
}
