package uploads

import "errors"

var (
	ErrForbidden           = errors.New("forbidden")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternal            = errors.New("internal error")
)
