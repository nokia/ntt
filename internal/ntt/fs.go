package ntt

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// hasASN1Extension returns true if file has suffix .asn or .asn1
func hasASN1Extension(file string) bool {
	return strings.HasSuffix(file, ".asn") || strings.HasSuffix(file, ".asn1")
}

// hasTTCN3Extension returns true if file has suffix .ttcn3 or .ttcn
func hasTTCN3Extension(file string) bool {
	return strings.HasSuffix(file, ".ttcn3") || strings.HasSuffix(file, ".ttcn")
}

// hasCExtension returns true if file has suffix .c, .cc, .cxx or .cpp
func hasCExtension(file string) bool {
	return strings.HasSuffix(file, ".c") ||
		strings.HasSuffix(file, ".cc") || strings.HasSuffix(file, ".cxx") || strings.HasSuffix(file, ".cpp")
}

// findTTCN3Files returns a list of TTCN-3 source files (.ttcn3, .ttcn) or error if
// directory could not be read.
func findTTCN3Files(dir string) ([]string, error) {
	return findFiles(dir, hasTTCN3Extension)
}

// findASN1Files returns a list of ASN.1 files (.asn, asn1) or error if directory could
// not be read.
func findASN1Files(dir string) ([]string, error) {
	return findFiles(dir, hasASN1Extension)
}

// findCFiles returns a list of C/C++ files (.c, .cc, .cxx) or error if directory
// could not be read.
func findCFiles(dir string) ([]string, error) {
	return findFiles(dir, hasCExtension)
}

// findFiles returns regular files from directory dir, for which matcher function
// returned true.
// findFiles returns an empty slice if no file was found. On error it returns
// nil and error.
func findFiles(dir string, matcher func(name string) bool) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sources := make([]string, 0, len(files))
	for _, file := range files {
		if file.Mode().IsRegular() && matcher(file.Name()) {
			sources = append(sources, filepath.Join(dir, file.Name()))
		}
	}

	return sources, nil
}
