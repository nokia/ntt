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
		{"", ""},
	}

	for _, test := range tests {
		f := &simpleFormatter{}
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
