/*
Package loader loads a TTCN-3 test suite for inspection and analysis.
*/
package loader

import (
	"errors"
	"sort"
	"strings"

	"github.com/nokia/ntt/internal/loc"
	st "github.com/nokia/ntt/internal/suite"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// A Config specifies how a test suite should be loaded.
type Config struct {
	// If IgnoreImports is set, import definitions are ignored. Only source
	// files are processed.
	IgnoreImports bool

	// If IgnoreTags is set, comments won't be searched for tags.
	IgnoreTags bool

	// If IgnoreComments is set, comments will be ignored by the parser.
	IgnoreComments bool

	// Fset is the location mapping to use. If nil, it may be initialized
	// lazily when required.
	Fset *loc.FileSet

	// Logf is the logger for the config.
	// If the user provides a logger, debug logging is enabled.
	// If if the K3_DEBUG environment variable is set, but the logger is
	// nil, default to log.Printf.
	Logf func(format string, arg ...interface{})

	*st.Suite
}

type Suite struct {
	Fset   *loc.FileSet
	Syntax []*ast.Module
}

// FromArgs interprets args an initial suite directory or list of ttcn3 source
// files to load from and updates the configuration. It returns the list of
// unconsumed arguments.
func (conf *Config) FromArgs(args []string) ([]string, error) {
	var rest []string
	for i, arg := range args {
		if arg == "--" {
			rest = args[i+1:]
			args = args[:1]
			break // consume "--" and return the remaining args
		}
	}

	if len(args) > 0 && strings.HasSuffix(args[0], ".t3xf") {
		return nil, errors.New("t3xf format is not supported")
	}

	suite, err := st.NewFromArgs(args)
	if err != nil {
		return nil, err
	}
	suite.SetEnv()
	sort.Strings(suite.Sources)
	conf.Suite = suite
	return rest, nil
}

func (conf *Config) Load() (*Suite, error) {
	suite := &Suite{}
	suite.Fset = loc.NewFileSet()
	if conf.Fset != nil {
		suite.Fset = conf.Fset
	}
	root, err := parseFiles(suite.Fset, conf.Sources)
	if err != nil {
		return nil, err
	}

	suite.Syntax = root
	return suite, nil
}
