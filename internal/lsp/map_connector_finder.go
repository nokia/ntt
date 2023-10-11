package lsp

import (
	"sort"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
)

func isDuplicate(list []protocol.Location, loc protocol.Location) bool {
	for _, l := range list {
		if l == loc {
			return true
		}
	}
	return false
}
func invokedFromMapOrConnect(n syntax.Node, syn *ttcn3.Tree, list []protocol.Location) []protocol.Location {
	p := n

	for {
		p = syn.ParentOf(p)
		if p == nil {
			break
		}
		node, ok := p.(*syntax.CallExpr)
		if !ok {
			continue
		}
		idNode, ok := node.Fun.(*syntax.Ident)
		if !ok {
			continue
		}
		if idNode.String() == "map" || idNode.String() == "connect" {
			loc := location(syntax.Span{Begin: syn.Position(node.Pos()), End: syn.Position(node.End()), Filename: syn.Filename()})
			if !isDuplicate(list, loc) {
				list = append(list, loc)
			}
			break
		}
	}
	return list
}

func findMapConnectStatementForPortIdMatchingTheNameFromFile(file string, idName string) []protocol.Location {
	list := make([]protocol.Location, 0, 4)
	syn := ttcn3.ParseFile(file)
	syn.Root.Inspect(func(n syntax.Node) bool {
		if n == nil {
			// called on node exit
			return false
		}

		switch node := n.(type) {
		case *syntax.Ident:
			if idName == node.String() {
				list = invokedFromMapOrConnect(n, syn, list)
			}
			return false
		default:
			return true
		}
	})
	return list
}

func FindMapConnectStatementForPortIdMatchingTheName(db *ttcn3.DB, name string) []protocol.Location {
	candidates := make([]string, 0, len(db.Uses))
	locs := make([]protocol.Location, 0, len(db.Uses))
	for file := range db.Uses[name] {
		candidates = append(candidates, file)
	}
	sort.Strings(candidates)
	for _, file := range candidates {
		locs = append(locs, findMapConnectStatementForPortIdMatchingTheNameFromFile(file, name)...)
	}
	return locs
}
