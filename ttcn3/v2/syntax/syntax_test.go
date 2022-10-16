package syntax

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuilder verifies that the builder produces the expected tree events.
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

func TestInspect(t *testing.T) {
	b := Builder{}
	root := b.Push(Root)
	b.Push(Module)
	b.PushToken(ModuleKeyword, 0, 1)
	b.PushToken(Name, 1, 2)
	b.PushToken(LeftBrace, 2, 3)
	b.PushToken(RightBrace, 3, 4)
	b.Pop()
	b.Pop()

	var got []string
	root.Inspect(func(n Node) bool {
		switch {
		case n.IsNonTerminal():
			got = append(got, fmt.Sprintf("%d: push %s", n.idx, n.Kind()))
		case n.IsTerminal():
			got = append(got, fmt.Sprintf("%d: add '%s'", n.idx, n.Kind()))
		case n == Nil:
			got = append(got, fmt.Sprintf("x: pop"))
		default:
			t.Errorf("unexpected node: %v", n)
		}
		return true
	})
	want := []string{
		"0: push Root",
		"1: push Module",
		"2: add 'module'",
		"3: add 'name'",
		"4: add '{'",
		"5: add '}'",
		"x: pop",
		"x: pop",
	}
	assert.Equal(t, want, got)
}

// TestNodeAPI verifies that the node methods for navgating the tree works.
func TestNodeAPI(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		var b Builder
		root := b.Push(Root)
		b.Pop()
		assert.Equal(t, Root, root.Kind())
		assert.Equal(t, Nil, root.Parent())
		assert.Equal(t, Nil, root.FirstToken())
		assert.Equal(t, Nil, root.LastToken())
		assert.Equal(t, Nil, root.FirstChild())
		assert.Equal(t, Nil, root.Next())
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

		assert.Equal(t, Nil, root.Parent())
		assert.Equal(t, root, x.Parent())
		assert.Equal(t, root, y.Parent())
		assert.Equal(t, root, z.Parent())
		assert.Equal(t, x, x.FirstToken())
		assert.Equal(t, x, root.FirstToken())
		assert.Equal(t, z, root.LastToken())
		assert.Equal(t, x, root.FirstChild())
		assert.Equal(t, Nil, root.Next())
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

func TestFindDescendant(t *testing.T) {
	tests := []struct {
		input string
		want  Kind
	}{
		{"", EOF},
		{"¶", EOF},
		{"Foo¶", EOF},
		{"Fo¶0", Identifier},
		{"12+ab", EOF},
		{"12¶+ab", Add},
		{"12¶ +ab", Root},
		{"12+ab¶", EOF},
	}

	for _, tt := range tests {
		src, pos := ExtractCursor(tt.input)
		// TODO(5nord): Use parser instead of tokenizer.
		k := Tokenize([]byte(src)).FindDescendant(pos).Kind()
		assert.Equal(t, fmt.Sprint(tt.want), fmt.Sprint(k))
	}
}

func TestSpans(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		// Only tokens carry file positions. No tokens, no span.
		{"", []string{"Root: -"}},
		{" ", []string{"Root: -"}},
		{"\n\n", []string{"Root: -"}},

		{"1", []string{"Root: 1:1-1:2", "integer: 1:1-1:2"}},
		{"foo //bar", []string{
			"Root: 1:1-1:10",
			"identifier: 1:1-1:4",
			"comment: 1:5-1:10",
		}},
		{"/*\nfoo*", []string{"Root: 1:1-2:5", "unterminated string: 1:1-2:5"}},
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			var got []string
			// TODO(5nord): Use parser instead of tokenizer.
			Tokenize([]byte(tt.input)).Inspect(func(n Node) bool {
				if n.IsValid() {
					s := fmt.Sprintf("%v: %v", n.Kind(), n.Span())
					got = append(got, s)
				}
				return true
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

var sinkSpan Span

func BenchmarkSpansRandomAccess(b *testing.B) {
	tree := benchBuildTree(1000)
	rand.Seed(1)
	var nodes []Node
	tree.Inspect(func(n Node) bool {
		if n.IsValid() {
			nodes = append(nodes, benchRandomNode(tree.tree))
		}
		return true
	})
	b.ResetTimer()
	var r Span
	for i := 0; i < b.N; i++ {
		for _, n := range nodes {
			r = n.Span()
		}
	}
	sinkSpan = r
}

func BenchmarkSpansTreeInspect(b *testing.B) {
	tree := benchBuildTree(1000)
	var nodes []Node
	tree.Inspect(func(n Node) bool {
		if n.IsValid() {
			nodes = append(nodes, n)
		}
		return true
	})
	b.ResetTimer()
	var r Span
	for i := 0; i < b.N; i++ {
		for _, n := range nodes {
			r = n.Span()
		}
	}
	sinkSpan = r
}

func BenchmarkSpansTokenList(b *testing.B) {
	tree := benchBuildTree(1000)
	var nodes []Node
	tree.Inspect(func(n Node) bool {
		if n.IsToken() {
			nodes = append(nodes, n)
		}
		return true
	})
	b.ResetTimer()
	var r Span
	for i := 0; i < b.N; i++ {
		for _, n := range nodes {
			r = n.Span()
		}
	}
	sinkSpan = r
}

func benchRandomNode(tree *tree) Node {
	pos := rand.Intn(len(tree.events) - 1)
	for i, e := range tree.events[pos:] {
		if e.Type() != CloseNode {
			return Node{i, tree}
		}
	}
	return Nil
}

func benchBuildTree(n int) Node {
	var ints []string
	for i := 0; i < n; i++ {
		ints = append(ints, strconv.Itoa(i))
	}
	return Tokenize([]byte(strings.Join(ints, "+")))
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

func printEvents(events []event) []string {
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
