package fs

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/span"
)

// A Store holds all open files.
type Store struct {
	files   map[span.URI]*File
	filesMu sync.Mutex
}

// Open a file and add it to the store.
func (s *Store) Open(path string) *File {
	if !strings.HasPrefix(path, "file://") {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if alt := searchCacheForFile(path); alt != "" {
				path = alt
			}
		}
	}
	uri := URI(path)

	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	if s.files == nil {
		s.files = make(map[span.URI]*File)
	}

	if f, found := s.files[uri]; found {
		return f
	}

	f := &File{
		store: s,
		uri:   uri,
		path:  path,
	}
	s.files[uri] = f

	return f
}

// searchCacheForFile searches for a file in every directory specified by
// environment variable NTT_CACHE.
//
// If found, the directory with joined file name will be returned.
//
// searchCacheForFile will return an empty string:

//  * if the file could not be found
//  * if NTT_CACHE is empty
//  * if file-string has a directory prefix.
//
// Purpose and behaviour of this function are similar to GNU Make's VPATH. It
// is used to prevent re-built of generated files.
func searchCacheForFile(file string) string {

	if file == "." || file == ".." {
		return ""
	}

	if dirPrefix, _ := filepath.Split(file); dirPrefix != "" {
		return ""
	}

	if cache := getenv("NTT_CACHE"); cache != "" {
		for _, dir := range strings.Split(cache, ":") {
			file := filepath.Join(dir, file)
			if _, err := os.Stat(file); err == nil {
				return file
			}
		}
	}

	return ""
}

func getenv(key string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}
	if strings.HasPrefix(key, "NTT") {
		return os.Getenv(strings.Replace(key, "NTT", "K3", 1))
	}
	return ""
}
