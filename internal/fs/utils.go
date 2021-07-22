package fs

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

// FindTTCN3Files returns a list of TTCN-3 source files (.ttcn3, .ttcn).
// FindTTCN3Files will return a nil slice on any error.
func FindTTCN3Files(dir string) []string {
	return findFiles(dir, HasTTCN3Extension)
}

// HasTTCN3Extension returns true if file has suffix .ttcn3 or .ttcn
func HasTTCN3Extension(file string) bool {
	return withExtension(".ttcn3", ".ttcn")(file)
}

// FindASN1Files returns a list of ASN.1 files (.asn, asn1).
func FindASN1Files(dir string) []string {
	return findFiles(dir, HasASN1Extension)
}

// HasASN1Extension returns true if file has suffix .asn or .asn1
func HasASN1Extension(file string) bool {
	return withExtension(".asn", ".asn1")(file)
}

// FindCFiles returns a list of C/C++ files (.c, .cc, .cxx, .cpp).
func FindCFiles(dir string) []string {
	return findFiles(dir, HasCExtension)
}

// FindFilesRecursive returns a list files from the whole directory subtree.
func FindFilesRecursive(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	var sources []string
	for _, file := range files {
		if file.Mode().IsRegular() {
			fname := file.Name()
			fname, _ = filepath.Abs(filepath.Join(dir, fname))
			sources = append(sources, ":"+fname)
		} else if file.Mode().IsDir() {
			sources = append(sources, FindFilesRecursive(filepath.Join(dir, file.Name()))...)
		}
	}
	return sources
}

// HasCExtension returns true if file has suffix .c, .cc, .cxx or .cpp
func HasCExtension(file string) bool {
	return withExtension(".c", ".cc", ".cxx", ".cpp")(file)
}

// FindK3EnvInCurrPath returns the path of the directory containing k3.env
func FindK3EnvInCurrPath(dir string) string {
	path := dir
	for len(path) > 0 {

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return ""
		}
		for _, file := range files {
			if file.Mode().IsRegular() && ((file.Name() == "k3.env") || file.Name() == "ntt.env") {
				return path
			}
		}
		path = filepath.Dir(path)
	}
	return ""
}

func findFiles(dir string, matcher func(name string) bool) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	var sources []string
	for _, file := range files {
		if file.Mode().IsRegular() && matcher(file.Name()) {
			sources = append(sources, filepath.Join(dir, file.Name()))
		}
	}
	return sources
}

func withExtension(exts ...string) func(string) bool {
	return func(file string) bool {
		for _, ext := range exts {
			if strings.HasSuffix(file, ext) {
				return true
			}
		}
		return false
	}
}
