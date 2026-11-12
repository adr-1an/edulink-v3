package attachments

import "errors"

var (
	ErrBadRequest          = errors.New("bad request")
	ErrForbidden           = errors.New("forbidden")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInternal            = errors.New("internal error")
)
