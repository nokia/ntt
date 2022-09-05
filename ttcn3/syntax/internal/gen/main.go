// This program reads the TTCN-3 grammar file and generates a parser for it.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/scanner"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/ebnf"
)

const (
	grammarFile = "../../../specs/ttcn3.ebnf"
)

type Data struct {
	TokenMap      map[string]string
	ValueTokens   []string
	KeywordTokens []string
	OtherTokens   []string

	Productions            []*ebnf.Production
	ImplementedProductions map[string]bool
}

var funcs = template.FuncMap{
	"first": First,
	"now":   time.Now,
	"text":  Text,
}

func main() {

	f, err := os.Open(grammarFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	g, err := ebnf.Parse(grammarFile, f)
	if err != nil {
		log.Fatal(err.Error())
	}

	data := Data{
		TokenMap:               Tokens(g),
		Productions:            Productions(g),
		ImplementedProductions: ImplementedProductions("./parser.go"),
	}

	for name := range data.TokenMap {
		switch {
		case strings.HasSuffix(name, "Keyword"):
			data.KeywordTokens = append(data.KeywordTokens, name)
		case strings.HasSuffix(name, "Literal"):
			data.ValueTokens = append(data.ValueTokens, name)
		default:
			data.OtherTokens = append(data.OtherTokens, name)
		}
	}
	sort.Strings(data.KeywordTokens)
	sort.Strings(data.ValueTokens)
	sort.Strings(data.OtherTokens)

	files, err := filepath.Glob("./internal/gen/templates/*")
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, file := range files {
		t, err := template.New(file).Funcs(funcs).ParseFiles(file)
		if err != nil {
			log.Fatal(err.Error())
		}
		var b bytes.Buffer
		if err := t.ExecuteTemplate(&b, filepath.Base(file), data); err != nil {
			log.Fatal(err.Error())
		}
		writeSource(filepath.Base(file), b.Bytes())
	}
}

func ImplementedProductions(file string) map[string]bool {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		log.Fatal(err.Error())
	}

	isParserMethod := func(fun *ast.FuncDecl) bool {
		if fun.Recv == nil || len(fun.Recv.List) != 1 {
			return false
		}
		typ, ok := fun.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			return false
		}
		return typ.X.(*ast.Ident).String() == "parser"
	}

	rules := make(map[string]bool)
	for _, decl := range f.Decls {
		if fun, ok := decl.(*ast.FuncDecl); ok && isParserMethod(fun) {
			if name := fun.Name.String(); strings.HasPrefix(name, "parse") {
				name = strings.TrimPrefix(name, "parse")
				rules[name] = true
			}
		}
	}
	return rules
}

func writeSource(file string, b []byte) {
	b2, err := format.Source(b)
	if err != nil {
		log.Printf("%s: warning: %s\n", file, err.Error())
	} else {
		b = b2
	}

	if err := os.WriteFile(file, b, 0644); err != nil {
		log.Fatal(err)
	}
}

// First returns the first token set of a given expression
func First(g ebnf.Grammar, x ebnf.Expression) []string {
	return first(g, x, make(map[ebnf.Expression]bool))
}

func first(g ebnf.Grammar, x ebnf.Expression, v map[ebnf.Expression]bool) []string {
	switch x := x.(type) {
	case *ebnf.Token:
		return []string{x.String}
	case ebnf.Sequence:
		var ret []string
		if len(x) > 0 {
			ret = append(ret, first(g, x[0], v)...)
		}
		if _, ok := x[0].(*ebnf.Option); ok && len(x) > 1 {
			ret = append(ret, first(g, x[1], v)...)
		}
		return ret
	case *ebnf.Name:
		if !v[x] {
			v[x] = true
			return first(g, g[x.String], v)
		}
		return nil

	case *ebnf.Option:
		return first(g, x.Body, v)
	case *ebnf.Repetition:
		return first(g, x.Body, v)
	case ebnf.Alternative:
		var ret []string
		for _, alt := range x {
			ret = append(ret, first(g, alt, v)...)
		}
		return ret
	case *ebnf.Group:
		return first(g, x.Body, v)
	case *ebnf.Production:
		if x.Expr == nil {
			return nil
		}
		return first(g, x.Expr, v)
	default:
		log.Printf("first: unhandled expression type: %T\n", x)
		return nil
	}
}

func Text(p *ebnf.Production) (string, error) {
	b, err := ioutil.ReadFile(grammarFile)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	s := bufio.NewScanner(bytes.NewReader(b[p.Pos().Offset:]))
	for s.Scan() {
		line := s.Text()
		end, ok := productionEnd(line)
		fmt.Fprintf(&buf, "// %s\n", line[:end])
		if ok {
			break
		}
	}

	return buf.String(), nil
}

func productionEnd(text string) (int, bool) {
	var s scanner.Scanner
	s.Init(strings.NewReader(text))
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if s.TokenText() == "." {
			return s.Pos().Offset, true
		}
	}
	return len(text), false
}

// Productions returns the productions of the grammar in the order they appear
// in the source file.
func Productions(g ebnf.Grammar) []*ebnf.Production {
	ret := make([]*ebnf.Production, 0, len(g))
	for _, prod := range g {
		if !IsLexical(prod.Name.String) {
			ret = append(ret, prod)
		}
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Pos().Offset < ret[j].Pos().Offset
	})
	return ret
}

func Tokens(g ebnf.Grammar) map[string]string {
	names := map[string]string{
		"+":  "Add",
		"-":  "Sub",
		"*":  "Mul",
		"/":  "Div",
		"&":  "Concat",
		"?":  "Any",
		"!":  "Exclude",
		"<":  "Less",
		">":  "Greater",
		"(":  "LeftParen",
		"[":  "LeftBracket",
		"{":  "LeftBrace",
		",":  "Comma",
		".":  "Dot",
		")":  "RightParen",
		"]":  "RightBracket",
		"}":  "RightBrace",
		";":  "Semicolon",
		":":  "Colon",
		"!=": "NotEqual",
		"->": "Arrow",
		"..": "DotDot",
		"::": "ColonColon",
		":=": "Assign",
		"<<": "ShiftLeft",
		"<=": "LessEqual",
		"<@": "RotateLeft",
		"==": "Equal",
		"=>": "DoubleArrow",
		">=": "GreaterEqual",
		">>": "ShiftRight",
		"@>": "RotateRight",

		"@default":     "AtDefault",
		"@local":       "AtLocal",
		"not_a_number": "NotANumber",

		"true":   "TrueLiteral",
		"false":  "FalseLiteral",
		"none":   "NoneLiteral",
		"pass":   "PassLiteral",
		"inconc": "InconcLiteral",
		"fail":   "FailLiteral",
		"error":  "ErrorLiteral",
	}

	m := make(map[string]string)
	for _, prod := range g {
		Inspect(prod.Expr, func(e ebnf.Expression) bool {
			if tok, ok := e.(*ebnf.Token); ok {
				name := fmt.Sprintf("%sKeyword", strings.Title(tok.String))
				if n, ok := names[tok.String]; ok {
					name = n
				}
				m[name] = tok.String
			}
			return true
		})
	}
	return m
}

// Inspect traverses the given expression and calls the given function for each.
func Inspect(e ebnf.Expression, fn func(e ebnf.Expression) bool) bool {
	if e == nil {
		return true
	}

	if !fn(e) {
		return false
	}

	switch e := e.(type) {
	case ebnf.Alternative:
		for _, alt := range e {
			if !Inspect(alt, fn) {
				return false
			}
		}
	case ebnf.Sequence:
		for _, seq := range e {
			if !Inspect(seq, fn) {
				return false
			}
		}

	case *ebnf.Group:
		if e.Body != nil {
			Inspect(e.Body, fn)
		}

	case *ebnf.Option:
		if e.Body != nil {
			Inspect(e.Body, fn)
		}

	case *ebnf.Repetition:
		if e.Body != nil {
			Inspect(e.Body, fn)
		}

	}
	return true
}

// IsLexical returns true, when given name is a lexical production.
func IsLexical(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return !unicode.IsUpper(ch)
}
