package profile

import (
	"context"
	"log"
	"time"
)

type GetProfileInput struct{ UserID int64 }

type ProfilePicture struct {
	PresignedURL string
	Filename     string
	ContentType  string
}

type GetProfileResult struct {
	Pfp             *ProfilePicture
	ID              int64
	Name            string
	Email           string
	Phone           string
	TwoFactorStatus string
	PublicProfile   bool
	InvDisabled     bool
	UpdatedAt       time.Time
	CreatedAt       time.Time
}

func (s *Service) GetProfile(ctx context.Context, i *GetProfileInput) (*GetProfileResult, error) {
	userID := i.UserID
	var err error

	// Get user data
	var user GetProfileResult

	var (
		bucketName  *string
		objKey      *string
		filename    *string
		contentType *string
	)

	if err = s.DB.QueryRowContext(ctx, `
		SELECT
		    so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type,
		    u.id, u.name, u.email, u.phone, u.two_factor_status, u.public_profile, u.staff_invitations_disabled, u.updated_at, u.created_at
		FROM users u
		LEFT JOIN user_profile_pictures pfps
		ON u.id = pfps.user_id
		LEFT JOIN storage_objects so
		ON so.id = pfps.storage_object_id
		WHERE u.id = $1
		`, userID).
		Scan(
			&bucketName,
			&objKey,
			&filename,
			&contentType,
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.TwoFactorStatus,
			&user.PublicProfile,
			&user.InvDisabled,
			&user.UpdatedAt,
			&user.CreatedAt,
		); err != nil {
		log.Println(err)
		return nil, err
	}

	if filename != nil && contentType != nil && bucketName != nil {
		url, err := s.S3.PresignedGetObject(ctx, *bucketName, *objKey, 15*time.Minute, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		user.Pfp = &ProfilePicture{
			PresignedURL: url.String(),
			Filename:     *filename,
			ContentType:  *contentType,
		}
	}

	return &user, nil
}
