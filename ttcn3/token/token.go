// Package token defines constants representing the lexical tokens of the Go
// programming language and basic operations on tokens (printing, predicates).
//
package token

import "strconv"

// Token is the set of lexical tokens of the Go programming language.
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	COMMENT
	PREPROC

	literal_beg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT   // main
	INT     // 12345
	FLOAT   // 123.45
	STRING  // "abc"
	BSTRING // '101?F'H
	MODIF   // @fuzzy
	literal_end

	operator_beg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	DIV // /

	SHL    // <<
	ROL    // <@
	SHR    // >>
	ROR    // @>
	CONCAT // &

	REDIR  // ->
	ANY    // ?
	EXCL   // !
	RANGE  // ..
	ASSIGN // :=

	EQ // ==
	NE // !=
	LT // <
	LE // <=
	GT // >
	GE // >=

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	DOT    // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
	operator_end

	keyword_beg
	// Keywords
	MOD // mod
	REM // rem

	AND // and
	OR  // or
	XOR // xor
	NOT // not

	AND4B // and4b
	OR4B  // or4b
	XOR4B // xor4b
	NOT4B // not4b

	MODULE
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",
	PREPROC: "PREPROC",

	IDENT:   "IDENT",
	INT:     "INT",
	FLOAT:   "FLOAT",
	STRING:  "STRING",
	BSTRING: "BSTRING",
	MODIF:   "MODIF",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",

	SHL:    "<<",
	ROL:    "<@",
	SHR:    ">>",
	ROR:    "@>",
	CONCAT: "&",

	REDIR: "->",
	ANY:   "?",
	EXCL:  "!",
	RANGE: "..",

	ASSIGN: ":=",
	EQ:     "==",
	NE:     "!=",
	LT:     "<",
	LE:     "<=",
	GT:     ">",
	GE:     ">=",

	LPAREN:    "(",
	LBRACK:    "[",
	LBRACE:    "{",
	COMMA:     ",",
	DOT:       ".",
	RPAREN:    ")",
	RBRACK:    "]",
	RBRACE:    "}",
	SEMICOLON: ";",
	COLON:     ":",

	MOD: "mod",
	REM: "rem",

	AND: "and",
	OR:  "or",
	XOR: "xor",
	NOT: "not",

	AND4B: "and4b",
	OR4B:  "or4b",
	XOR4B: "xor4b",
	NOT4B: "not4b",

	MODULE: "module",
}

// String returns the string corresponding to the token tok.
// For operators, delimiters, and keywords the string is the actual
// token character sequence (e.g., for the token ADD, the string is
// "+"). For all other tokens the string corresponds to the token
// constant name (e.g. for the token IDENT, the string is "IDENT").
//
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators. The highest
// precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
//
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 15
)

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
//
func (tok Token) Precedence() int {
	switch tok {
	case ASSIGN:
		return 1
	case EXCL:
		return 2
	case OR:
		return 3
	case XOR:
		return 4
	case AND:
		return 5
	case NOT:
		return 6
	case EQ, NE:
		return 7
	case LT, LE, GT, GE:
		return 8
	case SHR, SHL, ROR, ROL:
		return 9
	case OR4B:
		return 10
	case XOR4B:
		return 11
	case AND4B:
		return 12
	case NOT4B:
		return 13
	case ADD, SUB, CONCAT:
		return 14
	case MUL, DIV, REM, MOD:
		return 15
	}
	return LowestPrec
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
//
func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return literal_beg < tok && tok < literal_end }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
//
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }
