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
		if file := filepath.Join(path, "package.yml"); isRegular(file) {
			list = append(list, path)
		}

		if file := filepath.Join(path, "ttcn3_suites.json"); isRegular(file) {
			if b, err := ioutil.ReadFile(file); err == nil {
				var data Build
				if err := json.Unmarshal(b, &data); err == nil {
					for _, suite := range data.Suites {
						if suite.RootDir == "" {
							continue
						}
						if !filepath.IsAbs(suite.RootDir) {
							suite.RootDir = filepath.Join(path, suite.RootDir)
						}

						if file := filepath.Join(suite.RootDir, "package.yml"); isRegular(file) {
							list = append(list, suite.RootDir)
						}
					}
				}
			}
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
