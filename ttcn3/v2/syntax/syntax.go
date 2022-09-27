// Package syntax provides a fast high fidelity parser for TTCN-3.
package syntax

//go:generate go run ./internal/gen

import "fmt"

var (
	Nil = Node{}
)

// Node is a syntax node.
type Node struct {
	idx int
	*tree
}

// Kind returns the syntax node kind of the node, such as Comment, Identifier,
// VarDecl, etc.
func (n Node) Kind() Kind {
	return n.event().Kind()
}

// IsTerminal returns true if the node is a terminal node.
func (n Node) IsTerminal() bool {
	return n.Kind().IsTerminal()
}

// IsNonTerminal returns true if the node is a non-terminal node.
func (n Node) IsNonTerminal() bool {
	return n.Kind().IsNonTerminal()
}

// Parent returns the parent of the node or Invalid if the node is the root node.
func (n Node) Parent() Node {
	if n.idx == 0 {
		return Nil
	}
	switch te := n.event(); te.Type() {

	// If the node is a non-terminal, we know its parent and return.
	case OpenNode:
		return n.get(n.idx + te.parent())

	// If the node is a token, we have to scan for non-terminals.
	case AddToken:
		for i := n.idx + 1; i < len(n.tree.events); i++ {
			switch te := n.tree.events[i]; te.Type() {

			// If the next non-terminal is a CloseNode, we've found the parent.
			case CloseNode:
				return n.get(i + te.skip())

			// If the next non-terminal is a OpenNode, its a child like us and we return its parent.
			case OpenNode:
				return n.get(i + te.parent())
			}
		}

	default:
		panicEvent(te)
	}

	return Nil
}

// FirstToken returns the first token of a syntax node which is not a trivia or Invalid if
// there is none.
//
// Trivia tokens are tokens that are comments or preproccessor directives.
func (n Node) FirstToken() Node {
	switch te := n.event(); te.Type() {

	// If the node is a non-terminal, we iterate over the children
	// and return the first non-terminal.
	case OpenNode:
		for i := n.idx + 1; i < te.skip(); i++ {
			if te := n.tree.events[i]; te.Type() == AddToken && !te.Kind().IsTrivia() {
				return n.get(i)
			}
		}
		return Nil

	// If the node is already a token, we check for "trivianess" and return
	case AddToken:
		if te.Kind().IsTrivia() {
			return Nil
		}
		return n
	default:
		panicEvent(te)
	}
	return Nil
}

// LastToken returns the last token of a syntax node which is not a trivia or Invalid if none is available.
func (n Node) LastToken() Node {
	switch te := n.event(); te.Type() {
	case OpenNode:
		for i := te.skip() - 1; i >= n.idx; i-- {
			if te := n.tree.events[i]; te.Type() == AddToken && !te.Kind().IsTrivia() {
				return n.get(i)
			}
		}
		return Nil
	case AddToken:
		if te.Kind().IsTrivia() {
			return Nil
		}
		return n
	default:
		panicEvent(te)
	}
	return Nil
}

// FirstChild returns the first child of the node or nil if there is none.
func (n Node) FirstChild() Node {
	switch te := n.event(); te.Type() {
	case OpenNode:
		if c := n.get(n.idx + 1); c.event().Type() != CloseNode {
			return c
		}
		return Nil
	case AddToken:
		return Nil
	default:
		panicEvent(te)
	}
	return Nil
}

// Next returns the next sibling node or Invalid if there is none.
func (n Node) Next() Node {
	switch te := n.event(); te.Type() {
	case OpenNode:
		if i := te.skip() + 1; i < len(n.tree.events) && n.tree.events[i].Type() != CloseNode {
			return n.get(i)
		}
		return Nil

	case AddToken:
		if i := n.idx + 1; n.tree.events[i].Type() != CloseNode {
			return n.get(i)
		}
		return Nil
	default:
		panicEvent(te)
	}
	return Nil
}

// Inspect traverses the syntax tree in depth-first order. It calls f for each
// node recursively followed by a call of f(Invalid).
func (n Node) Inspect(f func(n Node) bool) {
	if !f(n) {
		return
	}
	for c := n.FirstChild(); c != Nil; c = c.Next() {
		c.Inspect(f)
	}
	f(Nil)
}

// Pos returns the position (offset) of the node in the source code. Or -1 if
// no position was available.
func (n Node) Pos() int {
	if tok := n.FirstToken(); tok != Nil {
		return int(tok.event().offset())
	}
	return -1
}

// End returns the end position of the node in the source code. Or -1 if none
// was available.
func (n Node) End() int {
	if tok := n.LastToken(); tok != Nil {
		return int(tok.event().offset() + tok.event().length())
	}
	return -1
}

// Len returns the length of the node or 0 if no length is available.
func (n Node) Len() int {
	if pos, end := n.Pos(), n.End(); pos != -1 && end != -1 {
		return end - pos
	}
	return 0
}

// Text returns the source code text of the node.
func (n Node) Text() string {
	if pos, end := n.Pos(), n.End(); pos != -1 && end != -1 {
		return string(n.tree.content[pos:end])
	}
	return ""
}

// Err returns the error of the subtree.
func (n Node) Err() error {
	return nil
}

func (n Node) event() treeEvent {
	return n.tree.events[n.idx]
}

func (n Node) get(idx int) Node {
	return Node{idx: idx, tree: n.tree}
}

// tree represents a parsed TTCN-3 source file.
type tree struct {
	name    string
	events  []treeEvent
	lines   []uint32
	content []byte
	errs    []error
}

// treeEvent represents a single event in a Tree.
//
// treeEvent is an opaque union and not an interface for performance reasons.
// Only two properties are exported (Type and Kind). The other properties are
// context-dependent and are accessable via the Node wrappers.
type treeEvent struct {
	kind  uint16
	flags uint16
	data  [2]int32
}

// treeEventType is the type of a treeEvent.
type treeEventType int

const (
	AddToken treeEventType = iota
	OpenNode
	CloseNode
)

// Create a new non terminal node
func node(kind Kind, children ...treeEvent) []treeEvent {
	var events []treeEvent
	events = append(events, treeEvent{kind: uint16(kind)})
	for _, c := range children {
		if c.Type() == OpenNode || c.Type() == CloseNode {
			c.setParent(-len(events))
		}
		events = append(events, c)
	}
	events[0].setSkip(len(events))
	events = append(events, newCloseNode(kind, 0, -len(events)))
	return events

}

func newAddToken(kind Kind, begin int, end int) treeEvent {
	return treeEvent{
		kind: uint16(kind),
		data: [2]int32{
			int32(begin),
			int32(end - begin),
		},
	}
}

func newOpenNode(kind Kind, parent int, skip int) treeEvent {
	return treeEvent{
		kind: uint16(kind),
		data: [2]int32{
			int32(parent),
			int32(skip),
		},
	}
}

func newCloseNode(kind Kind, parent int, skip int) treeEvent {
	return treeEvent{
		kind:  uint16(kind),
		flags: uint16(CloseNode),
		data: [2]int32{
			int32(parent),
			int32(skip),
		},
	}
}

// Type returns the type of the event, such as AddToken, OpenNode or CloseNode.
func (e treeEvent) Type() treeEventType {
	switch {
	case e.Kind().IsToken():
		return AddToken
	case e.flags&uint16(CloseNode) != 0:
		return CloseNode
	default:
		return OpenNode
	}
}

// Kind returns the syntax node kind of the event, such as Comment, Identifier, VarDecl, etc.
func (te treeEvent) Kind() Kind {
	return Kind(te.kind)
}

// MarshalJSON implements the json.Marshaler interface.
func (te treeEvent) MarshalJSON() ([]byte, error) {
	var s string
	switch te.Type() {
	case AddToken:
		s = fmt.Sprintf(`{"event":"add","kind":"%s","pos":%d,"len":%d}`, te.Kind().String(), te.offset(), te.length())
	case OpenNode:
		s = fmt.Sprintf(`{"event":"open","kind":"%s"}`, te.Kind().String())
	case CloseNode:
		s = fmt.Sprintf(`{"event":"close"}`)
	default:
		s = fmt.Sprintf(`{"event":"unknown", "d1":%d, "d2":%d, "d3":%d, "d4":%d}`,
			te.kind, te.flags, te.data[0], te.data[1])
	}
	return []byte(s), nil
}

// A treeEvent is essantially a union of a token and a node. Unions are usually
// implemented as interfaces.
//
// However the overhead of interfaces is subsstantial, because we have to deal
// with millions of events
//
// By using an opaque struct we can save a lot of memory and time (three times
// faster). The cost of this approach is that the code becomes more
// complicated.
//
// Functions such as treeEvent.offset or treeEvent.length must be used to
// access the union fields to avoid chaos.

func (te treeEvent) offset() int { assertToken(te); return int(te.data[0]) }
func (te treeEvent) length() int { assertToken(te); return int(te.data[1]) }
func (te treeEvent) parent() int { assertNode(&te); return int(te.data[0]) }
func (te treeEvent) skip() int   { assertNode(&te); return int(te.data[1]) }

func (te *treeEvent) setParent(idx int) { assertNode(te); te.data[0] = int32(idx) }
func (te *treeEvent) setSkip(idx int)   { assertNode(te); te.data[1] = int32(idx) }

func assertToken(te treeEvent) {
	if te.Type() != AddToken {
		panic("not a token")
	}
}
func assertNode(te *treeEvent) {
	if te.Type() == AddToken {
		panic("not a node")
	}
}

type Error struct {
	Errors []error
}

func (e *Error) Error() string {
	switch len(e.Errors) {
	case 0:
		return ""
	case 1:
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", e.Errors[0], len(e.Errors)-1)
}

// panicEvent panics.
//
// A Node is either a terminal or a non-terminal. We use the EventType
// `AddToken` and `OpenNode` to distinguish between the two.
//
// Should we encounter a Node with a different event type, we panic.
// Because any other event means that we either have a corrumpt Node or the
// semantics of the Node type has changed and we forgot to update our
// functions.
func panicEvent(te treeEvent) {
	panic(fmt.Sprintf("logic error: unexpected event type %d", te.Type()))
}
