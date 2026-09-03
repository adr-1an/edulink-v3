package students

import "errors"

var (
	ErrInvalidData          = errors.New("invalid data")
	ErrForbidden            = errors.New("forbidden")
	ErrProfilePictureExists = errors.New("profile picture already exists")
)
