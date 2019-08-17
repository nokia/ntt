// Package fs provides simplified filesystem routines and helper for common
// TTCN-3 build tasks.
package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// IsExist returns true if path exists.
func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// IsDirectory returns true is path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsRegular returns true is path is a regular file.
func IsRegular(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// ContainsVar returns true is path contains variables
func IsVar(path string) bool {
	hasVars := false
	mapper := func(name string) string { hasVars = true; return "" }
	os.Expand(path, mapper)
	return hasVars
}

// FindFile joins all elements into a path. If the path exists it will be
// returned by FindFile. It the does not exist FindFile will return an empty
// string.
func FindFile(elements ...string) string {
	if file := filepath.Join(elements...); IsExist(file) {
		return file
	}
	return ""
}

// FindFiles returns regular files from directory dir, for which matcher function
// returned true.
// FindFiles returns an empty slice if no file was found. On error it returns
// nil and error.
func FindFiles(dir string, matcher func(name string) bool) ([]string, error) {
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

// Basename returns the real directory base name, by making path absolute and
// then returning the last element.
func Basename(path string) string {
	abs, _ := filepath.Abs(path)
	return filepath.Base(abs)
}
