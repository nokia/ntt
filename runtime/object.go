package runtime

type Object interface {
	Inspect() string
	Type() ObjectType
}

type ObjectType string

const (
	UNKNOWN       ObjectType = "unknown"
	RUNTIME_ERROR            = "runtime error"
	RETURN_VALUE             = "return value"
	INTEGER                  = "integer"
	BOOL                     = "boolean"
)
