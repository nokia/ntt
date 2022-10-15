package syntax

import "strings"

const CURSOR = "¶"

// A TestNode is a simplified syntax node, which is easier to test.
type TestNode struct {
	Kind     string
	Children []TestNode
}

// N is a helper function to create a TestNode. First argument is the node kind
// and the remaining arguments are its children (like a S-expression).
func N(k string, c ...TestNode) TestNode {
	return TestNode{
		Kind:     k,
		Children: c,
	}
}

// Convert a Node tree into a TestNode tree.
func ConvertNode(n Node) TestNode {
	var children []TestNode
	for c := n.FirstChild(); c != Nil; c = c.Next() {
		children = append(children, ConvertNode(c))
	}
	k := n.Kind().String()
	if n.Kind().IsLiteral() {
		k = n.Text()
	}
	return TestNode{
		Kind:     k,
		Children: children,
	}
}

// ExtractCursor removes the cursor from the given string and returns the
// string and the cursor position. If the cursor is not found, the cursor
// position is -1.
func ExtractCursor(s string) (string, int) {
	return strings.Replace(s, CURSOR, "", -1), strings.Index(s, CURSOR)
}
