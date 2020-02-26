/*
Package ntt provides test suite configuration. A test suite configuration is
everything you need for building or running tests: TTCN-3, C/C++ and ASN.1
source files, codecs adapters, hooks, ...

Please note, the TTCN-3 and ASN.1 specs do not have the notion of packages or
even files; only modules in a global scope. This becomes hard to organize and
maintain when the number of files grows. Hence this package introduces the
concept of packages similar to Go:

A package is a directory containing TTCN-3, ASN.1 or C/C++ source files. A test
suite is a special package with an optional manifest (`package.yml`), the main
package so to say.
This manifest file specifies dependencies to other packages, as well as runtime
specific configuration like test_hooks, timeout, module parameters, ... . The
name of a package is the base-name of the directory. So choose the name
carefully. Here are some hints for good naming: Everything which belongs
together should be put together. Package names should describe their purpose,
they should be short and easy to understand. Avoid catchall packages like:
`utils`, `helpers`, `functions`, ...

You may use environment variables and variables from `k3.env` at any place, also
in `package.yml`. They become expanded when the suite is loaded.


Package YAML

The main package (aka suite) may contain a YAML configuration file `package.yml` to specify
dependencies to other package and configure custom runtime configuration. Below
keys are supported:

        name                 Name of the package. Default is the directory name.

        sources              Source files.

        imports              Directories the suite depends on.

        timeout              Default timeout for tests in seconds.

        test_hook            Path to test_hook script. Default is `$K3_NAME.control`.

        parameters_file      Path to module parameters file. Default is `$K3_NAME.parameters`


All keys can be overwritten by environment variables. For instance, to overwrite
`test_hook` just set environment variable `K3_TEST_HOOK`. Is it not
recommended to overwrite list values like `K3_SOURCES` and `K3_IMPORTS`,
because you might run into quoting issues with white-spaces.


Ad-hoc Packages

It is possible to create a suite just from a directory without any
`package.yml`. The Config.Sources will contain the paths of all .ttcn3 sources
files in that directory. Alternatively you may specify a list of .ttcn3 files
directory. The associated will then be the current working directory.

Like with the configuration file, most values can be overwritten by environment
variables.

*/
package ntt

import (
	"fmt"
	"os"
)

// NewFromArgs creates a suite from command line arguments. It expects either
// one directory as argument or a list of regular .ttcn3 files.
//
// Calling NewFromArgs with an empty argument list will create a suite from
// current working directory or, if set, from K3_SOURCE_DIR.
func NewFromArgs(args ...string) (*Suite, error) {
	switch len(args) {
	case 0:
		if source_dir := getenv("source_dir"); source_dir != "" {
			return NewFromDirectory(source_dir)
		}
		return NewFromDirectory(".")
	case 1:
		info, err := os.Stat(args[0])
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return NewFromDirectory(args[0])
		}
	}
	return NewFromFiles(args...)
}

// NewFromFiles creates a suite from a .ttcn3 files list. No root folder will be
// associated with the suite and `package.yml` will be ignored.
//
// Except for `K3_SOURCES` You may use the usual environment variables to
// overwrite default configuration.
//
// Allowing only ttcn3 files is a compromise between keeping k3 tool simple and
// allowing user to create ad-hoc packages for proto-typing. This constraint
// might be lifted in the future.
func NewFromFiles(files ...string) (*Suite, error) {
	for _, f := range files {
		if !hasTTCN3Extension(f) {
			return nil, fmt.Errorf("%q is not a .ttcn3 file.", f)
		}
	}

	suite := &Suite{}
	suite.AddSources(files...)
	return suite, nil
}

// NewFromDirectory creates a suite from directory dir. If the directory
// contains a manifest file (`package.yml`) it will be loaded. You may use the
// usual environment variables to overwrite configuration. If no source files
// were specified, NewFromDirectory will assign all .ttcn3 it finds.
//
// NewFromDirectory returns an error if package.yml was not readable or hab
// syntax errors or if environment variables have incompatible type (e.g.
// float64 for K3_TIMEOUT)
func NewFromDirectory(dir string) (*Suite, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a direcoty.", dir)
	}

	suite := &Suite{}
	suite.SetRoot(dir)
	return suite, nil
}

// SplitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func SplitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}
