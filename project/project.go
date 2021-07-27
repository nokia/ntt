// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

import "github.com/nokia/ntt/internal/fs"

// Project describes a TTCN-3 project.
type Project interface {
	// Root is the test suite root folder. It is usually the folder where the manifest is.
	Root() string

	// Sources returns a slice of files and directories containing TTCN-3 source files.
	Sources() ([]string, error)

	// Imports returns a slice of additional directories required to build a test executable.
	// Codecs, adapters and libraries are specified by Imports, typically.
	Imports() ([]string, error)
}

// Files returns all .ttcn3 available. It will not return generated .ttcn3 files.
// On error Files will return an error.
func Files(p Project) ([]string, error) {
	files, err := p.Sources()
	if err != nil {
		return nil, err
	}

	dirs, err := p.Imports()
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		f := fs.FindTTCN3Files(dir)
		files = append(files, f...)
	}

	return files, nil
}

// ContainsFile returns true, when path is managed by Project.
func ContainsFile(p Project, path string) bool {
	path = normalize(path)

	files, _ := Files(p)
	for _, file := range files {
		if normalize(file) == path {
			return true
		}
	}
	return false
}

// normalize turns URIs and paths into absolute paths.
func normalize(path string) string {
	return fs.Open(path).URI().Filename()
}
