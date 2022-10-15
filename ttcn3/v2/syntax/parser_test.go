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
		{"module M {}", N("Root",
			N("Module",
				N("module"),
				N("Name",
					N("M")),
				N("{"),
				N("}")))},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertNode(Parse([]byte(tt.input)))
			assert.Equal(t, tt.want, got)
		})
	}
}
