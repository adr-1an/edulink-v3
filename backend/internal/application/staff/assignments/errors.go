package assignments

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrForbidden           = errors.New("forbidden")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
)
