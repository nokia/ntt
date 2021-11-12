// Package fs provides a primitive virtual file system.
package fs

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/span"
)

var store = Store{}

// Open a file.
//
// path can be any identifier, URL, ...
func Open(path string) *File {
	return store.Open(path)
}

// PathSlice returns a string slice of the File objects passed as argument.
func PathSlice(files ...*File) []string {
	ret := make([]string, 0, len(files))
	for i := range files {
		if files[i] != nil {
			ret = append(ret, files[i].String())
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

// Path returns a decoded file path when you pass a URI with file:// scheme.
func Path(s string) string {
	if !strings.HasPrefix(s, "file://") {
		return s
	}
	return span.URIFromURI(s).Filename()
}

// URI turns paths into URIs
func URI(path string) span.URI {
	path = string(span.URINormalizeAuthority(path))
	if u, _ := url.Parse(path); u.Scheme != "" {
		if vol := filepath.VolumeName(path); vol != "" {
			// this is a windows path
			return span.URIFromPath(path)
		}
		// VSCode tends to overquote URIs. URIFromURI normalizes them a little.
		return span.URIFromURI(path)
	}
	return span.URIFromPath(path)
}

// Content returns the content of (virtual) file specified by path.
func Content(path string) ([]byte, error) {
	return Open(path).Bytes()
}

// SetContent of (virtual) file specified by path.
func SetContent(path string, b []byte) {
	Open(path).SetBytes(b)
}
