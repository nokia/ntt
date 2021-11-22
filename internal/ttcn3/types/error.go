package types

import (
	"fmt"

	"github.com/nokia/ntt/internal/loc"
)

type RedefinitionError struct {
	Name           string
	OldPos, NewPos loc.Position
}

func (err *RedefinitionError) Error() string {
	return fmt.Sprintf("redefinition of %q", err.Name)
}

type UnknownIdentifierError struct {
	Name string
	Pos  loc.Position
}

func (err *UnknownIdentifierError) Error() string {
	return fmt.Sprintf("unknown identifier %q", err.Name)
}

type NoFieldError struct {
	Type  string
	Field string
	Pos   loc.Position
}

func (err *NoFieldError) Error() string {
	return fmt.Sprintf("type %q has no field or method %q", err.Type, err.Field)
}

type InvalidTypeError struct {
	Actual, Expected Type
	Pos              loc.Position
}

func (err *InvalidTypeError) Error() string {
	return fmt.Sprintf("invalid type %q, expected %q", err.Actual, err.Expected)
}
