package auth

import "errors"

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
)
