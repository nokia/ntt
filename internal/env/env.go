package env

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/subosito/gotenv"
)

// Load environment files ntt.env and k3.env
func Load() {
	gotenv.Load(fromCache("ntt.env"))
	gotenv.Load(fromCache("k3.env"))
}

// Lookup key in process environment. Is key begins with "NTT_" also lookup key
// with "K3_" prefix.
func Getenv(key string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	if strings.HasPrefix(key, "NTT") {
		return os.Getenv(strings.Replace(key, "NTT", "K3", 1))
	}
	return ""
}

func fromCache(file string) string {
	if cache := Getenv("NTT_CACHE"); cache != "" {
		for _, dir := range strings.Split(cache, ":") {
			file := filepath.Join(dir, file)
			if _, err := os.Stat(file); err == nil {
				return file
			}
		}
	}

	return file
}
