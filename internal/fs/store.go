package fs

import (
	"sync"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/lsp/span"
)

// A Store holds all open files.
type Store struct {
	files   map[span.URI]*File
	filesMu sync.Mutex
}

// Open a file and add it to the store.
func (s *Store) Open(path string) *File {
	return s.open(cache.Lookup(path))
}

func (s *Store) open(path string) *File {
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
