package types

// Type represents a type in TTCN-3
type Type interface {
	// Underlying  returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

func Compatible(a, b Type) bool {
	switch {
	case isAnyOf(Typ[Invalid], a, b):
		// Invalid types are compatible with all types. This is necessary to
		// propagate errors and prevent spurious type mismatches.
		return true

	case isAnyOf(Typ[Numerical], a, b):
		return isAnyOf(a, Typ[Integer], Typ[Float]) || isAnyOf(b, Typ[Integer], Typ[Float])

	case isAnyOf(Typ[String], a, b):
		return isAnyOf(a, Typ[Charstring], Typ[UniversalCharstring], Typ[Bitstring], Typ[Hexstring], Typ[Octetstring]) ||
			isAnyOf(b, Typ[Charstring], Typ[UniversalCharstring], Typ[Bitstring], Typ[Hexstring], Typ[Octetstring])
	}

	//
	// If the root types are not the same, they are not compatible.
	if a.Underlying() != b.Underlying() {
		return false
	}

	return true
}

func isAnyOf(t Type, ts ...Type) bool {
	for _, t2 := range ts {
		if t == t2 {
			return true
		}
	}
	return false
}

type Kind int

const (
	Invalid Kind = iota
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

	// TODO(5nord) Merge strings types into String and merge integer, float into
	// Numerical. Or make them sort of untyped types like unused, omit and
	// template?
	String
	Numerical
)

type BasicType struct {
	kind Kind
	name string
}

func (b *BasicType) Kind() Kind       { return b.kind }
func (b *BasicType) Underlying() Type { return b }
func (b *BasicType) String() string   { return b.name }
