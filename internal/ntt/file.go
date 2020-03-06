package ntt

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/span"
)

type File struct {
	uri     span.URI // URI
	path    string   // Original path used on construction
	bytes   []byte   // nil is file hasn't been read yet
	err     error    // error of previous read
	version int

	handle *memoize.Handle
}

func (f *File) URI() span.URI  { return f.uri }
func (f *File) Path() string   { return f.path }
func (f *File) String() string { return f.path }
func (f *File) ID() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(strconv.Itoa(f.version)+f.URI().Filename())))
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

// PathSlice returns a string slice of the File object passed as argument.
func PathSlice(files ...*File) []string {
	ret := make([]string, 0, len(files))
	for i := range files {
		if files[i] != nil {
			ret = append(ret, files[i].Path())
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}
