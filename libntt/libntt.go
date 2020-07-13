package main

import "C"
import (
	"fmt"
	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loader"
	"github.com/nokia/ntt/internal/runtime"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"encoding/json"
)

func fatal(err error) {
	switch err := err.(type) {
	case *exec.ExitError:
		waitStatus := err.Sys().(syscall.WaitStatus)
		os.Exit(waitStatus.ExitStatus())
	case errors.ErrorList:
		errors.PrintError(os.Stderr, err)
	default:
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(1)
}

func loadSuite(args []string, conf loader.Config) runtime.Suite {
	// Update configuration with TTCN-3 source files from args
	if _, err := conf.FromArgs(args); err != nil {
		fatal(err)
	}

	// Load suite
	suite, err := conf.Load()
	if err != nil {
		fatal(err)
	}

	return suite
}

//export NttListTests
func NttListTests(path * C.char) * C.char {
	str := C.GoString(path)
	args := []string{ str }

	suite := loadSuite(args,loader.Config{
		IgnoreImports:  true,
		IgnoreTags:     true,
		IgnoreComments: true,
	} )

	var tests strings.Builder

	for _, test := range suite.Tests() {
		tests.WriteString(test.FullName())
		tests.WriteString("\n")
	}
	return C.CString(tests.String())
}
//export NttListImports
func NttListImports(path * C.char) * C.char {
	args := []string{ C.GoString(path)}

	suite := loadSuite(args, loader.Config{
		IgnoreTags:     true,
		IgnoreComments: true,
	})

	var imports strings.Builder
	for _, mod := range suite.Modules() {
		for _, imp := range mod.Imports {
			imports.WriteString(mod.Name() + "\t" + imp.ImportedModule()+"\n")
		}
	}

	return C.CString(imports.String())
}

type M map[string]interface{}

//export NttLoadSuite
func NttLoadSuite(path * C.char) * C.char {
	args := []string{ C.GoString(path)}
	suite := loadSuite(args, loader.Config{
		IgnoreTags: false,
		IgnoreImports: false,
		IgnoreComments: true,
	})

	var testcases []M

	for _, test := range suite.Tests() {
		var t = make(M)
		t["mod"] = test.Module()
		t["tags"] = test.Tags()
		t["name"] = test.FullName()

		testcases = append(testcases, t)
	}

	data, err := json.MarshalIndent(testcases, "", "  ")
	if err != nil {
		fatal(err)
	}

	return C.CString(string(data))
}

func main() { }

