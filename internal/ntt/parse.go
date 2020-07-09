package ntt

import (
	"context"
	"crypto/sha1"
	"fmt"
	"runtime"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

// Limits the number of parallel parser calls per process.
var parseLimit = make(chan struct{}, runtime.NumCPU())

// ParseInfo represents a memoized syntax tree of a .ttcn3 file
type ParseInfo struct {

	// Module is the syntax tree of a module. Note, multiple modules in a
	// single file is prohibited.
	Module *ast.Module

	// Err is the error result of last parse.
	Err error

	// FileSet is required decode position tags into file:line:column triplets.
	FileSet *loc.FileSet

	handle *memoize.Handle
}

func (info *ParseInfo) id() string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprint(info.Module))))
}

// Position decodes a Pos tag into a Position. If file has not been parsed
// before, an empty Position is returned.
func (info *ParseInfo) Position(pos loc.Pos) loc.Position {
	if info.FileSet == nil {
		return loc.Position{}
	}

	return info.FileSet.Position(pos)
}

// Pos encodes a line and column tuple into a offset-based Pos tag. If file nas
// not been parsed yet, Pos will return loc.NoPos.
func (info *ParseInfo) Pos(line int, column int) loc.Pos {
	if info.FileSet == nil {
		return loc.NoPos
	}

	// We asume every FileSet has only one file, to make this work.
	return loc.Pos(int(info.FileSet.File(loc.Pos(1)).LineStart(line)) + column - 1)
}

// Parse returns the cached TTCN-3 syntax of the file. The actual TTCN-3 parser is
// called for every unique file exactly once.
func (suite *Suite) Parse(file string) *ParseInfo {
	f := suite.File(file)
	f.handle = suite.store.Bind(f.id(), func(ctx context.Context) interface{} {
		data := ParseInfo{}

		b, err := f.Bytes()
		if err != nil {
			data.Err = err
			return &data
		}

		parseLimit <- struct{}{}
		defer func() { <-parseLimit }()

		var mods []*ast.Module
		data.FileSet = loc.NewFileSet()
		mods, data.Err = parser.ParseModules(data.FileSet, f.Path(), b, 0)

		// It's easier to support only one module per file.
		if len(mods) == 1 {
			data.Module = mods[0]
		} else if len(mods) > 1 {
			data.Err = fmt.Errorf("file %q contains more than one module.", f.String())
		}
		return &data
	})

	v := f.handle.Get(context.TODO())
	return v.(*ParseInfo)
}
