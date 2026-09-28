package portal_users

import "errors"

var (
	ErrForbidden            = errors.New("forbidden")
	ErrAccountDisabled      = errors.New("account disabled")
	ErrAccountAlreadyActive = errors.New("already active")
)
