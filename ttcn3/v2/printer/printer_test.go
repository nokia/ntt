package printer

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleFormatter(t *testing.T) {
	tests := []struct {
		input string
		want  interface{}
		skip  bool
	}{
		{input: "", want: ""},
		{input: "foo;", want: "foo;\n"},
		{input: "foo;\n\n", want: "foo;\n"},
		{input: "//foo", want: "//foo\n"},
		{input: "//foo\n", want: "//foo\n"},

		// Remove leading whitespace
		{input: "    leading;", want: "leading;\n"},

		// Remove trailing whitespace
		{input: "trailing;   ", want: "trailing;\n"},

		// At max one blank between tokens.
		{input: "import from   all", want: "import from all\n"},

		// Convert all other whitespace to blanks.
		{input: "import \tfrom\tall", want: "import from all\n"},

		// Replace line breaks with \n
		{input: "foo;\r\nbar;", want: "foo;\nbar;\n"},
		{input: "foo;\rbar;", want: "foo; bar;\n"}, // \r is not a line break
		{input: "foo;\n\rbar;", want: "foo;\nbar;\n"},
		{input: "foo;\vbar;", want: "foo;\nbar;\n"},
		{input: "foo;\fbar;", want: "foo;\nbar;\n"},

		// Keep at most one newline
		{input: "foo;\r\n\r\nbar;", want: "foo;\n\nbar;\n"},

		// User defined spaces
		{input: "control{}", want: "control{}\n"},
		{input: "control {}", want: "control {}\n"},
		{input: "control {} // Foo", want: "control {} // Foo\n"},
		{input: "control  {}  ", want: "control {}\n"},
		{input: "control \n{}", want: "control\n{}\n"},
		{input: "control\n {}", want: "control\n{}\n"},

		// Verify that {, [ and ( increment the indentation level
		{input: "{foo", want: "{foo\n"},
		{input: "{\nfoo", want: "{\n\tfoo\n"},
		{input: "{\n foo", want: "{\n\tfoo\n"},
		{input: "{\nfoo}\nbar", want: "{\n\tfoo}\nbar\n"},
		{input: "{\n[\n(\n1,2\n)\n]\n}", want: "{\n\t[\n\t\t(\n\t\t\t1,2\n\t\t)\n\t]\n}\n"},

		// Verify that tokens with newlines have correct indentation
		{input: "{// Foo\nBar", want: "{ // Foo\n\tBar\n"},              //  Bar must be indented.
		{input: "{\n/*\n* foo\n*/", want: "{\n\t/*\n\t * foo\n\t */\n"}, //  Comment must be indented, with one extra space.

		// Verify that comments and := are aligned.
		{input: "{x := 1,\nx2:= 123}", want: "{x := 1,\n\tx2 := 123}\n"},
		{input: "{\nx := 1,\nx2:= 123}", want: "{\n\tx  := 1,\n\tx2 := 123}\n"},
		{input: "{\nx := 1, // a\nx2:= 123 /* b */}", want: "{\n\tx  := 1,  // a\n\tx2 := 123 /* b */}\n"},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			if test.skip {
				t.Skip()
			}
			var buf bytes.Buffer
			err := Fprint(&buf, test.input)
			got := buf.String()
			switch want := test.want.(type) {
			case string:
				assert.Nil(t, err)
				assert.Equal(t, want, got)
			case error:
				assert.True(t, errors.Is(want, err))
				assert.Nil(t, got)
			default:
				t.Fatalf("test implementation error: unexpected type %T", want)
			}
		})
	}
}
