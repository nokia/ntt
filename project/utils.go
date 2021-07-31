package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
)

func getenv(key string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	if strings.HasPrefix(key, "NTT") {
		return os.Getenv(strings.Replace(key, "NTT", "K3", 1))
	}
	return ""
}

func fix(base, path string) string {
	if filepath.IsAbs(path) || path[0] == '$' {
		return path
	}
	return filepath.Join(base, path)
}

func glob(s string) []string {
	found, _ := filepath.Glob(s)
	return found
}

func slugify(s string) string {
	return strings.ReplaceAll(slug.Make(s), "-", "_")
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
		return info.Mode().IsRegular()
	}
	return false
}

func isDir(path string) bool {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		path = p
	}
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}
	return false
}
