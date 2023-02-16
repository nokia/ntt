// Package cmake provides basic support for reading CMake cache files.
package cmake

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

type Cache struct {
	path string
}

func (c *Cache) Get(name string) string {

	f, err := os.Open(c.path)
	if err != nil {
		log.Verboseln("cmake:", err.Error())
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if v := strings.SplitN(s.Text(), ":", 2); v[0] == name {
			w := strings.SplitN(v[1], "=", 2)
			log.Debugln("cmake: get:", name, "=", w[1])
			return w[1]
		}
	}
	return ""
}

// FindCache returns the path to the CMakeCache.txt file, by walking up
// the current working directory and the runtime directory specified by
// environment variable K3R
func FindCache() *Cache {
	find := func(path string) string {
		var res string
		fs.WalkUp(path, func(path string) bool {
			if file := filepath.Join(path, "CMakeCache.txt"); fs.IsRegular(file) {
				res = file
			}
			return true
		})
		return res
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Verboseln("cmake: cwd:", err.Error())
		return nil
	}
	if f := find(cwd); f != "" {
		return &Cache{path: f}
	}
	if k3r := os.Getenv("K3R"); strings.HasSuffix(k3r, "src/k3r/k3r") {
		return &Cache{path: find(filepath.Dir(k3r))}
	}

	return nil
}
