package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	schoolService "app/internal/application/staff/schools"
	"app/internal/cryptoutil"
	"app/internal/permissions"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type RequestSchoolDeletionInput struct {
	UserID   int64
	SchoolID int64
}
type SchoolDeletionChallenge struct {
	Token     string
	Purpose   TwoFactorChallengePurpose
	ExpiresAt time.Time
}
type RequestSchoolDeletionResult struct{ Challenge *SchoolDeletionChallenge }

func (s *Service) RequestSchoolDeletion(ctx context.Context, i *RequestSchoolDeletionInput) (*RequestSchoolDeletionResult, error) {
	userID, schoolID := i.UserID, i.SchoolID

	// Permission check
	if !permissions.Can(
		permissions.SchoolOwner,
		userID,
		schoolID,
		ctx,
		s.DB,
	) {
		return nil, ErrForbidden
	}

	// Check whether 2FA is enabled
	var tfa string

	err := s.DB.QueryRowContext(ctx, `
		SELECT two_factor_status
		FROM users
		WHERE id = $1
	`, userID).Scan(&tfa)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if tfa == TwoFAStatusEnabled {
		token, err := gonanoid.New(128)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		tokenHash := cryptoutil.MakeHash256(token)

		id, err := s.Sf.NextID()
		if err != nil {
			log.Println(err)
			return nil, err
		}

		meta := SchoolDeletionChallengeMetadata{
			UserID:   userID,
			SchoolID: schoolID,
		}

		metadata, err := json.Marshal(meta)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		expiresAt := time.Now().Add(15 * time.Minute)

		_, err = s.DB.ExecContext(ctx, `
			INSERT INTO two_factor_challenges (
				id,
				user_id,
				purpose,
				token_hash,
				expires_at,
				metadata
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			id,
			userID,
			ChallengePurposeSchoolDeletion,
			tokenHash,
			expiresAt,
			metadata,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return &RequestSchoolDeletionResult{Challenge: &SchoolDeletionChallenge{Token: token, Purpose: ChallengePurposeSchoolDeletion, ExpiresAt: expiresAt}}, nil
	}

	// No 2FA required -> delete immediately
	err = s.Schools.DeleteSchool(
		ctx,
		&schoolService.DeleteSchoolInput{
			UserID:   userID,
			SchoolID: schoolID,
		},
	)
	if err != nil {
		return nil, err
	}

	return &RequestSchoolDeletionResult{}, nil
}
