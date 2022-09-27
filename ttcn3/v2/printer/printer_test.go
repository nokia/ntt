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

		// Remove leading whitespace
		{input: "    leading;", want: "leading;"},

		// Remove trailing whitespace
		{input: "trailing;   ", want: "trailing;"},

		// At max one blank between tokens.
		{input: "import from   all", want: "import from all"},

		// Convert all other whitespace to blanks.
		{input: "import \tfrom\tall", want: "import from all"},
	}

	for _, test := range tests {
		f := &printer{}
		got, err := f.Bytes([]byte(test.input))
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
	}
}
