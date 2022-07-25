package ttcn3

type Source struct {
	Filename string
	Events   []NodeEvent `json:"events"`
}

type NodeEvent struct {
	// Kind is the type of an node event (AddToken, OpenFooBar, CloseFooBar).
	Kind string `json:"kind"`

	// Text is the text of a AddToken event.
	Text string `json:"text,omitempty"`

	// Offset is the position of the first character belonging to the node.
	Offs int `json:"offs"`

	// End is the position of the first character immediately after the node.
	Len int `json:"len"`

	// Other is the index of the matching node event.
	Other int `json:"other,omitempty"`
}
