package token

import "testing"

func TestUnquote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`""`, ``},
		{`""""`, `"`},
		{`"\""`, `"`},
		{`"\n"`, "\n"},
		{`"\t"`, "\t"},
		{`"Hello World!"`, `Hello World!`},
	}
	for _, tt := range tests {
		actual, err := Unquote(tt.input)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		if actual != tt.expected {
			t.Errorf("got=%v, want=%v", actual, tt.expected)
		}
	}

}
