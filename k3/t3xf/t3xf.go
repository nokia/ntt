package t3xf

// Reference is a T3XF offset.
type Reference int

// String is a T3XF string.
type String struct {
	l int
	b []byte
}

func NewString(l int, b []byte) *String {
	return &String{l, b}
}

// Len returns the number of elements in the string. For a bitstring, this is
// the number of bits, for a charstring the number of characters.
func (s *String) Len() int {
	return s.l
}

// Bytes returns the byte slice for the string.
func (s *String) Bytes() []byte {
	return s.b
}
