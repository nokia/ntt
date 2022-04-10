package ntt

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/project"
)

// Suite represents a TTCN-3 test suite.
type Suite struct {
	id int // A unique session id

	p *project.Project

	// Environent handling
	envFiles []*fs.File

	// Manifest stuff
	name     string
	testHook *fs.File

	// Memoization
	store memoize.Store
}

func (suite *Suite) lazyInit() {
	if suite.p == nil {
		suite.p = &project.Project{}
	}
}

// Id returns the unique session id (aka NTT_SESSION_ID). This ID is the smallest
// integer available on this machine.
func (suite *Suite) Id() (int, error) {
	if suite.id == 0 {
		if s, ok := env.LookupEnv("NTT_SESSION_ID)"); ok {
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

func (suite *Suite) Root() string {
	if suite.p != nil {
		return suite.p.Root()
	}
	return ""
}

// SetRoot set the root folder for Suite.
//
// The root folder is the main-package, which may contain a manifest file
// (`package.yml`)
func (suite *Suite) SetRoot(folder string) error {
	p, err := project.Open(folder)
	suite.p = p
	return err
}

func (suite *Suite) SetSourceDir(folder string) {
	suite.lazyInit()
	suite.p.Config.SourceDir = folder
}

func (suite *Suite) LatestResults() (*results.DB, error) {
	b, err := fs.Open("test_results.json").Bytes()
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
	env.LoadFiles()

	// TODO(5nord) We still have to figure how this sharedDir could be handled
	// more elegantly, maybe even with support for Windows.
	//
	// Change SharedDir to /tmp/k3 to be compatible with legacy k3 scripts.
	session.SharedDir = "/tmp/k3"
}
