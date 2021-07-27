package lsp

import (
	"sync"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
)

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
	roots map[string]*ntt.Suite
}

// Owners returns all Suites that require the opened file.
func (s *Suites) Owners(uri protocol.DocumentURI) []*ntt.Suite {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ret []*ntt.Suite
	for _, suite := range s.roots {
		if project.ContainsFile(suite, string(uri.SpanURI())) {
			ret = append(ret, suite)
		}
	}
	return ret
}

// AddFolder tries to find determine a got root folder for folder and add it to
// the list of know suites.
func (s *Suites) AddFolder(folder string) {
	root := s.FindRoot(folder)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Folder is already known.
	if _, found := s.roots[root]; found {
		return
	}

	if s.roots == nil {
		s.roots = make(map[string]*ntt.Suite)
	}

	suite := &ntt.Suite{}
	suite.SetRoot(root)
	s.roots[root] = suite
}

// FindRoot takes a folder and uses various heuristics to determine its root folder.
func (s *Suites) FindRoot(folder string) string {
	// TODO(5nord) Implement various heuristics
	return folder
}
