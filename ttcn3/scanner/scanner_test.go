package scanner

import (
	"testing"

	"github.com/nokia/ntt/ttcn3/token"
)

type tt struct {
	tok token.Token
	lit string
}

type elt struct {
	tok token.Token
	lit string
}

var tokens = [...]elt{
	// value tokens
	{token.IDENT, `_`},
	{token.IDENT, `_foo`},
	{token.IDENT, `f__2bar`},

	{token.INT, `0`},
	{token.INT, `1`},
	{token.INT, `100000000000000000000000000000`},
	{token.INT, `3`},

	{token.FLOAT, `0.0`},
	{token.FLOAT, `0.00001`},
	{token.FLOAT, `1E2`},
	{token.FLOAT, `1e2`},
	{token.FLOAT, `12.3e-4`},

	{token.STRING, `"@foo "`},
	{token.STRING, `""`},
	{token.STRING, `"\""`},
	{token.STRING, `"\\"`},
	{token.STRING, `""""`},

	{token.BSTRING, `'1101'b`},
	{token.BSTRING, `'11?*1'x`},

	{token.MODIF, `@`},
	{token.MODIF, `@lazy`},
	{token.MODIF, `@123`},

	{token.ADD, `+`},
	{token.SUB, `-`},
	{token.MUL, `*`},
	{token.DIV, `/`},
	{token.MOD, `mod`},
	{token.REM, `rem`},

	{token.AND, `and`},
	{token.OR, `or`},
	{token.XOR, `xor`},
	{token.NOT, `not`},

	{token.AND4B, `and4b`},
	{token.OR4B, `or4b`},
	{token.XOR4B, `xor4b`},
	{token.NOT4B, `not4b`},
	{token.SHR, `>>`},
	{token.SHL, `<<`},
	{token.ROR, `@>`},
	{token.ROL, `<@`},
	{token.CONCAT, `&`},

	{token.REDIR, `->`},
	{token.ANY, `?`},
	{token.EXCL, `!`},
	{token.RANGE, `..`},
	{token.ASSIGN, `:=`},

	{token.EQ, `==`},
	{token.NE, `!=`},
	{token.LT, `<`},
	{token.LE, `<=`},
	{token.GT, `>`},
	{token.GE, `>=`},

	{token.LPAREN, `(`},
	{token.LBRACK, `[`},
	{token.LBRACE, `{`},
	{token.COMMA, `,`},
	{token.DOT, `.`},

	{token.RPAREN, `)`},
	{token.RBRACE, `}`},
	{token.RBRACK, `]`},
	{token.SEMICOLON, `;`},
	{token.COLON, `:`},
}

const whitespace string = "\t \n \r\n\n"

var source = func() []byte {
	var src []byte
	for _, v := range tokens {
		src = append(src, v.lit...)
		src = append(src, whitespace...)
	}
	return src
}()

var fset = token.NewFileSet()

func TestScan(t *testing.T) {

	// error handler
	eh := func(p token.Position, msg string) {
		t.Errorf("Error handler called ( %s: %s)", p, msg)
	}

	var s Scanner
	s.Init(fset.AddFile("", fset.Base(), len(source)), source, eh, ScanComments)

	// start position
	epos := token.Position{
		Filename: "",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	i := 0
	for {
		pos, tok, lit := s.Scan()

		if tok == token.EOF {
			break
		}

		if i < len(tokens) {
			e := tokens[i]

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

	if i+1 < len(tokens) {
		t.Errorf("Missing tokens")
	}
}
