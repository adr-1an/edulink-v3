package members

import (
	"context"
	"log"
	"time"
)

type InvitationUser struct {
	ID    string
	Name  string
	Email string
}
type InvitationSchool struct {
	ID         string
	Name       string
	RegionCode string
}
type UserInvitation struct {
	ID        string
	School    InvitationSchool
	SentBy    InvitationUser
	Status    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
type ListUserInvitationsResult struct {
	Invitations []UserInvitation
}
type ListUserInvitationsInput struct {
	UserID int64
}

func (s *Service) ListUserInvitations(ctx context.Context, i *ListUserInvitationsInput) (*ListUserInvitationsResult, error) {
	userID := i.UserID
	// UserInvitation struct

	var invitations []UserInvitation

	// Get data from DB
	rows, err := s.DB.QueryContext(ctx, `
		WITH uid AS (
		    SELECT email FROM users WHERE id = $1
		)
		SELECT
		    si.id, s.id, s.name, s.region_code, u.id, u.name,
		    u.email, si.status, si.created_at, si.expires_at
		FROM staff_invitations si
		JOIN schools s ON si.school_id = s.id
		JOIN users u ON si.sent_by_user = u.id
		WHERE si.user_email = (SELECT email FROM uid)
		AND si.status = 'pending'
		AND si.expires_at > NOW()
	`, userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var i UserInvitation

		if err := rows.Scan(
			&i.ID,
			&i.School.ID,
			&i.School.Name,
			&i.School.RegionCode,
			&i.SentBy.ID,
			&i.SentBy.Name,
			&i.SentBy.Email,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		invitations = append(invitations, i)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListUserInvitationsResult{Invitations: invitations}, nil
}
