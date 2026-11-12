package scores

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrForbidden           = errors.New("forbidden")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrConflict            = errors.New("conflict")
)
