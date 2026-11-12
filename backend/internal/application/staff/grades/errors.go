package grades

import "errors"

var ErrForbidden = errors.New("forbidden")

var ErrMissingLevelVar = errors.New("missing level var")

var ErrInvalidName = errors.New("invalid name")

var ErrLevelOutOfRange = errors.New("level out of range")

var ErrConflict = errors.New("conflict")

var ErrBadRequest = errors.New("bad request")
