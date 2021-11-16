package runtime

import "errors"

type Object interface {
	Inspect() string
	Type() ObjectType
}

type ObjectType string

const (
	UNKNOWN       ObjectType = "unknown"
	UNDEFINED                = "undefined"
	RUNTIME_ERROR            = "runtime error"
	RETURN_VALUE             = "return value"
	INTEGER                  = "integer"
	FLOAT                    = "float"
	BOOL                     = "boolean"
	STRING                   = "string"
	BITSTRING                = "bitstring"
	FUNCTION                 = "function"
	LIST                     = "list"
	BUILTIN_OBJ              = "builtin function"
	VERDICT                  = "verdict"
)

var ErrSyntax = errors.New("invalid syntax")

var (
	Undefined = &undefined{}
)

type undefined struct{}

func (u *undefined) Inspect() string  { return "undefined" }
func (u *undefined) Type() ObjectType { return UNDEFINED }
