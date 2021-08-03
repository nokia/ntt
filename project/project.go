// Package project collects information about test suite organisation by
// implementing various heuristics.
package project

import (
	"crypto/sha1"
	"fmt"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/k3"
)

// Interface describes a TTCN-3 project.
type Interface interface {
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
func Files(p Interface) ([]string, error) {
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

// FindAllFiles returns all .ttcn3 files including auxiliary files from
// k3 installation
func FindAllFiles(p Interface) []string {
	files, _ := Files(p)
	// Use auxilliaryFiles from K3 to locate file
	for _, dir := range k3.FindAuxiliaryDirectories() {
		for _, file := range fs.FindTTCN3Files(dir) {
			files = append(files, file)
		}
	}
	return files
}

// ContainsFile returns true, when path is managed by Interface.
func ContainsFile(p Interface, path string) bool {

	// The same file may be referenced by URI or by path. To normalize it
	// we convert everything into URIs.
	uri := fs.URI(path)

	files, _ := Files(p)
	for _, file := range files {
		if fs.URI(file) == uri {
			return true
		}
	}
	return false
}

// Fingerprint calculates a sum to identify a test suite based on its modules.
func Fingerprint(p Interface) string {
	var inputs []string
	files, _ := Files(p)
	for _, file := range files {
		inputs = append(inputs, fs.Stem(file))
	}
	return fmt.Sprintf("project_%x", sha1.Sum([]byte(fmt.Sprint(inputs))))
}
