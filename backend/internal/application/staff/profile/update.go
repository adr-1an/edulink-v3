package profile

import (
	"context"
	"log"
	"strings"
)

type UpdateProfileInput struct {
	UserID        int64
	Name          string
	Phone         string
	PublicProfile bool
	InvDisabled   bool
}

func (s *Service) UpdateProfile(ctx context.Context, i *UpdateProfileInput) error {
	userID := i.UserID
	var err error

	// Parse
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)

	// Validate
	if len(i.Name) > 64 || (len(i.Phone) < 3 && i.Phone != "") || len(i.Phone) > 64 {
		return ErrInvalidData
	}

	// Update
	if _, err = s.DB.ExecContext(ctx, `
		UPDATE users
		SET name = $1, phone = $2, public_profile = $3, staff_invitations_disabled = $4, updated_at = NOW()
		WHERE id = $5
	`, i.Name, i.Phone, i.PublicProfile, i.InvDisabled, userID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
