package roles

import "errors"

var ErrForbidden = errors.New("forbidden")

var ErrUnprocessableEntity = errors.New("unprocessable entity")

var ErrConflict = errors.New("conflict")

var ErrBadRequest = errors.New("bad request")
