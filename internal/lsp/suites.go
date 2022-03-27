package lsp

import (
	"fmt"
	"sort"
	"sync"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
)

type Suite struct {
	*project.Config
	DB        *ttcn3.DB
	modulesMu sync.Mutex
	modules   map[string]string

	filesMu sync.Mutex
	files   map[string]bool
}

func (s *Suite) AddSources(srcs ...string) {
	s.Config.Sources = append(s.Config.Sources, srcs...)
	s.DB.Index(srcs...)
}

func (s *Suite) Files() []string {
	s.filesMu.Lock()
	defer s.filesMu.Unlock()
	if s.files == nil {
		s.files = make(map[string]bool)
		files, _ := project.Files(s.Config)
		for _, dir := range k3.Includes() {
			for _, file := range fs.FindTTCN3Files(dir) {
				files = append(files, file)
			}
		}
		for _, dir := range files {
			s.files[dir] = true
		}
	}

	files := make([]string, 0, len(s.files))
	for file := range s.files {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

func (s *Suite) ContainsFile(file string) bool {
	s.filesMu.Lock()
	defer s.filesMu.Unlock()
	return s.files[file]
}

func (s *Suite) FindModule(name string) (string, error) {

	var files []string
	for file := range s.DB.Modules[name] {
		if s.files[file] {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		log.Debugf("no such module %q. Falling back to legacy implementation\n", name)
		return s.guessModuleByFileName(name)
	}

	if len(files) > 1 {
		log.Debugf("warning: multiple modules with name %q: %v\n", name, files)
	}

	return files[0], nil
}

func (s *Suite) guessModuleByFileName(name string) (string, error) {
	s.modulesMu.Lock()
	defer s.modulesMu.Unlock()

	if s.modules == nil {
		s.modules = make(map[string]string)
		for _, file := range s.Files() {
			name := fs.Stem(file)
			s.modules[name] = file
		}
	}
	if file, ok := s.modules[name]; ok {
		return file, nil
	}

	// Use NTT_CACHE to locate file
	if f := fs.Open(name + ".ttcn3").Path(); fs.IsRegular(f) {
		s.modules[name] = f
		return f, nil
	}
	return "", fmt.Errorf("no such module %q", name)
}

// Suites implements the mapping between ttcn3 source files and multiple test
// suites.
//
// The language server must be able to correctly handle multiple test suites in
// parallel. For example, if a user edits a file, which is imports from two
// independent test suites the language server must not throw a
// redeclaration-error.
//
//
// How do we solve this problem?
// Virtually all lsp features initiated from a open file. For every
// didOpen-event we check if the opened file is owned by at least one test
// suite.
// If there's not, we load a new suite by using the file's directory and
// various other inputs as heuristic.
type Suites struct {
	mu    sync.Mutex
	roots map[string]*Suite

	db ttcn3.DB
}

// FirstSuite returns the first test suite owning the file or an error if not
// owning suite was found.
func (s *Suites) FirstSuite(uri string) (*Suite, error) {
	if suites := s.Owners(protocol.DocumentURI(uri)); len(suites) > 0 {
		return suites[0], nil
	}
	return nil, fmt.Errorf("File %q seem not to belong to any test suite. Skipping execution.", uri)
}

// Owners returns all Suites that require the opened file.
func (s *Suites) Owners(uri protocol.DocumentURI) []*Suite {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ret []*Suite
	for _, suite := range s.roots {
		if suite.ContainsFile(string(uri.SpanURI())) {
			ret = append(ret, suite)
		}
	}
	return ret
}

// AddSuite add a TTCN-3 test suite to the list of known suites.
// the list of know suites.
func (s *Suites) AddSuite(root project.Suite) {
	log.Printf("Adding %q to list of known test suites\n", root)
	conf, err := project.NewConfig(
		project.AutomaticRoot(root.RootDir),
		project.AutomaticEnv(),
		project.WithSourceDir(root.SourceDir),
		project.WithK3(),
		project.WithDefaults(),
	)

	// If opening a project fails, we want to continue with what we have.
	// Therefore we log the error and continue with the configuration we got.
	if err != nil {
		log.Println(err.Error())
	}
	suite := &Suite{
		Config: conf,
		DB:     &s.db,
	}
	s.mu.Lock()

	if s.roots == nil {
		s.roots = make(map[string]*Suite)
	}

	s.roots[root.RootDir] = suite
	s.mu.Unlock()

	// Update index
	files, _ := project.Files(conf)
	s.db.Index(files...)
}
