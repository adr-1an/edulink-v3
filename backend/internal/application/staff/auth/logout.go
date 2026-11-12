package auth

import (
	"context"
)

type LogoutInput struct{ TokenHash []byte }

func (s *Service) Logout(ctx context.Context, i *LogoutInput) error {
	tokenHash := i.TokenHash

	// Delete token
	res, err := s.DB.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1
		AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return err
	}

	// Check rows affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUnauthorized
	}

	return nil
}
