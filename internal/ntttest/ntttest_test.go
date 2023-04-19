package ntttest_test

import (
	"testing"

	"github.com/nokia/ntt/internal/ntttest"
)

func TestCutCursor(t *testing.T) {
	tests := []struct {
		input string
		pos   int
		text  string
	}{
		{"", -1, ""},
		{"¶", 0, ""},
		{"text", -1, "text"},
		{"¶text", 0, "text"},
		{"text¶", 4, "text"},
		{"te¶xt", 2, "text"},
		{"te¶xt¶", 2, "text¶"},
	}

	for _, test := range tests {
		text, pos := ntttest.CutCursor(test.input)
		if pos != test.pos {
			t.Errorf("ExtractCursor(%q) pos = %d, want %d", test.input, pos, test.pos)
		}
		if text != test.text {
			t.Errorf("ExtractCursor(%q) text = %q, want %q", test.input, text, test.text)
		}
	}
}
