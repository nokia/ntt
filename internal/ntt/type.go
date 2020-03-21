package ntt

// Type represents a type in TTCN-3
type Type interface {
	// Underlying  returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

type BasicKind int

const (
	Invalid BasicKind = iota
	Bitstring
	Boolean
	Charstring
	Component
	Float
	Hexstring
	Integer
	Octetstring
	Omit
	Template
	Timer
	UniversalCharstring
	Verdict
)

type BasicType struct {
	kind BasicKind
	name string
}

func (b *BasicType) Kind() BasicKind  { return b.kind }
func (b *BasicType) Underlying() Type { return b }
func (b *BasicType) String() string   { return b.name }
