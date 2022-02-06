package lsp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	var (
		locs []protocol.Location
		file = params.TextDocument.URI
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	start := time.Now()
	defer log.Debug(fmt.Sprintf("Goto Definition took %s.", time.Since(start)))

	// Legacy Mode
	if e := os.Getenv("_NTT_USE_DB"); e == "" {
		return definitionsOld(s.Owners(file), string(file.SpanURI()), line, col), nil
	}

	// Fallback to cTags
	for _, suite := range s.Owners(file) {
		locs = append(locs, cTags(suite, string(file.SpanURI()), line, col)...)
	}
	return unifyLocs(locs), nil

}

func definitionsOld(suites []*ntt.Suite, file string, line, col int) []protocol.Location {
	var locs []protocol.Location
	for _, suite := range suites {
		id, _ := suite.DefinitionAt(file, line, col)
		log.Debugf("IdentifierInfo: %#v", id)
		if id != nil && id.Def != nil {
			locs = append(locs, location(id.Def.Position))
		} else {
			locs = append(locs, cTags(suite, file, line, col)...)
		}
	}
	return unifyLocs(locs)
}

func cTags(suite *ntt.Suite, file string, line int, col int) []protocol.Location {
	stast := time.Now()
	defer log.Debug(fmt.Sprintf("cTags fallback took %s", time.Since(stast)))

	var ret []protocol.Location

	tree := suite.Parse(file)
	if tree == nil {
		log.Debug(fmt.Sprintf("Parsing %q failed.", file))
		return nil
	}

	log.Debug(fmt.Sprintf("Parse: %+v", tree))

	// If Module is nil, there's no need to continue, because there's no
	// AST to iterate over.
	if tree.Module == nil {
		return nil
	}
	id := suite.IdentifierAt(tree, line, col)
	if id == nil {
		return nil
	}

	for _, mod := range tree.ImportedModules() {
		file, _ := suite.FindModule(mod)
		if file == "" {
			continue
		}
		tree, tags := suite.Tags(file)
		for _, t := range tags {
			if ast.Name(t) == ast.Name(id) {
				ret = append(ret, location(tree.Position(t.Pos())))
			}
		}
	}
	return ret
}

func unifyLocs(locs []protocol.Location) []protocol.Location {
	m := make(map[protocol.Location]bool)
	for _, loc := range locs {
		m[loc] = true
	}

	ret := make([]protocol.Location, 0, len(m))
	for loc := range m {
		ret = append(ret, loc)
	}
	return ret
}
