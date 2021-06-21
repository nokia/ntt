package ntt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/internal/session"
)

// Suite represents a TTCN-3 test suite.
type Suite struct {
	id int // A unique session id

	// Module handling (maps module names to paths)
	modulesMu sync.Mutex
	modules   map[string]string

	// Environent handling
	envFiles []*fs.File

	// Manifest stuff
	name     string
	root     *fs.File
	sources  []*fs.File
	imports  []*fs.File
	testHook *fs.File

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
func (suite *Suite) File(path string) *fs.File {
	if !strings.HasPrefix(path, "file://") {
		if ok, _ := fileExists(path); !ok {
			if s := suite.searchCacheForFile(path); s != "" {
				path = s
			}
		}
	}

	return fs.Open(path)
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

	if file == "." || file == ".." {
		return ""
	}

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

func (suite *Suite) Root() *fs.File {
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

func (suite *Suite) LatestResults() (*results.DB, error) {
	b, err := suite.File("test_results.json").Bytes()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var db results.DB
	return &db, json.Unmarshal(b, &db)
}

func init() {
	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
