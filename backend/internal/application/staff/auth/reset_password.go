package auth

import (
	"context"
	"database/sql"
	"errors"

	helpers2 "app/internal/helpers"
	staffhelpers "app/internal/helpers/staff"

	"github.com/alexedwards/argon2id"
)

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

func (s *Service) ResetPassword(ctx context.Context, i *ResetPasswordInput) error {
	tokenHash := helpers2.MakeHash256(i.Token)

	// Validate
	if len(i.NewPassword) < 8 {
		return ErrUnprocessableEntity
	}

	// Get ID
	var userID int64
	err := s.DB.QueryRowContext(ctx, `
		DELETE FROM verification_tokens
	   		WHERE token_hash = $1
			AND created_at >= NOW() - INTERVAL '24 hours'
	   		AND purpose = $2
		RETURNING user_id
	`, tokenHash, staffhelpers.TokenPurposePasswordReset).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnauthorized
		} else {
			return err
		}
	}

	// Revoke all sessions
	_, err = s.DB.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return err
	}

	// Hash the new password
	newPasswdHash, err := argon2id.CreateHash(i.NewPassword, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	// Update password
	_, err = s.DB.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`, newPasswdHash, userID)
	if err != nil {
		return err
	}

	return nil
}
