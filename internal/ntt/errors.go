package ntt

import (
	"github.com/nokia/ntt/internal/loc"
	errors "golang.org/x/xerrors"
)

var (
	ErrSyntax           = errors.New("syntax error")
	ErrNoSuchIdentifier = errors.New("no such identifier")
	ErrRedefinition     = errors.New("redefinition")
	ErrNoIdentFound     = errors.New("no identifier found")
)

type SyntaxError struct {
	Pos, End loc.Position
}
