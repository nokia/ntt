package lsp

import (
	"fmt"
	"sync"

	"github.com/nokia/ntt/internal/log"
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

// FirstSuite returns the first test suite owning the file or an error if not
// owning suite was found.
func (s *Suites) FirstSuite(uri string) (*ntt.Suite, error) {
	if suites := s.Owners(protocol.DocumentURI(uri)); len(suites) > 0 {
		return suites[0], nil
	}
	return nil, fmt.Errorf("File %q seem not to belong to any test suite. Skipping execution.", uri)
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

// AddSuite add a TTCN-3 test suite to the list of known suites.
// the list of know suites.
func (s *Suites) AddSuite(root string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Folder is already known.
	if _, found := s.roots[root]; found {
		log.Printf("%q is already known\n", root)
		return
	}

	if s.roots == nil {
		s.roots = make(map[string]*ntt.Suite)
	}

	log.Printf("Adding %q to list of known test suites\n", root)
	suite := &ntt.Suite{}
	suite.SetRoot(root)
	s.roots[root] = suite
}
