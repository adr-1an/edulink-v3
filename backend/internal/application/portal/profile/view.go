package profile

import (
	helpers2 "app/internal/helpers"
	"context"
	"log"
	"time"
)

type ViewProfileInput struct {
	UserID int64
}

type ViewProfileResult struct {
	Profile profile
}

type user struct {
	Name  string
	Email string
}

type school struct {
	Owner  user
	Name   string
	Region string
}

type profile struct {
	Name              string
	LastName          string
	ProfilePictureURL *string
	Email             string
	Phone             string
	DateOfBirth       time.Time
	AccountType       string

	School school
}

func (s *Service) ViewProfile(ctx context.Context, i *ViewProfileInput) (*ViewProfileResult, error) {
	var p profile

	var pfpObjectKey *string
	var pfpBucketName *string

	if err := s.DB.QueryRowContext(ctx, `
		SELECT
		    u.name, u.email,
		    s.name, s.region_code,
		    pu.name, pu.last_name, pu.email, pu.date_of_birth, COALESCE(pu.phone, ''), account_type,
		    so.object_key, so.bucket_name
		FROM portal_users pu
		JOIN schools s
		ON pu.school_id = s.id
		JOIN users u
		ON s.owner_id = u.id
		LEFT JOIN portal_user_profile_pictures pfp ON pfp.portal_user_id = pu.id
		LEFT JOIN storage_objects so ON so.id = pfp.storage_object_id AND status = $2
		WHERE pu.id = $1
	`, i.UserID, helpers2.StatusDone).
		Scan(
			&p.School.Owner.Name,
			&p.School.Owner.Email,

			&p.School.Name,
			&p.School.Region,

			&p.Name,
			&p.LastName,
			&p.Email,
			&p.DateOfBirth,
			&p.Phone,
			&p.AccountType,

			&pfpObjectKey,
			&pfpBucketName,
		); err != nil {
		log.Println(err)
		return nil, err
	}

	if pfpObjectKey != nil && pfpBucketName != nil {
		// Generate presigned URL
		presignedUrl, err := s.S3.PresignedGetObject(ctx, *pfpBucketName, *pfpObjectKey, 5*time.Minute, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		url := presignedUrl.String()

		p.ProfilePictureURL = &url
	}

	return &ViewProfileResult{
		Profile: p,
	}, nil
}
