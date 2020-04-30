package ntt

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/internal/span"
)

type Suite struct {
	id int // A unique session id

	// File handling
	filesMu sync.Mutex
	files   map[span.URI]*File

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

// Id returns the unique session id (aka K3_SESSION_ID). This ID is the smallest
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
// Environment variable K3_CACHE will be used to find path, if path is a single
// file-name without leading directory.
func (suite *Suite) File(path string) *File {

	if cache, _ := suite.Getenv("NTT_CACHE"); cache != "" {
		if dir, file := filepath.Split(path); dir == "" {
			for _, dir := range strings.Split(cache, ":") {
				file := filepath.Join(dir, file)
				if _, err := os.Stat(file); err == nil {
					path = file
					goto found
				}
			}
		}
	}

found:
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

// Position returns human readable description of the pos tag.
func (suite *Suite) Position(file string, pos loc.Pos) loc.Position {
	return suite.Parse(file).Position(pos)
}

func (suite *Suite) Pos(file string, line int, column int) loc.Pos {
	return suite.Parse(file).Pos(line, column)
}

func init() {
	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
