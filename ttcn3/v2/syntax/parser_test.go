package syntax

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	tests := []struct {
		input string
		want  TestNode
	}{
		{"", N("Root")},
		{"//", N("Root", N("//"))},
		{"module M {};", N("Root",
			N("Module",
				N("module"),
				N("M"),
				N("{"),
				N("}"),
				N(";")))},

		{"/*1*/ module /*2*/ M /*3*/ { /*4*/ }; /*5*/", N("Root",
			N("Module",
				N("/*1*/"),
				N("module"),
				N("/*2*/"),
				N("M"),
				N("/*3*/"),
				N("{"),
				N("/*4*/"),
				N("}"),
				N(";")),
			N("/*5*/"))},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertNode(Parse([]byte(tt.input)))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNameLiteral(t *testing.T) {
	input := "var integer a := b, x := y"
	var names, ids []string
	Parse([]byte(input)).Inspect(func(n Node) bool {
		if n.Kind() == Name {
			names = append(names, n.Text())
		}
		if n.Kind() == Identifier {
			ids = append(ids, n.Text())
		}
		return true
	})
	assert.Equal(t, []string{"a", "x"}, names)
	assert.Equal(t, []string{"integer", "b", "y"}, ids)
}

func TestComma(t *testing.T) {
	tests := []struct {
		input   string
		noError bool
	}{
		{"type record R { int a, int b }", true},
		{"type record R { int a, int b, }", true},
		{"type record R { int a, int b; }", false},
		{"type record R { int a int b }", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			root := Parse([]byte(tt.input))
			if err := root.Err(); tt.noError {
				assert.NoError(t, err)
			} else {
				t.Logf("%v", root.tree)
				assert.Error(t, err)
			}
		})
	}

}
