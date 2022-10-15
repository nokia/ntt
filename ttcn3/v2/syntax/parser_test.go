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
		{"module M {}", N("Root",
			N("Module",
				N("module"),
				N("Name",
					N("M")),
				N("{"),
				N("}")))},

		{"/*1*/ module /*2*/ M /*3*/ { /*4*/ } /*5*/", N("Root",
			N("Module",
				N("/*1*/"),
				N("module"),
				N("Name",
					N("/*2*/"),
					N("M")),
				N("/*3*/"),
				N("{"),
				N("/*4*/"),
				N("}")),
			N("/*5*/"))},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertNode(Parse([]byte(tt.input)))
			assert.Equal(t, tt.want, got)
		})
	}
}
