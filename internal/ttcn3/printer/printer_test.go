package printer_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/internal/ttcn3/printer"
	"github.com/pmezard/go-difflib/difflib"
)

var todos = map[string]bool{
	//"ModuleEmpty":       true,
	"ModuleEmpty2":      true,
	"ModuleLanguage":    true,
	"Attributes":        true,
	"AttributesShort":   true,
	"Whitespaces":       true,
	"Import":            true,
	"ImportLanguage":    true,
	"ImportGroup":       true,
	"ImportExcept":      true,
	"ImportIndented":    true,
	"ImportInline":      true,
	"ImportsSorted":     true,
	"ImportSyntaxError": true,
}

func TestPrinter(t *testing.T) {
	for _, test := range tests("testdata/") {
		t.Run(test.Name(), func(t *testing.T) {
			var buf bytes.Buffer
			if err := printer.Print(&buf, test.fset, test.node); err != nil {
				t.Error(err.Error())
			}

			actual := buf.String()
			if actual != test.expected && !todos[test.Name()] {
				diff := difflib.UnifiedDiff{
					A:        difflib.SplitLines(test.expected),
					B:        difflib.SplitLines(actual),
					FromFile: "Expected",
					ToFile:   "Actual",
					Context:  3,
				}
				text, _ := difflib.GetUnifiedDiffString(diff)
				t.Errorf("%s:\n%s", test.path, text)
			} else if actual == test.expected && todos[test.Name()] {
				t.Errorf("test was expected to fail, but succeeded.")
			}

		})
	}
}

// Load tests from disk
func tests(path string) []Test {
	var tests []Test
	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {

		// Ignore file system errors
		if err != nil {
			return nil
		}

		// Ignore everything, but '**/*.expected.ttcn3'
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".expected.ttcn3") {
			return nil
		}

		expected, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		input, err := makeInput(path, expected)
		if err != nil {
			panic(err)
		}

		fset := loc.NewFileSet()
		mods, _ := parser.ParseModules(fset, path, input, parser.AllErrors)
		if len(mods) != 1 {
			panic(fmt.Sprintf("test requires %q to have exactly one module", path))
		}

		tests = append(tests, Test{
			path:     path,
			input:    string(input),
			expected: string(expected),
			fset:     fset,
			node:     mods[0],
		})
		return nil
	})
	return tests
}

func makeInput(path string, content []byte) ([]byte, error) {
	path = strings.ReplaceAll(path, ".expected.ttcn3", ".input.ttcn3")
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if err == nil {
		return ioutil.ReadFile(path)
	}

	return regexp.MustCompile(`\s+`).ReplaceAll(content, []byte(" ")), nil
}

type Test struct {
	path     string
	input    string
	expected string

	fset *loc.FileSet
	node ast.Node
}

func (t Test) Name() string {
	stem := strings.TrimSuffix(filepath.Base(t.path), ".expected.ttcn3")
	path := strings.TrimPrefix(filepath.Dir(t.path), "testdata")
	return fmt.Sprintf("%s%s", path, stem)
}
