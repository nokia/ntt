package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/parser"
	"github.com/nokia/ntt/ttcn3/token"
)

var (
	count = make(map[string]int)
	line  = 0
)

func main() {

	// Scanner reads stdin line by line
	r := bufio.NewReader(os.Stdin)

	var (
		text string
		err  error
	)
	for {
		line++

		text, err = r.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if err != nil {
			break
		}

		// We are only interested in ptrx events ending with +consume. Until we
		// require other events for coverage, do the filtering before splitting
		// the whole string.
		if !strings.HasSuffix(text, "+consume") {
			continue
		}
		s := strings.Split(text, "|")
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
		ast, err2 := parser.ParseExpr(loc.NewFileSet(), "stdin", tmpl)
		if err2 != nil {
			log.Println(err2)
			continue
		}

		// Calculate coverage
		cover(typ, ast)
	}

	if err != io.EOF {
		log.Println(err)
	}

	// Report coverage
	keys := make([]string, 0, len(count))
	for k := range count {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {

		// Do not report matches of whole structs:
		//
		//      MessageA 0          <-- Do not display those lines
		//      MessageA.Field1 23
		//      MessageA.Field2 5
		//      MessageA.Field3 0
		//
		if i+1 < len(keys) && strings.HasPrefix(keys[i+1], k+".") {
			continue
		}

		fmt.Printf("%s\t%d\n", k, count[k])
	}
}

func cover(typ string, tmpl ast.Expr) {

	switch x := tmpl.(type) {

	// Low hanging fruits: We assume the ntt runtime log format uses assignment
	// lists (`{ a:=1, b:=2}`) for structured types and value lists (`{1,2}`)
	// for list types, so we don't need a ttcn3 typesystem, yet. We don't get
	// full coverage for unions and optional fields, though.
	case *ast.CompositeLiteral:
		// An empty list means full coverage.
		if len(x.List) == 0 {
			count[typ]++
			return
		}

		for i := range x.List {
			switch x := x.List[i].(type) {
			case *ast.BinaryExpr:
				// The only binary expressions we expect are assignments (field := value)
				if x.Op.Kind != token.ASSIGN {
					notImplemented(typ, x)
					continue
				}

				// We expect the left hand side to be a plain identifier.
				id, ok := x.X.(*ast.Ident)
				if !ok {
					notImplemented(typ, x.X)
					continue
				}

				// Descend into right hand side.
				cover(fmt.Sprintf("%s.%s", typ, id.String()), x.Y)

			default:
				// All other expressions are interpreted as list elements.
				cover(fmt.Sprintf("%s[%d]", typ, i), x)
			}
		}
		return

	// Scalar types are easy to evaluate
	case *ast.ValueLiteral:
		switch x.Tok.Kind {
		// `omit`, `?` and `*` do not increase counter, because they do not
		// contribute to message coverage.
		case token.ANY, token.MUL, token.OMIT:
			count[typ] += 0

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
			token.NULL:
			count[typ]++
		}

	// Handle enums
	case *ast.Ident:
		count[typ]++

	// Handle permution/superset/...
	case *ast.CallExpr:
		count[typ]++

	default:
		notImplemented(typ, tmpl)
	}
}

func notImplemented(typ string, n ast.Node) {
	fmt.Fprintf(os.Stderr, "line %d: Type %T for field %s not implemented.", line, n, typ)
}
