// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

// Project describes a TTCN-3 project.
type Project interface {
	// Root is the test suite root folder. It is usually the folder where the manifest is.
	Root() string

	// Sources returns a slice of files and directories containing TTCN-3 source files.
	Sources() []string

	// Imports returns a slice of additional directories required to build a test executable.
	// Codecs, adapters and libraries are specified by Imports, typically.
	Imports() []string

	// ContainsFile returns true, when path is managed by Project.
	ContainsFile(path string) bool
}
