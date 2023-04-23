package syntax_test

import (
	"testing"

	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/stretchr/testify/assert"
)

func TestDoc(t *testing.T) {
	t.Parallel()

	testDoc := func(t *testing.T, input string) string {
		root, _, _, err := syntax.Parse(t.Name(), input)
		if err != nil {
			t.Fatal(err)
		}
		return syntax.Doc(root)
	}

	t.Run("empty", func(t *testing.T) {
		input := `module Foo {}`
		want := ""
		assert.Equal(t, want, testDoc(t, input))
	})

	t.Run("comments", func(t *testing.T) {
		input := `
		// foo
		// bar
		module Foo {}`
		want := "foo\nbar\n"
		assert.Equal(t, want, testDoc(t, input))
	})

	t.Run("comments", func(t *testing.T) {
		input := `
		// foo
		// bar

		module Foo {}`
		want := ""
		got := testDoc(t, input)
		assert.Equal(t, want, got)
	})

	t.Run("comments", func(t *testing.T) {
		input := `
		// foo

		// bar
		module Foo {}`
		want := "bar\n"
		assert.Equal(t, want, testDoc(t, input))
	})

	t.Run("comments", func(t *testing.T) {
		input := `
		/* foo */
		/* bar */
		module Foo {}`
		want := "foo\nbar\n"
		assert.Equal(t, want, testDoc(t, input))
	})

	t.Run("comments", func(t *testing.T) {
		input := `
		/* foo */ /* bar */
		module Foo {}`
		want := "foo bar\n"
		assert.Equal(t, want, testDoc(t, input))
	})
}
