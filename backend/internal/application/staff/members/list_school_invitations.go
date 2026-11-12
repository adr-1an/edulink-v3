package members

import (
	"context"
	"log"
	"time"

	"app/internal/helpers/staff/schools"
)

type SchoolInvitationUser struct {
	ID    string
	Name  string
	Email string
}
type SchoolInvitation struct {
	ID        string
	UserEmail string
	SentBy    SchoolInvitationUser
	Status    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
type ListSchoolInvitationsResult struct {
	Access      schools.AccessData
	Invitations []SchoolInvitation
	SchoolID    int64
}
type ListSchoolInvitationsInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) ListSchoolInvitations(ctx context.Context, i *ListSchoolInvitationsInput) (*ListSchoolInvitationsResult, error) {
	userID := i.UserID
	schoolID := i.SchoolID
	if !schools.Can(schools.PermissionSchoolInviteList, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	var invites []SchoolInvitation

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			u.id, u.name, u.email,
			i.id, i.user_email, i.status, i.created_at, i.expires_at
		FROM staff_invitations i
		JOIN users u
			ON u.id = i.sent_by_user
		WHERE i.school_id = $1
		AND status = 'pending'
	`, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var i SchoolInvitation

		if err := rows.Scan(
			&i.SentBy.ID,
			&i.SentBy.Name,
			&i.SentBy.Email,
			&i.ID,
			&i.UserEmail,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		invites = append(invites, i)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, s.DB, userID, schoolID)
	if err != nil {
		return nil, err
	}

	return &ListSchoolInvitationsResult{Access: access, Invitations: invites, SchoolID: schoolID}, nil
}
