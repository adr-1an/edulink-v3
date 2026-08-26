package schools

import "errors"

var (
	ErrInvalidName       = errors.New("invalid name")
	ErrInvalidRegionCode = errors.New("invalid region code")
	ErrBadRequest        = errors.New("bad request")
	ErrNoPermission      = errors.New("no permission")
	ErrForbidden         = errors.New("forbidden")
)
