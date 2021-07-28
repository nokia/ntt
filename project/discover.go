package project

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Discover walks towards the file system root and collects
// known test suite layouts.
//
// This initial version of Discover returns a string slice, but this may change
// in future releases.
func Discover(path string) []string {

	var list []string

	walkUp(path, func(path string) bool {
		// Check source directories
		if file := filepath.Join(path, "package.yml"); isRegular(file) {
			list = append(list, path)
		}
		for _, dir := range readSuites(filepath.Join(path, "ttcn3_suites.json")) {
			list = append(list, dir)
		}

		// Check build directories
		for _, file := range glob(path + "/*build*/ttcn3_suites.json") {
			list = append(list, readSuites(file)...)
		}
		for _, file := range glob(path + "/build/native/*/sct/ttcn3_suites.json") {
			list = append(list, readSuites(file)...)
		}
		return true
	})

	// Remove duplicate entries
	result := make([]string, 0, len(list))
	visited := make(map[string]bool)
	for _, v := range list {
		if !visited[v] {
			visited[v] = true
			result = append(result, v)
		}
	}
	return result
}

func glob(s string) []string {
	found, _ := filepath.Glob(s)
	return found
}

func readSuites(file string) []string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	var (
		data Build
		list []string
	)

	if err := json.Unmarshal(b, &data); err != nil {
		return nil
	}

	for _, suite := range data.Suites {
		if suite.RootDir == "" {
			continue
		}
		if !filepath.IsAbs(suite.RootDir) {
			suite.RootDir = filepath.Join(filepath.Dir(file), suite.RootDir)
		}

		if file := filepath.Join(suite.RootDir, "package.yml"); isRegular(file) {
			list = append(list, suite.RootDir)
		}
	}

	return list
}

func walkUp(path string, f func(path string) bool) {
	for {
		if !f(path) {
			break
		}

		if abs, _ := filepath.Abs(path); abs == "/" {
			break
		}

		path = filepath.Clean(filepath.Join(path, ".."))
	}
}

func isRegular(path string) bool {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		path = p
	}

	if info, err := os.Stat(path); err == nil {
		return !info.IsDir()
	}

	return false
}
