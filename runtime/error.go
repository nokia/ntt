package runtime

import "fmt"

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Type() ObjectType {
	return RUNTIME_ERROR
}

func (e *Error) Inspect() string {
	return fmt.Sprintf("Error: %s", e.Error())
}

func Errorf(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func IsError(v interface{}) bool {
	_, ok := v.(*Error)
	return ok
}
