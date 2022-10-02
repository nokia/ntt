package printer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleFormatter(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
	}{
		{input: "", want: ""},
		{input: "foo;", want: "foo;"},
		{input: "//foo", want: "//foo"},
		{input: "//foo\n", want: "//foo\n"},

		// Remove leading whitespace
		{input: "    leading;", want: "leading;"},

		// Remove trailing whitespace
		{input: "trailing;   ", want: "trailing;"},

		// At max one blank between tokens.
		{input: "import from   all", want: "import from all"},

		// Convert all other whitespace to blanks.
		{input: "import \tfrom\tall", want: "import from all"},

		// Replace line breaks with \n
		{input: "foo;\r\nbar;", want: "foo;\nbar;"},
		{input: "foo;\rbar;", want: "foo; bar;"}, // \r is not a line break
		{input: "foo;\n\rbar;", want: "foo;\nbar;"},
		{input: "foo;\vbar;", want: "foo;\nbar;"},
		{input: "foo;\fbar;", want: "foo;\nbar;"},

		// Keep at most one newline
		{input: "foo;\r\n\r\nbar;", want: "foo;\n\nbar;"},

		// User defined spaces
		{input: "control{}", want: "control{}"},
		{input: "control {}", want: "control {}"},
		{input: "control {} // Foo", want: "control {} // Foo"},
		{input: "control  {}  ", want: "control {}"},
		{input: "control \n{}", want: "control\n{}"},
		{input: "control\n {}", want: "control\n{}"},

		// Verify that {,[ and ( increment the indentation level
		{input: "{foo", want: "{foo"},
		{input: "{\nfoo", want: "{\n\tfoo"},
		{input: "{\n foo", want: "{\n\tfoo"},
		{input: "{\nfoo}\nbar", want: "{\n\tfoo}\nbar"},
		{input: "{\n[\n(\n1,2\n)\n]\n}", want: "{\n\t[\n\t\t(\n\t\t\t1,2\n\t\t)\n\t]\n}"},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			t.Logf("input: %q", test.input)
			got, err := Bytes([]byte(test.input))
			switch want := test.want.(type) {
			case string:
				assert.Nil(t, err)
				assert.Equal(t, want, string(got))
			case error:
				assert.True(t, errors.Is(want, err))
				assert.Nil(t, got)
			default:
				t.Fatalf("test implementation error: unexpected type %T", want)
			}
		})
	}
}
