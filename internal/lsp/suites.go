package lsp

import (
	"fmt"
	"sync"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
)

type Suite struct {
	*ntt.Suite
	DB        *ttcn3.DB
	modulesMu sync.Mutex
	modules   map[string]string
}

func (s *Suite) AddSources(srcs ...string) {
	s.Suite.AddSources(srcs...)
	s.DB.Index(srcs...)
}

func (s *Suite) FindModule(name string) (string, error) {
	var files []string
	for file := range s.DB.Modules[name] {
		if project.ContainsFile(s, file) {
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
		for _, file := range project.FindAllFiles(s) {
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

func NewSuite() *Suite {
	return &Suite{
		Suite: &ntt.Suite{},
		DB:    &ttcn3.DB{},
	}
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
		if project.ContainsFile(suite, string(uri.SpanURI())) {
			ret = append(ret, suite)
		}
	}
	return ret
}

// AddSuite add a TTCN-3 test suite to the list of known suites.
// the list of know suites.
func (s *Suites) AddSuite(root project.Suite) {
	s.mu.Lock()

	// Although the folder is known, it might be necessary to re-read it due to
	// a newly saved File

	if s.roots == nil {
		s.roots = make(map[string]*Suite)
	}

	log.Printf("Adding %q to list of known test suites\n", root)
	suite := NewSuite()
	suite.SetRoot(root.RootDir)
	suite.SetSourceDir(root.SourceDir)
	suite.DB = &s.db
	s.roots[root.RootDir] = suite
	files, _ := suite.Files()
	s.mu.Unlock()

	for _, dir := range k3.FindAuxiliaryDirectories() {
		files = append(files, fs.FindTTCN3Files(dir)...)
	}

	s.db.Index(files...)
}
