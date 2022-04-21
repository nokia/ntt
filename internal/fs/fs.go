// Package fs provides a primitive virtual file system.
package fs

import (
	"net/url"
	"path"
	"path/filepath"

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

// JoinPath joins any number of path elements into a single path, separating
// them with an OS specific Separator. Empty elements are ignored. The result
// is Cleaned. If the argument list is empty or all its elements are empty,
// JoinPath returns an empty string.
//
// If baseUrl is an absolute URL, JoinPath will will keep the scheme of the
// first argument and separats the rest with a slash.
func JoinPath(baseUrl string, elem ...string) string {
	// We know first argument is a path, when there's a volume name.
	if filepath.VolumeName(baseUrl) != "" {
		elem = append([]string{baseUrl}, elem...)
		return filepath.Join(elem...)
	}

	// We know first argument is an URL, when there's a scheme.
	if url, err := url.Parse(baseUrl); err == nil && url.Scheme != "" {
		elem = append([]string{url.Path}, elem...)
		url.Path = path.Join(elem...)
		return url.String()
	}

	// Everything else is treated as a path.
	elem = append([]string{baseUrl}, elem...)
	return filepath.Join(elem...)
}

// Path returns a decoded file path when you pass a URI with file:// scheme.
func Path(s string) string {
	if IsURI(s) {
		return span.URIFromURI(s).Filename()
	}
	return s
}

func IsURI(s string) bool {
	url, err := url.Parse(s)
	return err == nil && url.Scheme != ""
}

// URI turns paths into URIs
func URI(path string) span.URI {
	path = string(span.URINormalizeAuthority(path))
	if u, _ := url.Parse(path); u != nil && u.Scheme != "" {
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
