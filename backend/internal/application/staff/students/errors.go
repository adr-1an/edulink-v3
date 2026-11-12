package students

import "errors"

var (
	ErrBadRequest           = errors.New("bad request")
	ErrInvalidData          = errors.New("invalid data")
	ErrForbidden            = errors.New("forbidden")
	ErrProfilePictureExists = errors.New("profile picture already exists")
	ErrConflict             = errors.New("conflict")
)

var ErrPayloadTooLong = errors.New("too many students")

type InvalidRowError struct{ Row int }

func (e *InvalidRowError) Error() string { return "invalid student row" }

type EmailConflictError struct{ Emails []string }

func (e *EmailConflictError) Error() string { return "conflicting student emails" }
