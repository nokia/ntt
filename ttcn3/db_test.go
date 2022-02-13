package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3"
)

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
		mod := moduleFrom("file1.ttcn3", "M1")
		if defs := db.VisibleModules("E", mod); len(defs) != 0 {
			t.Errorf("Expected 0 definitions, got %v", defs)
		}
	})

	t.Run("regular", func(t *testing.T) {
		db := ttcn3.DB{}
		db.Index("file1.ttcn3", "file2.ttcn3", "file3.ttcn3")

		expected := []string{"M2:file1.ttcn3", "M1:file1.ttcn3", "M1:file2.ttcn3"}
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

		expected := []string{"M2:file1.ttcn3", "M1:file1.ttcn3", "M1:file2.ttcn3"}
		actual := importedDefs(&db, "E", "M1")
		if !equal(actual, expected) {
			t.Errorf("Mismatch:\n\twant=%v,\n\t got=%v", expected, actual)
		}
	})

}

func TestLookup(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple",
			input: `module M {var integer x := ¶x}`,
			want:  []string{"x0"}},
		{
			name:  "duplicates",
			input: `module x {type x ¶x}`,
			want:  []string{"x0", "x2"}},
		{
			name: "multi",
			input: `module M {type integer x}
				module M {type x      ¶x}`,
			want: []string{"x0", "x2"}},
		{
			name: "imports",
			input: `module M2 {type integer x}
				module M1 {type ¶x y}`,
			want: []string{}},
		{
			name: "imports",
			input: `module M2 {type integer x}
				module M1 {type ¶x y}
				module M1 {import from M2 all}`,
			want: []string{}},
		{
			name: "imports",
			input: `module M2 {type integer x}
				module M1 {type ¶x y; import from M2 all}`,
			want: []string{"x0"}},
		{
			name:  "imports",
			input: `module M1 {type ¶x integer; import from x all}`,
			want:  []string{"x1"}},
		{
			name:  "imports",
			input: `module M2 {} module M1 {import from ¶M2 all}`,
			want:  []string{"M20", "M21"}},
		{
			name: "imports",
			input: `module M1 {type ¶x integer}
				module M1 {import from x all}`,
			want: []string{}},
		{
			name:  "imports",
			input: `module x {type ¶x integer; import from x all}`,
			want:  []string{"x0", "x2"}},
		{
			name:  "imports",
			input: `module x {} module x {import from x all; var integer x := ¶x}`,
			want:  []string{"x0", "x1", "x2", "x3"}},
		{
			name:  "dot",
			input: `module M {var integer x := ¶M.x}`,
			want:  []string{"M0"}},
		{
			name:  "dot",
			input: `module M {var integer x := M.¶x}`,
			want:  []string{"x0"}},
		{
			name:  "dot",
			input: `module M {type record R {int x}; type R.¶x x}`,
			want:  []string{"x0"}},
		{
			name:  "dot",
			input: `module R {type record R {int x}; type R.¶x x}`,
			want:  []string{"x0", "x2"}},
		{
			name:  "dot",
			input: `module R {type record R {int x}; type R.R.¶x x}`,
			want:  []string{"x0"}},
		{
			name: "dot imports",
			input: `module M2 {type record R {int x}}
				module M1 {type R.¶x x; import from M2 all}`,
			want: []string{"x0"}},
		{
			name: "dot imports",
			input: `module M2 {type record R {int x}}
				module R  {type M2.R.¶x x; import from M2 all}`,
			want: []string{"x0"}},
		{
			name: "dot ambigious",
			input: `module M2 {type record R {int x}}
				module R  {type R.¶x x; import from M2 all}`,
			want: []string{"x0", "x2"}},
		{
			name:  "typeref",
			input: `module M { type record R {int x}; control { var M.R r; r.¶x := 23 }}`,
			want:  []string{"x0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cursor loc.Pos
			cursor, tt.input = extractCursor(tt.input)

			tree := parseFile(t, t.Name(), tt.input)
			ids := enumerateIDs(tree.Root)

			db := &ttcn3.DB{}
			db.Index(tree.Filename())

			actual := []string{}
			n := tree.ExprAt(cursor)
			for _, d := range tree.LookupWithDB(n, db) {
				actual = append(actual, ids[d.Ident])
			}

			if !equal(actual, tt.want) {
				t.Errorf("Mismatch:\n\twant=%v,\n\t got=%v", tt.want, actual)
			}
		})
	}
}
