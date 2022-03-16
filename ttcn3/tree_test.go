package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3"
	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: `module M2 {} module M1 {import from ¶M2 all}`,
			want: []string{
				"*ast.Ident(M2)",
				"*ast.ImportDecl",
				"*ast.ModuleDef",
				"*ast.Module(M1)",
				"*ast.NodeList",
			}},
		{

			input: `module M {
				  function func<type T>(T x) {
				    while (true) { ¶T := x; }
				  }
		  		}`,
			want: []string{
				"*ast.Ident(T)",
				"*ast.BinaryExpr",
				"*ast.ExprStmt",
				"*ast.BlockStmt",
				"*ast.WhileStmt",
				"*ast.BlockStmt",
				"*ast.FuncDecl(func)",
				"*ast.ModuleDef(func)",
				"*ast.Module(M)",
				"*ast.NodeList",
			}},
	}

	for _, tt := range tests {
		cursor, source := extractCursor(tt.input)
		tree := parseFile(t, t.Name(), source)

		var actual []string
		for _, n := range tree.SliceAt(cursor) {
			actual = append(actual, nodeDesc(n))
		}

		assert.Equal(t, tt.want, actual)
	}
}

func TestExprAt(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "<nil>"},
		{"x", "<nil>"},
		{"¶", "<nil>"},
		{"x¶", "<nil>"},
		{"¶x", "*ast.Ident(x)"},
		{"¶x[-]", "*ast.Ident(x)"},
		{"¶x[-][-]", "*ast.Ident(x)"},
		{"¶x()[-]", "*ast.Ident(x)"},
		{"¶x.y.z", "*ast.Ident(x)"},
		{"x¶.y.z", "<nil>"},
		{"x.¶y.z", "*ast.SelectorExpr(x.y)"},
		{"x.y.¶z", "*ast.SelectorExpr(x.y.z)"},
		{"x[-].¶y.z", "*ast.SelectorExpr(.y)"},
		{"foo(23, ¶x:= 1)", "*ast.Ident(x)"},
		{"a := { ¶x:= 1}", "*ast.Ident(x)"},
		{"a := { [¶x]:= 1}", "*ast.Ident(x)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cursor, source := extractCursor(tt.input)
			tree := parseFile(t, t.Name(), source)
			actual := nodeDesc(tree.ExprAt(cursor))
			assert.Equal(t, tt.want, actual)
		})
	}
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
			name: "dot",
			input: `module R {
					type record R {int x}
					type R Foo;
					var Foo f := f.¶x;
				}`,
			want: []string{"x0"}},
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
		{
			name:  "index",
			input: `module M { type record of record {int x} R; control { var M.R r; r.¶x := 23 }}`,
			want:  []string{}},
		{
			name:  "index",
			input: `module M { type record of record {int x} R; control { var M.R r; r[-].¶x := 23 }}`,
			want:  []string{"x0"}},
		{
			name:  "index",
			input: `module M { type record of record {int x} R; control { var M.R[-] r; r.¶x := 23 }}`,
			want:  []string{"x0"}},
		{
			name:  "index",
			input: `module M { type record R {int x}; control { var M.R r[2]; r[0].¶x := 23 }}`,
			want:  []string{"x0"}},
		{
			name: "call",
			input: `module M {
				    type record R {int x};
				    function f() return R {}
				    control { f().¶x := 23 }}`,
			want: []string{"x0"}},
		{
			name: "call",
			input: `module M {
				    type record R {int x};
				    type function F() return R;
				    function f() return F {}
				    control { f()().¶x := 23 }}`,
			want: []string{"x0"}},

		{
			name:  "components",
			input: `module M {type component C { var int x; }; function f() runs on C { ¶x }}`,
			want:  []string{"x0"}},
		{
			name: "components",
			input: `module M {
			           type component C extends D, E {}
			           type component D { var int x; }
			           type component E { var int x; };
				   function f(int x) runs on C { while (true) {¶x} }
				}`,
			want: []string{"x0", "x1", "x2"}},
		{
			name: "components",
			input: `module M {
				    type record R {int x}
			            type component C {var R r}
				    function f() runs on C { while (true) {r.¶x} }
			    	}`,
			want: []string{"x0"}},
		{
			name: "components",
			input: `module M {
				    type record R {int x}
				    function f() runs on R { while (true) {¶x} }
				}`,
			want: []string{"x0"}},
		{
			name: "parameters",
			input: `module M {
					const int x := 1;
					function f(int x) {}
					signature f(int x);
					template int f(int x) := 23;
					type function f(int x);
					control { f(¶x := 23)}
				}`,
			want: []string{"x1", "x2", "x3", "x4"}},
		{
			name: "parameters",
			input: `module M {
					type record R { FuncType f }
					type function FuncType(int x);
					control { var R r; r.f(¶x := 23)}
				}`,
			want: []string{"x0"}},
		{
			name: "parameters",
			input: `module M {
					type record R { record of FuncType f }
					type function FuncType(int x);
					control { var R r; r.f[0](¶x := 23)}
				}`,
			want: []string{"x0"}},
		{
			name: "field assignment",
			input: `module M {
					const int x := 1;
					type record R { int x }
					template R t(int x) := { ¶x := x }
				}`,
			want: []string{"x1"}},
		{
			name: "field assignment",
			input: `module M {
					const int x := 1;
					type record R { int x }
					type record S { int x }
					control {
						var R x;
						x := { ¶x := x }
						var S x;
					}
				}`,
			want: []string{"x1", "x2"}},
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
