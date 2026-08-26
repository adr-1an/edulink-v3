package auth

import (
	"app/internal/helpers"
	"context"
	"log"
)

type CheckTokenInput struct {
	Token string
}

func (s *Service) CheckToken(ctx context.Context, i *CheckTokenInput) (bool, error) {
	tokenHash := helpers.MakeHash256(i.Token)

	// Check if token is valid
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM sessions
			WHERE token_hash = $1
			AND expires_at > NOW()
			AND revoked_at IS NULL
		)
	`, tokenHash).Scan(&exists); err != nil {
		log.Println(err)
		return false, err
	}

	if !exists {
		return false, ErrUnauthorized
	}

	return exists, nil
}
