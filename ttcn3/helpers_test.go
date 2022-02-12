package ttcn3_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

const CURSOR = "¶"

type SliceMap map[string][]string

func extractCursor(input string) (loc.Pos, string) {
	return loc.Pos(strings.Index(input, CURSOR) + 1), strings.Replace(input, CURSOR, "", 1)
}

func parseFile(t *testing.T, name string, input string) *ttcn3.Tree {
	t.Helper()
	file := fmt.Sprintf("%s.ttcn3", name)
	fs.SetContent(file, []byte(input))
	tree := ttcn3.ParseFile(file)
	if tree.Err != nil {
		t.Fatalf("%s", tree.Err.Error())
	}
	return tree
}

func enumerateIDs(root ast.Node) map[*ast.Ident]string {
	// build ID map
	ids := make(map[*ast.Ident]string)
	counter := make(map[string]int)
	ast.Inspect(root, func(n ast.Node) bool {
		if x, ok := n.(*ast.Ident); ok {
			i := counter[x.String()]
			counter[x.String()]++
			ids[x] = fmt.Sprintf("%s%d", x.String(), i)
		}
		return true
	})

	return ids
}

func parentNodes(tree *ttcn3.Tree, cursor loc.Pos) (n ast.Expr, s []ast.Node) {
	s = tree.SliceAt(cursor)
	if len(s) < 2 {
		return nil, nil
	}

	if tok, ok := s[0].(ast.Token); ok && tok.Kind == token.IDENT {
		n, s = s[1].(ast.Expr), s[2:]
	}
	if len(s) > 0 {
		if x, ok := s[0].(*ast.SelectorExpr); ok && n == x.Sel {
			n, s = x, s[1:]
		}
	}
	return n, s
}

func importedDefs(db *ttcn3.DB, id string, module string) []string {
	var s []string
	mod := moduleFrom("file1.ttcn3", module)
	for _, d := range db.FindImportedDefinitions(id, mod) {
		name := ast.Name(d.Node)
		file := d.Tree.Position(d.Node.Pos()).Filename
		s = append(s, fmt.Sprintf("%s:%s", name, file))
	}
	return s
}

func moduleFrom(file, module string) *ast.Module {
	tree := ttcn3.ParseFile(file)
	for _, m := range tree.Modules() {
		if ast.Name(m) == module {
			return m
		}
	}
	return nil
}

func testMapsEqual(t *testing.T, a, b SliceMap) {
	if !equalSliceMap(a, b) {
		t.Errorf("Maps not equal:\n\t got = %v\n\twant = %v", a, b)
	}
}

func equalSliceMap(a, b SliceMap) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !equal(v, b[k]) {
			return false
		}
	}
	return true
}

func makeSliceMap(m map[string]map[string]bool) SliceMap {
	sm := SliceMap{}
	for k, v := range m {
		sm[k] = make([]string, 0, len(v))
		for kk := range v {
			sm[k] = append(sm[k], kk)
		}
	}
	return sm
}

// Unwrap first node from NodeLists
func unwrapFirst(n ast.Node) ast.Node {
	switch n := n.(type) {
	case ast.NodeList:
		if len(n) == 0 {
			return nil
		}
		return unwrapFirst(n[0])
	case *ast.ExprStmt:
		return unwrapFirst(n.Expr)
	case *ast.DeclStmt:
		return unwrapFirst(n.Decl)
	case *ast.ModuleDef:
		return unwrapFirst(n.Def)
	default:
		return n
	}
}

// equal returns true if a and b are equal, order is ignored.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for i := range a {
		m[a[i]]++
		m[b[i]]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func nameSlice(scp *ttcn3.Scope) []string {
	s := make([]string, 0, len(scp.Names))
	for k := range scp.Names {
		s = append(s, k)
	}
	return s
}
