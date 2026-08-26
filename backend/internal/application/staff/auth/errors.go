package auth

import "errors"

var (
	ErrUnauthorized                = errors.New("unauthorized")
	ErrUnprocessableEntity         = errors.New("unprocessable entity")
	ErrNotFound                    = errors.New("not found")
	ErrInvalidCode                 = errors.New("invalid code")
	ErrExpiredChallenge            = errors.New("expired challenge")
	ErrForbidden                   = errors.New("forbidden")
	ErrInvalidTwoFactorCode        = errors.New("invalid two factor code")
	ErrUnsupportedChallengePurpose = errors.New("unsupported challenge purpose")
)
