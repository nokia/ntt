/*
Package loader loads a TTCN-3 test suite for inspection and analysis.
*/
package loader

import (
	"errors"
	"strings"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/runtime"
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

	// Logf is the logger for the config.
	// If the user provides a logger, debug logging is enabled.
	// If if the NTT_DEBUG environment variable is set, but the logger is
	// nil, default to log.Printf.
	Logf func(format string, arg ...interface{})

	// Sources is the initial list of source files to load a test suite from.
	Sources []string

	// ImportPackages is the list of ttcn3 packages (codecs, adapters,
	// libraries), the suite if dependening from.
	ImportPackages []string

	// Dir is the source directory of the test suite.
	Dir string

	// Name of the test suite.
	Name string
}

// FromArgs interprets args an initial suite directory or list of ttcn3 source
// files to load from and updates the configuration with Source, Imports, ...
//
// It returns the list of unconsumed arguments. An error is returned if files
// could not be read or if config files had syntax errors.
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

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return nil, err
	}

	name, err := suite.Name()
	if err != nil {
		return nil, err
	}
	conf.Name = name

	if dir := suite.Root(); dir != nil {
		conf.Dir = dir.Path()
	}

	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	conf.Sources = ntt.PathSlice(srcs...)

	imps, err := suite.Imports()
	if err != nil {
		return nil, err
	}
	conf.ImportPackages = ntt.PathSlice(imps...)
	return rest, nil
}

// Load loads a suite.
func (conf Config) Load() (runtime.Suite, error) {
	return NewSuite(conf).load()
}
