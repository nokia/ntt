package ttcn3_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
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

func enumerateIDs(root syntax.Node) map[*syntax.Ident]string {
	// build ID map
	ids := make(map[*syntax.Ident]string)
	counter := make(map[string]int)
	root.Inspect(func(n syntax.Node) bool {
		if x, ok := n.(*syntax.Ident); ok {
			i := counter[x.String()]
			counter[x.String()]++
			ids[x] = fmt.Sprintf("%s%d", x.String(), i)
		}
		return true
	})

	return ids
}

func parentNodes(tree *ttcn3.Tree, cursor loc.Pos) (n syntax.Expr, s []syntax.Node) {
	s = tree.SliceAt(cursor)
	if len(s) < 2 {
		return nil, nil
	}

	if tok, ok := s[0].(syntax.Token); ok && tok.Kind() == syntax.IDENT {
		n, s = s[1].(syntax.Expr), s[2:]
	}
	if len(s) > 0 {
		if x, ok := s[0].(*syntax.SelectorExpr); ok && n == x.Sel {
			n, s = x, s[1:]
		}
	}
	return n, s
}

func importedDefs(db *ttcn3.DB, id string, module string) []string {
	var s []string
	mod := moduleFrom("file1.ttcn3", module)
	for _, d := range db.VisibleModules(id, mod) {
		name := syntax.Name(d.Node)
		file := syntax.SpanOf(d.Node).Filename
		s = append(s, fmt.Sprintf("%s:%s", name, file))
	}
	return s
}

func moduleFrom(file, module string) *syntax.Module {
	tree := ttcn3.ParseFile(file)
	for _, m := range tree.Modules() {
		if syntax.Name(m.Node) == module {
			return m.Node.(*syntax.Module)
		}
	}
	return nil
}

func testMapsEqual(t *testing.T, a, b SliceMap) {
	t.Helper()
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
func unwrapFirst(n syntax.Node) syntax.Node {
	switch n := n.(type) {
	case *syntax.Root:
		return unwrapFirst(&n.NodeList)

	case *syntax.NodeList:
		if len(n.Nodes) == 0 {
			return nil
		}
		return unwrapFirst(n.Nodes[0])
	case *syntax.ExprStmt:
		return unwrapFirst(n.Expr)
	case *syntax.DeclStmt:
		return unwrapFirst(n.Decl)
	case *syntax.ModuleDef:
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

func nodeDesc(n syntax.Node) string {
	s := fmt.Sprintf("%T", n)
	if n := syntax.Name(n); n != "" {
		s += fmt.Sprintf("(%s)", n)
	}
	return s
}
