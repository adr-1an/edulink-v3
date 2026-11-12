package profile

import "errors"

var (
	ErrInvalidData       = errors.New("invalid data")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
)
