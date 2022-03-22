package project

import (
	"path/filepath"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

// Discover walks towards the file system root and collects
// known test suite layouts.
//
// Discover returns a list of potential test suite root directories.
func Discover(path string) []Suite {

	// Convert possible URIs to proper file system paths.
	path = fs.Path(path)

	var list []Suite

	fs.WalkUp(path, func(path string) bool {
		// Check source directories
		if file := filepath.Join(path, ManifestFile); fs.IsRegular(file) {
			log.Debugf("discovered manifest: %q\n", file)
			list = append(list, Suite{RootDir: path, SourceDir: path})
		}
		list = append(list, readIndices(filepath.Join(path, IndexFile))...)

		// Check build directories
		for _, file := range fs.Glob(path + "/*build*/" + IndexFile) {
			list = append(list, readIndices(file)...)
		}
		for _, file := range fs.Glob(path + "/build/native/*/sct/" + IndexFile) {
			list = append(list, readIndices(file)...)
		}
		return true
	})

	// If we could not find any manifest, try guess a root directory based on known naming schemes.
	if len(list) == 0 {
		fs.WalkUp(path, func(path string) bool {
			if tests := fs.Glob(path + "/testcases/*"); len(tests) > 0 {
				log.Debugf("discovered testcases folder in %q\n", path)
				list = append(list, Suite{RootDir: path, SourceDir: path})
				return false
			}
			return true
		})
	}

	// Remove duplicate entries
	result := make([]Suite, 0, len(list))
	visited := make(map[Suite]bool)
	for _, v := range list {
		if !visited[v] {
			visited[v] = true
			result = append(result, v)
		}
	}
	return result
}

func readIndices(file string) []Suite {
	var list []Suite

	si, err := ReadIndex(file)
	if err != nil {
		return nil
	}

	log.Debugf("reading suites from index file: %q\n", file)
	for _, suite := range si.Suites {
		if suite.RootDir != "" {
			log.Debugf("using root_dir: %q\n", suite.RootDir)
			list = append(list, suite)
		}
	}

	return list
}
