package members

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"app/internal/cryptoutil"
)

type Invitation struct {
	ID          string
	SentToEmail string
	SchoolName  string
	SentByName  string
	SentByEmail string
	Status      string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}
type ViewInvitationResult struct {
	Value Invitation
}
type ViewInvitationInput struct {
	Token string
}

func (s *Service) ViewInvitation(ctx context.Context, i *ViewInvitationInput) (*ViewInvitationResult, error) {
	token := i.Token
	tokenHash := cryptoutil.MakeHash256(token)

	var inv Invitation

	// Get invitation data
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
		    si.id, si.user_email, s.name, u.name, u.email, si.status, si.created_at, si.expires_at
		FROM staff_invitations si
		JOIN schools s ON s.id = si.school_id
		LEFT JOIN users u ON u.id = si.sent_by_user
		WHERE si.token_hash = $1
	`, tokenHash).
		Scan(
			&inv.ID,
			&inv.SentToEmail,
			&inv.SchoolName,
			&inv.SentByName,
			&inv.SentByEmail,
			&inv.Status,
			&inv.CreatedAt,
			&inv.ExpiresAt,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check expiry
	if inv.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpired
	}

	return &ViewInvitationResult{Value: inv}, nil
}
