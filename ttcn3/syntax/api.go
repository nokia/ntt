/*
Package parser implements a tolerant TTCN-3 parser library.

It implements most of TTCN-3 core language specification 4.10 (2018), Advanced
Parametrisation, Behaviour Types, Performance and Realtime Testing, simplistic
preprocessor support, multi-line string literals for Titan TestPorts, and
optional semicolon for backward compatibility.
*/
package syntax

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
)

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
type Mode uint

const (
	PedanticSemicolon = 1 << iota // expect semicolons pedantically
	IgnoreComments                // ignore comments
	Trace                         // print a trace of parsed productions
	AllErrors                     // report all errors (not just the first 10 on different lines)
)

// Parse the source code of a single file and return the corresponding syntax
// tree. The source code may be provided via the filename of the source file,
// or via the src parameter.
//
// If src != nil, Parse parses the source from src and the filename is only
// used when recording position information. The type of the argument for the
// src parameter must be string, []byte, or io.Reader. If src == nil, Parse
// parses the file specified by filename.
//
// The mode parameter controls the amount of source text parsed and other
// optional parser functionality.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax errors
// were found, the result is a partial AST (with Bad* nodes representing the
// fragments of erroneous source code). Multiple errors are returned via a
// ErrorList which is sorted by file position.
func Parse(filename string, src interface{}) (root *Root, names map[string]bool, uses map[string]bool, err error) {
	fset := loc.NewFileSet()

	// get source
	text, err := readSource(filename, src)
	if err != nil {
		return nil, nil, nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse source
	p.init(fset, filename, text, AllErrors)
	for p.tok != EOF {
		p.Root.Nodes = append(p.Root.Nodes, p.parse())

		if p.tok != EOF && !topLevelTokens[p.tok] {
			p.error(p.pos(1), fmt.Sprintf("unexpected token %s", p.tok))
			break
		}

		if p.tok == COMMA || p.tok == SEMICOLON {
			p.consume()
		}

	}
	return p.Root, p.names, p.uses, nil
}

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return ioutil.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return fs.Content(filename)
}
