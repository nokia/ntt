// Package printer implements pretty printers for TTCN-3 source code.
package printer

var std = &simpleFormatter{}

// Bytes formats src in canonical TTCN-3 style and returns the result or an
// (I/O or syntax) error. src is expected to syntactically correct TTCN-3
// source text.
func Bytes(src []byte) ([]byte, error) {
	return std.Bytes(src)
}

// simpleFormatter is a simple formatter that only fixes indentation and
// various whitespace issues.
type simpleFormatter struct{}

func (f *simpleFormatter) Bytes(src []byte) ([]byte, error) {
	return src, nil
}
