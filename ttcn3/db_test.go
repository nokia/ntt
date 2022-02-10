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

var (
	file1 = `
	module M1 { import from M2 all }
	module M2 { type enumerated E { E1 } }`

	file2 = `
	module M1 { type component C { var integer x := E}; const int E }
	module M3 { const integer E, E1 := 1 }`

	file3 = `module MX { }`
)

func init() {
	fs.SetContent("file1.ttcn3", []byte(file1))
	fs.SetContent("file2.ttcn3", []byte(file2))
	fs.SetContent("file3.ttcn3", []byte(file3))
}

type SliceMap map[string][]string

func TestIndex(t *testing.T) {
	db := ttcn3.DB{}
	db.Index("file1.ttcn3", "file2.ttcn3")
	testMapsEqual(t, makeSliceMap(db.Modules), SliceMap{
		"M1": []string{"file1.ttcn3", "file2.ttcn3"},
		"M2": []string{"file1.ttcn3"},
		"M3": []string{"file2.ttcn3"},
	})
	testMapsEqual(t, makeSliceMap(db.Names), SliceMap{
		"M1": []string{"file1.ttcn3", "file2.ttcn3"},
		"M2": []string{"file1.ttcn3"},
		"E":  []string{"file1.ttcn3", "file2.ttcn3"},
		"E1": []string{"file1.ttcn3", "file2.ttcn3"},
		"C":  []string{"file2.ttcn3"},
		"M3": []string{"file2.ttcn3"},
	})

}

func TestFindImportedDefinitions(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		db := ttcn3.DB{}
		if defs := db.FindImportedDefinitions("E", "M1"); len(defs) != 0 {
			t.Errorf("Expected 0 definitions, got %v", defs)
		}
	})

	t.Run("regular", func(t *testing.T) {
		db := ttcn3.DB{}
		db.Index("file1.ttcn3", "file2.ttcn3", "file3.ttcn3")

		expected := []string{"M2:file1.ttcn3"}
		actual := importedDefs(&db, "E", "M1")
		if !equal(actual, expected) {
			t.Errorf("Mismatch:\n\twant=%v,\n\t got=%v", expected, actual)
		}
	})

	t.Run("false positive", func(t *testing.T) {
		db := ttcn3.DB{}
		db.Index("file1.ttcn3", "file2.ttcn3", "file3.ttcn3")

		db.Modules["M2"]["file3.ttcn3"] = true // false positiv entry
		db.Names["E"]["file3.ttcn3"] = true    // false positiv entry

		expected := []string{"M2:file1.ttcn3"}
		actual := importedDefs(&db, "E", "M1")
		if !equal(actual, expected) {
			t.Errorf("Mismatch:\n\twant=%v,\n\t got=%v", expected, actual)
		}
	})

}

func TestFindDefinitions(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{`{var integer x := ¶x}`, []string{"x0"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testFindDefinition(t, tt.input, tt.want)
		})
	}
}

func testFindDefinition(t *testing.T, input string, expected []string) {
	// extract cursor position
	cursor := strings.Index(input, CURSOR)
	input = strings.Replace(input, CURSOR, "", 1)

	// parse
	file := fmt.Sprintf("%s.ttcn3", t.Name())
	fs.SetContent(file, []byte(input))
	tree := ttcn3.ParseFile(file)
	ids := idMap(tree.Root)

	// index
	db := ttcn3.DB{}
	db.Index(file)

	actual := []string{}
	n, stack := parentNodes(tree, cursor)
	for _, d := range db.FindDefinitions(n, tree, stack...) {
		actual = append(actual, ids[d.Ident])
	}

	if !equal(actual, expected) {
		t.Errorf("Mismatch:\n\twant=%v,\n\t got=%v", expected, actual)
	}
}
func idMap(root ast.Node) map[*ast.Ident]string {
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
func parentNodes(tree *ttcn3.Tree, cursor int) (n ast.Expr, s []ast.Node) {
	s = tree.SliceAt(loc.Pos(cursor + 1))
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

func importedDefs(db *ttcn3.DB, id string, mod string) []string {
	var s []string
	for _, d := range db.FindImportedDefinitions(id, mod) {
		name := ast.Name(d.Node)
		file := d.Tree.Position(d.Node.Pos()).Filename
		s = append(s, fmt.Sprintf("%s:%s", name, file))
	}
	return s
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
