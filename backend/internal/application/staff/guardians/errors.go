package guardians

import "errors"

var (
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrForbidden           = errors.New("forbidden")
	ErrNoRowsInserted      = errors.New("no rows inserted")
)
