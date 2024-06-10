package t3xf_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/k3/t3xf"
	"github.com/nokia/ntt/k3/t3xf/opcode"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/stretchr/testify/assert"
)

func TestCompiler(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{input: "1", want: []string{
			"natlong 1",
		}},

		{input: "1+2*3", want: []string{
			"natlong 1",
			"natlong 2",
			"natlong 3",
			"mul",
			"add",
		}},

		{input: "var integer a := 1", want: []string{
			"integer",
			"name a",
			"var",
			"natlong 1",
			"ref 20",
			"assign",
		}},

		{input: "var integer a[1][2]", want: []string{
			"integer",
			"scan",
			"natlong 2",
			"block",
			"array",
			"scan",
			"natlong 1",
			"block",
			"array",
			"name a",
			"var",
		}},

		{input: "const integer a := 1, b := 2", want: []string{
			"scan",
			"natlong 1",
			"block",
			"integer",
			"name a",
			"const",
			"scan",
			"natlong 2",
			"block",
			"integer",
			"name b",
			"const",
		}},

		{input: `module M {} with { extension "e" }`, want: []string{
			"scan",
			"utf8 e",
			"extension",
			"block",
			"scan",
			"block",
			"name M",
			"modulew",
		}},

		{input: `function f() {}`, want: []string{
			"skip",
			"skip",
			"name f",
			"function",
		}},

		{input: `function f(integer x, inout integer y, out template integer z) {}`, want: []string{
			"scan",
			"integer",
			"name x",
			"in",
			"integer",
			"name y",
			"inout",
			"integer",
			"permitt",
			"name z",
			"out",
			"block",
			"skip",
			"name f",
			"function",
		}},

		{input: `function f() return integer {}`, want: []string{
			"integer",
			"skip",
			"skip",
			"name f",
			"functionv",
		}},

		{input: `function f() return template integer {}`, want: []string{
			"integer",
			"permitt",
			"skip",
			"skip",
			"name f",
			"functionv",
		}},

		{input: `control {}`, want: []string{
			"skip",
			"control",
		}},

		{input: `type charstring T ("a".."z")`, want: []string{
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"subtype",
			"name T",
			"type",
		}},

		{input: `type charstring T length(2)`, want: []string{
			"charstring",
			"any",
			"natlong 2",
			"length",
			"subtype",
			"name T",
			"type",
		}},

		{input: `type charstring T ("a".."z") length(2)`, want: []string{
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			"name T",
			"type",
		}},

		{input: `type record of charstring T ("a".."z") length(2)`, want: []string{
			//"scan",
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			//"block",
			"recordof",
			"name T",
			"type",
		}},

		{input: `type record length(1) of charstring T ("a".."z") length(2)`, want: []string{
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			"recordof",
			"any",
			"natlong 1",
			"length",
			"subtype",
			"name T",
			"type",
		}},

		{input: `type record T { charstring F ("a".."z") }`, want: []string{
			"scan",
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"subtype",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},

		{input: `type record T { charstring F length(2) }`, want: []string{
			"scan",
			"charstring",
			"any",
			"natlong 2",
			"length",
			"subtype",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},

		{input: `type record T { charstring F ("a".."z") length(2) }`, want: []string{
			"scan",
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},

		{input: `type record T { record of charstring F ("a".."z") length(2) }`, want: []string{
			"scan",
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			"recordof",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},

		{input: `type record T { record length(1) of charstring F ("a".."z") length(2) }`, want: []string{
			"scan",
			"charstring",
			"utf8 a",
			"utf8 z",
			"range",
			"natlong 2",
			"length",
			"subtype",
			"recordof",
			"any",
			"natlong 1",
			"length",
			"subtype",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},

		{input: `type record T { record {} F ({},{}) length(2) }`, want: []string{
			"scan",
			"skip",
			"record",
			"mark",
			"mark",
			"vlist",
			"mark",
			"vlist",
			"collect",
			"natlong 2",
			"length",
			"subtype",
			"name F",
			"field",
			"block",
			"record",
			"name T",
			"type",
		}},
	}

	for _, tt := range tests {
		root, _, _ := syntax.Parse([]byte(tt.input))
		if root.Err() != nil {
			t.Fatalf("syntax.Parse(%q) returned error: %v", tt.input, root.Err())
		}

		c := t3xf.NewCompiler()
		c.Compile(root)
		b, err := c.Assemble()
		if err != nil {
			t.Fatalf("c.Assemble() returned error: %v", err)
		}
		got := []string{}
		pc := 0
		for pc < len(b) {
			n, op, arg := t3xf.Decode(b[pc:])
			pc += n
			if op == opcode.LINE {
				continue
			}
			s := op.String()
			if arg != nil {
				s += " " + fmt.Sprintf("%v", arg)
			}
			got = append(got, s)
		}
		assert.Equal(t, tt.want, got, tt.input)
	}
}
