package opcode

// Descriptor describes a t3xf opcode.
type Descriptor struct {
	Opcode int // The opcode number

	// Context describe the contexts in which the instruction can be used.
	Context []string `json:",omitempty"`

	// Operations describe the stack operations that the instruction performs.
	Operations  Operation
	Description string // A description of the opcode
}

// Operation describes a stack operation. Pre is the stack before the
// operation, Post is the stack after the operation.
type Operation struct {
	Pre  []string `yaml:",flow"`
	Post []string `yaml:",flow"`
}
