package syntax

import (
	"github.com/nokia/ntt/internal/loc"
)

func Begin(n Node) loc.Position {
	if tok := n.FirstTok(); tok != nil {
		return tok.(*tokenNode).file.Position(tok.Pos())
	}
	return loc.Position{}
}

func End(n Node) loc.Position {
	if tok := n.LastTok(); tok != nil {
		return tok.(*tokenNode).file.Position(tok.End())
	}
	return loc.Position{}
}
