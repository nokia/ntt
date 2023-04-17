package parser

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	tokn "github.com/nokia/ntt/ttcn3/token"
)

type tt struct {
	tok tokn.Kind
	lit string
}

type elt struct {
	tok tokn.Kind
	lit string
}

var ttokens = [...]elt{
	// value tokens
	{tokn.IDENT, `_`},
	{tokn.IDENT, `_foo`},
	{tokn.IDENT, `f__2bar`},
	{tokn.IDENT, `%definitionId`}, // Titan macros are like identifiers

	{tokn.INT, `0`},
	{tokn.INT, `1`},
	{tokn.INT, `100000000000000000000000000000`},
	{tokn.INT, `3`},

	{tokn.FLOAT, `0.0`},
	{tokn.FLOAT, `0.00001`},
	{tokn.FLOAT, `1E2`},
	{tokn.FLOAT, `1e2`},
	{tokn.FLOAT, `12.3e-4`},

	{tokn.STRING, `"@foo "`},
	{tokn.STRING, `""`},
	{tokn.STRING, `"\""`},
	{tokn.STRING, `"\\"`},
	{tokn.STRING, `""""`},

	{tokn.BSTRING, `'1101'b`},
	{tokn.BSTRING, `'11?*1'x`},

	{tokn.MODIF, `@`},
	{tokn.MODIF, `@lazy`},
	{tokn.MODIF, `@123`},

	{tokn.ADD, `+`},
	{tokn.SUB, `-`},
	{tokn.MUL, `*`},
	{tokn.DIV, `/`},
	{tokn.MOD, `mod`},
	{tokn.REM, `rem`},

	{tokn.AND, `and`},
	{tokn.OR, `or`},
	{tokn.XOR, `xor`},
	{tokn.NOT, `not`},

	{tokn.AND4B, `and4b`},
	{tokn.OR4B, `or4b`},
	{tokn.XOR4B, `xor4b`},
	{tokn.NOT4B, `not4b`},
	{tokn.SHR, `>>`},
	{tokn.SHL, `<<`},
	{tokn.ROR, `@>`},
	{tokn.ROL, `<@`},
	{tokn.CONCAT, `&`},

	{tokn.REDIR, `->`},
	{tokn.ANY, `?`},
	{tokn.EXCL, `!`},
	{tokn.RANGE, `..`},
	{tokn.ASSIGN, `:=`},

	{tokn.EQ, `==`},
	{tokn.NE, `!=`},
	{tokn.LT, `<`},
	{tokn.LE, `<=`},
	{tokn.GT, `>`},
	{tokn.GE, `>=`},

	{tokn.LPAREN, `(`},
	{tokn.LBRACK, `[`},
	{tokn.LBRACE, `{`},
	{tokn.COMMA, `,`},
	{tokn.DOT, `.`},

	{tokn.RPAREN, `)`},
	{tokn.RBRACE, `}`},
	{tokn.RBRACK, `]`},
	{tokn.SEMICOLON, `;`},
	{tokn.COLON, `:`},
	{tokn.COLONCOLON, `::`},
}

const whitespace string = "\t \n \r\n\n"

var source = func() []byte {
	var src []byte
	for _, v := range ttokens {
		src = append(src, v.lit...)
		src = append(src, whitespace...)
	}
	return src
}()

var fset = loc.NewFileSet()

func TestScan(t *testing.T) {

	// error handler
	eh := func(p loc.Position, msg string) {
		t.Errorf("Error handler called ( %s: %s)", p, msg)
	}

	var s Scanner
	s.Init(fset.AddFile("", fset.Base(), len(source)), source, eh)

	// start position
	epos := loc.Position{
		Filename: "",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	i := 0
	for {
		pos, tok, lit := s.Scan()

		if tok == tokn.EOF {
			break
		}

		if i < len(ttokens) {
			e := ttokens[i]

			if tok != e.tok {
				t.Errorf("Bad token for %q. Expected %q, got %q",
					lit, e.tok, tok)
			}
			if !tok.IsOperator() {
				if lit != e.lit {
					t.Errorf("Bad literal. Expected %q, got %q",
						e.lit, lit)
				}
			}
			if fset.Position(pos) != epos {
				t.Errorf("Bad position. Expected %q, got %q",
					epos, fset.Position(pos))
			}

			epos.Offset += len(e.lit) + len(whitespace)
			epos.Line += 3

		} else {
			t.Errorf("Unexpected token for %q", lit)
		}

		i++
	}

	if i+1 < len(ttokens) {
		t.Errorf("Missing tokens")
	}
}
