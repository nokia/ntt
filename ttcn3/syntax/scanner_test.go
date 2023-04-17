package syntax

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
)

type tt struct {
	tok Kind
	lit string
}

type elt struct {
	tok Kind
	lit string
}

var ttokens = [...]elt{
	// value tokens
	{IDENT, `_`},
	{IDENT, `_foo`},
	{IDENT, `f__2bar`},
	{IDENT, `%definitionId`}, // Titan macros are like identifiers

	{INT, `0`},
	{INT, `1`},
	{INT, `100000000000000000000000000000`},
	{INT, `3`},

	{FLOAT, `0.0`},
	{FLOAT, `0.00001`},
	{FLOAT, `1E2`},
	{FLOAT, `1e2`},
	{FLOAT, `12.3e-4`},

	{STRING, `"@foo "`},
	{STRING, `""`},
	{STRING, `"\""`},
	{STRING, `"\\"`},
	{STRING, `""""`},

	{BSTRING, `'1101'b`},
	{BSTRING, `'11?*1'x`},

	{MODIF, `@`},
	{MODIF, `@lazy`},
	{MODIF, `@123`},

	{ADD, `+`},
	{SUB, `-`},
	{MUL, `*`},
	{DIV, `/`},
	{MOD, `mod`},
	{REM, `rem`},

	{AND, `and`},
	{OR, `or`},
	{XOR, `xor`},
	{NOT, `not`},

	{AND4B, `and4b`},
	{OR4B, `or4b`},
	{XOR4B, `xor4b`},
	{NOT4B, `not4b`},
	{SHR, `>>`},
	{SHL, `<<`},
	{ROR, `@>`},
	{ROL, `<@`},
	{CONCAT, `&`},

	{REDIR, `->`},
	{ANY, `?`},
	{EXCL, `!`},
	{RANGE, `..`},
	{ASSIGN, `:=`},

	{EQ, `==`},
	{NE, `!=`},
	{LT, `<`},
	{LE, `<=`},
	{GT, `>`},
	{GE, `>=`},

	{LPAREN, `(`},
	{LBRACK, `[`},
	{LBRACE, `{`},
	{COMMA, `,`},
	{DOT, `.`},

	{RPAREN, `)`},
	{RBRACE, `}`},
	{RBRACK, `]`},
	{SEMICOLON, `;`},
	{COLON, `:`},
	{COLONCOLON, `::`},
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

		if tok == EOF {
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
