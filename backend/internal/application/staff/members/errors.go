package members

import "errors"

var ErrForbidden = errors.New("forbidden")

var ErrInvalidToken = errors.New("invalid token")

var ErrNotFound = errors.New("not found")

var ErrExpired = errors.New("expired")

var ErrInvalidEmail = errors.New("invalid email")

var ErrInvalidMailImportance = errors.New("invalid mail importance")

var ErrStaffMemberEmailConflict = errors.New("staff member email conflict")

var ErrInvitationEmailConflict = errors.New("invitation email conflict")

var ErrTargetPrivacyRestricted = errors.New("target privacy restricted")

var ErrBadRequest = errors.New("bad request")
