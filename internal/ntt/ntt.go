package ntt

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/internal/span"
)

// Suite represents a TTCN-3 test suite.
type Suite struct {
	id int // A unique session id

	// File handling
	filesMu sync.Mutex
	files   map[span.URI]*File

	// Module handling (maps module names to paths)
	modulesMu sync.Mutex
	modules   map[string]string

	// Environent handling
	envFiles []*File

	// Manifest stuff
	name     string
	root     *File
	sources  []*File
	imports  []*File
	testHook *File

	// Memoization
	store memoize.Store
}

// Id returns the unique session id (aka NTT_SESSION_ID). This ID is the smallest
// integer available on this machine.
func (suite *Suite) Id() (int, error) {
	if suite.id == 0 {
		if s, _ := suite.lookupProcessEnv("NTT_SESSION_ID)"); s != "" {
			id, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return 0, err
			}
			suite.id = int(id)
			return suite.id, nil
		}
		id, err := session.Get()
		if err != nil {
			return 0, err
		}
		suite.id = id
	}
	return suite.id, nil
}

// File returns a new file struct for reading.
//
// Environment variable NTT_CACHE will be used to find path, if path is a single
// file-name without leading directory.
func (suite *Suite) File(path string) *File {

	if s := suite.searchCacheForFile(path); s != "" {
		path = s
	}

	uri := span.NewURI(path)

	suite.filesMu.Lock()
	defer suite.filesMu.Unlock()

	if suite.files == nil {
		suite.files = make(map[span.URI]*File)
	}

	if f, found := suite.files[uri]; found {
		return f
	}

	f := &File{
		uri:  uri,
		path: path,
	}
	suite.files[uri] = f
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
func (suite *Suite) searchCacheForFile(file string) string {
	if dirPrefix, _ := filepath.Split(file); dirPrefix != "" {
		return ""
	}

	if cache, _ := suite.Getenv("NTT_CACHE"); cache != "" {
		for _, dir := range strings.Split(cache, ":") {
			file := filepath.Join(dir, file)
			if _, err := os.Stat(file); err == nil {
				return file
			}
		}
	}

	return ""
}

func (suite *Suite) Root() *File {
	return suite.root
}

// SetRoot set the root folder for Suite.
//
// The root folder is the main-package, which may contain a manifest file
// (`package.yml`)
func (suite *Suite) SetRoot(folder string) {
	suite.root = suite.File(folder)
	suite.sources = nil
}

func init() {
	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
