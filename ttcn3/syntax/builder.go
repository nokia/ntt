package syntax

// A Builder is used to build a syntax tree. The zero value is ready to use.
type Builder struct {
	tree
	stack []int
}

// SetName sets a name for the syntax tree.
func (b *Builder) SetName(name string) {
	b.name = name
}

// Push pushes appends a new non-terminal of given kind to the syntax tree.
func (b *Builder) Push(k Kind) Node {
	parent := 0
	if len(b.stack) > 0 {
		parent = b.stack[len(b.stack)-1]
	}
	idx := len(b.events)

	b.stack = append(b.stack, idx)
	b.events = append(b.events, newOpenNode(k, parent-idx, 0))
	return Node{tree: &b.tree, idx: idx}
}

// PushToken appends a new token to the syntax tree.
func (b *Builder) PushToken(k Kind, begin int, end int) Node {
	b.events = append(b.events, newAddToken(k, begin, end))
	return Node{tree: &b.tree, idx: len(b.events) - 1}
}

// Pop finishes a non-terminal. Invalid pop calls will panic.
func (b *Builder) Pop() {
	if len(b.stack) == 0 {
		panic("syntax: builder: pop on empty stack")
	}

	// Pop the current non-terminal index from the stack.
	idx := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	b.events[idx].setSkip(len(b.events) - idx)
	b.events = append(b.events, newCloseNode(
		b.events[idx].Kind(),
		idx+b.events[idx].parent()-len(b.events),
		idx-len(b.events)),
	)
}
