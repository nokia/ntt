package runtime

import "errors"

var (
	ErrExists          = errors.New("already exists")
	ErrRange           = errors.New("value out of range")
	ErrTypeMismatch    = errors.New("type mismatch")
	ErrInvalidArgCount = errors.New("invalid argument count")
	ErrNotImplemented  = errors.New("not implemented")
)
