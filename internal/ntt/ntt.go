package ntt

import (
	"io/ioutil"
	"sync"

	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/internal/span"
)

type Suite struct {
	id int // A unique session id

	// File handling
	filesMu sync.Mutex
	files   map[span.URI]*File
}

// Id returns the unique session id (aka K3_SESSION_ID). This ID is the smallest
// integer available on this machine.
func (s *Suite) Id() (int, error) {
	if s.id == 0 {
		id, err := session.Get()
		if err != nil {
			return 0, err
		}
		s.id = id
	}
	return s.id, nil
}

// File returns a new file struct for reading.
func (s *Suite) File(path string) *File {
	uri := span.NewURI(path)

	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	if s.files == nil {
		s.files = make(map[span.URI]*File)
	}

	if f, found := s.files[uri]; found {
		return f
	}

	f := &File{
		uri:  uri,
		path: path,
	}
	s.files[uri] = f
	return f
}

type File struct {
	uri     span.URI // URI
	path    string   // Original path used on construction
	bytes   []byte   // nil is file hasn't been read yet
	err     error    // error of previous read
	version int
}

// Bytes returns the contents of the file
func (f *File) Bytes() ([]byte, error) {
	if f.bytes == nil && f.err == nil {
		f.bytes, f.err = ioutil.ReadFile(f.path)
		f.version = 0
	}

	return f.bytes, f.err
}

// SetBytes set the contents of the file.
func (f *File) SetBytes(b []byte) {
	f.bytes = b
	f.err = nil
	f.version++
}

// Reset sets the content to zero. This is identcal to SetBytes(nil)
func (f *File) Reset() {
	f.SetBytes(nil)
}

func init() {
	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
