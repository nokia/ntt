package syntax

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	b := Builder{}

	assert.Panics(t, func() { b.Pop() }, "Calling Pop on an empty stack should panic")

	b.Push(Root)
	b.Push(Module)
	b.PushToken(Integer, 0, 1)
	b.PushToken(Float, 1, 2)
	b.Pop()
	b.PushToken(Identifier, 2, 3)
	b.Pop()

	expected := []string{
		"enter Root: parent=0 skip=6",
		"enter Module: parent=-1 skip=3",
		"integer",
		"float",
		"exit Module: parent=-4 skip=-3",
		"identifier",
		"exit Root: parent=-6 skip=-6",
	}

	assert.Equal(t, expected, printEvents(b.events))
}

func TestNode(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var b Builder
		root := b.Push(Root)
		b.Pop()
		assert.Equal(t, Root, root.Kind())
		assert.Equal(t, Invalid, root.Parent())
		assert.Equal(t, Invalid, root.FirstToken())
		assert.Equal(t, Invalid, root.LastToken())
		assert.Equal(t, Invalid, root.FirstChild())
		assert.Equal(t, Invalid, root.Next())
		assert.Equal(t, -1, root.Pos())
		assert.Equal(t, -1, root.End())
		assert.Equal(t, 0, root.Len())
	})

	t.Run("Tokens", func(t *testing.T) {
		var b Builder
		root := b.Push(Root)
		x := b.PushToken(Add, 4, 5)
		y := b.PushToken(Sub, 5, 6)
		z := b.PushToken(Mul, 7, 8)
		b.Pop()

		assert.Equal(t, Invalid, root.Parent())
		assert.Equal(t, root, x.Parent())
		assert.Equal(t, root, y.Parent())
		assert.Equal(t, root, z.Parent())
		assert.Equal(t, x, x.FirstToken())
		assert.Equal(t, x, root.FirstToken())
		assert.Equal(t, z, root.LastToken())
		assert.Equal(t, x, root.FirstChild())
		assert.Equal(t, Invalid, root.Next())
		assert.Equal(t, 4, root.Pos())
		assert.Equal(t, 8, root.End())
		assert.Equal(t, 4, root.Len())
	})

}

func TestScanner(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		// white spaces and trivia
		{"empty", "", []string{"[0:1): EOF"}},
		{"spaces", "	 ", []string{"[2:3): EOF"}},
		{"comments", "///**/\n/**///", []string{
			"[0:7): comment",
			"[7:11): comment",
			"[11:13): comment",
			"[13:14): EOF"}},
		{"preproc", "#", []string{"[0:1): preprocessor directive", "[1:2): EOF"}},
		{"preproc", "#foo bar\n", []string{"[0:9): preprocessor directive", "[9:10): EOF"}},

		// identifiers
		{"ids", "_ f1o", []string{"[0:1): identifier", "[2:5): identifier", "[5:6): EOF"}},
		{"ids", "%definitionId", []string{"[0:13): identifier", "[13:14): EOF"}},

		// integers
		{"integer", "0", []string{"[0:1): integer", "[1:2): EOF"}},
		{"integer", "00", []string{"[0:2): malformed token", "[2:3): EOF"}},
		{"integer", "0h", []string{"[0:2): malformed token", "[2:3): EOF"}},

		{"modifier", "@foo", []string{"[0:4): modifier", "[4:5): EOF"}},
		{"modifier", "@", []string{"[0:1): unknown token", "[1:2): EOF"}},
		// floats
		{"float", "0.0", []string{"[0:3): float", "[3:4): EOF"}},
		{"float", "0.01", []string{"[0:4): float", "[4:5): EOF"}},
		{"float", "1E2", []string{"[0:3): float", "[3:4): EOF"}},
		{"float", "1e+2", []string{"[0:4): float", "[4:5): EOF"}},
		{"float", "1.2e-2", []string{"[0:6): float", "[6:7): EOF"}},
		{"float", "0.0e", []string{"[0:4): malformed token", "[4:5): EOF"}},
		{"float", "0.0h", []string{"[0:4): malformed token", "[4:5): EOF"}},
		{"float", "0.0e-", []string{"[0:5): malformed token", "[5:6): EOF"}},

		// strings
		{"string", `"foo"`, []string{"[0:5): string", "[5:6): EOF"}},
		{"string", `"`, []string{"[0:1): unterminated string", "[1:2): EOF"}},
		{"string", `""`, []string{"[0:2): string", "[2:3): EOF"}},
		{"string", `"""`, []string{"[0:3): unterminated string", "[3:4): EOF"}},
		{"string", `""""`, []string{"[0:4): string", "[4:5): EOF"}},
		{"string", `"\"`, []string{"[0:3): unterminated string", "[3:4): EOF"}},
		{"string", `"\\"`, []string{"[0:4): string", "[4:5): EOF"}},
		{"string", `"\""`, []string{"[0:4): string", "[4:5): EOF"}},

		// bitstring
		{"bitstring", "'", []string{"[0:1): unterminated string", "[1:2): EOF"}},
		{"bitstring", "''", []string{"[0:2): malformed token", "[2:3): EOF"}},
		{"bitstring", "''4", []string{"[0:3): malformed token", "[3:4): EOF"}},
		{"bitstring", "''b", []string{"[0:3): bitstring", "[3:4): EOF"}},
		{"bitstring", "''hex", []string{"[0:5): bitstring", "[5:6): EOF"}},
		{"bitstring", "'/**/ 'hex", []string{"[0:10): bitstring", "[10:11): EOF"}},
		{"bitstring", "'1?00 0101'B", []string{"[0:12): bitstring", "[12:13): EOF"}},

		// multi character operators
		{"ops", "!=", []string{"[0:2): !=", "[2:3): EOF"}},
		{"ops", "->", []string{"[0:2): ->", "[2:3): EOF"}},
		{"ops", "..", []string{"[0:2): ..", "[2:3): EOF"}},
		{"ops", "::", []string{"[0:2): ::", "[2:3): EOF"}},
		{"ops", ":=", []string{"[0:2): :=", "[2:3): EOF"}},
		{"ops", "<<", []string{"[0:2): <<", "[2:3): EOF"}},
		{"ops", "<=", []string{"[0:2): <=", "[2:3): EOF"}},
		{"ops", "<@", []string{"[0:2): <@", "[2:3): EOF"}},
		{"ops", "==", []string{"[0:2): ==", "[2:3): EOF"}},
		{"ops", "=>", []string{"[0:2): =>", "[2:3): EOF"}},
		{"ops", ">=", []string{"[0:2): >=", "[2:3): EOF"}},
		{"ops", ">>", []string{"[0:2): >>", "[2:3): EOF"}},
		{"ops", "@>", []string{"[0:2): @>", "[2:3): EOF"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, testScan(tt.input))
		})
	}
}

func testScan(input string) []string {
	s := NewScanner([]byte(input))
	var nodes []string
	for {
		k, begin, end := s.Scan()
		nodes = append(nodes, fmt.Sprintf("[%d:%d): %s", begin, end, k.String()))
		if k == EOF {
			break
		}
	}
	return nodes
}

func testParse(input string) []string {
	var nodes []string
	Parse([]byte(input)).Inspect(func(n Node) bool {
		nodes = append(nodes, fmt.Sprintf("%T ", n))
		return true
	})
	return nodes
}

func printEvents(events []treeEvent) []string {
	var ret []string
	for _, e := range events {
		var s string
		switch e.Type() {
		case OpenNode:
			s = fmt.Sprintf("enter %s: parent=%d skip=%d", e.Kind(), e.parent(), e.skip())
		case AddToken:
			s = fmt.Sprintf("%s", e.Kind())
		case CloseNode:
			s = fmt.Sprintf("exit %s: parent=%d skip=%d", e.Kind(), e.parent(), e.skip())
		}

		ret = append(ret, s)
	}
	return ret
}
