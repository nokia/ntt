package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

var (
	count = make(map[string]int)
	line  = 0
)

func main() {

	// Scanner reads stdin line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line++
		// We are only interested in ptrx events ending with +consume. Until we
		// require other events for coverage, do the filtering before splitting
		// the whole string.
		if !strings.HasSuffix(scanner.Text(), "+consume") {
			continue
		}
		s := strings.Split(scanner.Text(), "|")
		if s[1] != "ptrx" {
			continue
		}

		// Extract type and value from template string
		tmpl := strings.TrimPrefix(s[4], "value=")
		if tmpl == "" {
			continue
		}
		idx := strings.IndexByte(tmpl, ':')
		if idx < 0 {
			continue
		}
		typ := tmpl[:idx]
		tmpl = tmpl[idx+1:]

		// Parse template
		ast, err := parser.ParseExpr(loc.NewFileSet(), "stdin", tmpl)
		if err != nil {
			log.Println(err)
			continue
		}

		// Calculate coverage
		if err := cover(typ, ast); err != nil {
			log.Println(err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	// Report coverage
	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, count[k])
	}
}

func cover(typ string, tmpl ast.Expr) error {

	switch x := tmpl.(type) {

	// Low hanging fruits: We assume the k3 runtime log format uses assignment
	// lists (`{ a:=1, b:=2}`) for structured types and value lists (`{1,2}`)
	// for list types, so we don't need a ttcn3 typesystem, yet. We don't get
	// full coverage for unions and optional fields, though.
	case *ast.CompositeLiteral:
		// An empty list means full coverage.
		if len(x.List) == 0 {
			count[typ]++
			return nil
		}

		for i := range x.List {
			switch x := x.List[i].(type) {
			case *ast.BinaryExpr:
				if x.Op.Kind != token.ASSIGN {
					return notImplemented(typ, x)
				}
				if id, ok := x.X.(*ast.Ident); ok {
					if err := cover(fmt.Sprintf("%s.%s", typ, id.String()), x.Y); err != nil {
						return err
					}
					continue
				}
				return notImplemented(typ, x.X)

			default:
				if err := cover(fmt.Sprintf("%s[%d]", typ, i), x); err != nil {
					return err
				}
			}
		}
		return nil

	// Scalar types are easy to evaluate
	case *ast.ValueLiteral:
		switch x.Tok.Kind {
		// `?` and `*` do not increase counter, because they do not
		// contribute to message coverage.
		case token.ANY, token.MUL:
			count[typ] += 0
			return nil
		case token.SUB,
			token.INT,
			token.BSTRING,
			token.ERROR,
			token.FAIL,
			token.FALSE,
			token.FLOAT,
			token.INCONC,
			token.NAN,
			token.NONE,
			token.PASS,
			token.STRING,
			token.TRUE,
			token.OMIT,
			token.NULL:
			count[typ]++
			return nil
		}

	// Handle enums
	case *ast.Ident:
		count[typ]++
		return nil
	}
	return notImplemented(typ, tmpl)
}

func notImplemented(typ string, n ast.Node) error {
	return errors.New(fmt.Sprintf("%d: Type %T for field %s not implemented.", line, n, typ))
}
