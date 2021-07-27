// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

// Project describes a TTCN-3 project.
type Project interface {
	Root() string
	Sources() []string
	Imports() []string
	ContainsFile(path string) bool
}
