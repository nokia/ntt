package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// IsASN1File returns true if file has suffix .asn or .asn1
func IsASN1File(file string) bool {
	return strings.HasSuffix(file, ".asn") || strings.HasSuffix(file, ".asn1")
}

// IsTTCN3File returns true if file has suffix .ttcn3 or .ttcn
func IsTTCN3File(file string) bool {
	return strings.HasSuffix(file, ".ttcn3") || strings.HasSuffix(file, ".ttcn")
}

// IsCFile returns true if file has suffix .c, .cc, .cxx or .cpp
func IsCFile(file string) bool {
	return strings.HasSuffix(file, ".c") ||
		strings.HasSuffix(file, ".cc") || strings.HasSuffix(file, ".cxx") || strings.HasSuffix(file, ".cpp")
}

// TTCN3Files returns a list of TTCN-3 source files (.ttcn3, .ttcn) or error if
// directory could not be read.
func TTCN3Files(dir string) ([]string, error) {
	return FindFiles(dir, IsTTCN3File)
}

// ASN1Files returns a list of ASN.1 files (.asn, asn1) or error if directory could
// not be read.
func ASN1Files(dir string) ([]string, error) {
	return FindFiles(dir, IsASN1File)
}

// CFiles returns a list of C/C++ files (.c, .cc, .cxx) or error if directory
// could not be read.
func CFiles(dir string) ([]string, error) {
	return FindFiles(dir, IsCFile)
}

// FindTTCN3Files returns a list of .ttcn3 files. If path is a directory .ttcn3
// files inside that directory are returned.
func FindTTCN3Files(path string) ([]string, error) {

	sources := make([]string, 0)

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	switch {
	case info.Mode().IsRegular():
		if !IsTTCN3File(path) {
			return nil, fmt.Errorf("expected a .ttcn3 file: " + path)
		}
		sources = append(sources, path)
	case info.Mode().IsDir():
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if f.Mode().IsRegular() && IsTTCN3File(f.Name()) {
				sources = append(sources, filepath.Join(path, f.Name()))
			}
		}
	default:
		return nil, fmt.Errorf("expected directory or regular file: " + path)
	}

	return sources, nil
}
