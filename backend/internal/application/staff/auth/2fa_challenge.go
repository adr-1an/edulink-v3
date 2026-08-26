package auth

import (
	helpers2 "app/internal/helpers"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
)

type CompleteChallengeInput struct {
	Code           string
	ChallengeToken string
}

type CompleteChallengeResultType int

const (
	CompleteChallengeLogin CompleteChallengeResultType = iota
	CompleteChallengeSchoolDeletion
)

type CompleteChallengeResult struct {
	Type CompleteChallengeResultType

	SessionToken   string
	SchoolDeletion *SchoolDeletionChallengeMetadata
}

type SchoolDeletionChallengeMetadata struct {
	UserID   int64 `json:"userId"`
	SchoolID int64 `json:"schoolId"`
}

func (s *Service) CompleteChallenge(
	ctx context.Context,
	i *CompleteChallengeInput,
) (*CompleteChallengeResult, error) {
	i.Code = strings.TrimSpace(i.Code)

	if len(i.Code) != 6 {
		return nil, ErrInvalidCode
	}

	tokenHash := helpers2.MakeHash256(i.ChallengeToken)

	var (
		id         int64
		userID     int64
		purpose    TwoFactorChallengePurpose
		dbMetadata json.RawMessage
		expiresAt  time.Time
	)

	err := s.DB.QueryRowContext(ctx, `
        SELECT id, user_id, purpose, metadata, expires_at
        FROM two_factor_challenges
        WHERE token_hash = $1
    `, tokenHash).Scan(
		&id,
		&userID,
		&purpose,
		&dbMetadata,
		&expiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if expiresAt.Before(time.Now()) {
		return nil, ErrExpiredChallenge
	}

	switch purpose {
	case ChallengePurposeLogin,
		ChallengePurposeSchoolDeletion:
		// Supported, continue

	default:
		return nil, ErrUnsupportedChallengePurpose
	}

	if err := s.verifyTFA(
		ctx,
		userID,
		i.Code,
		id,
		tokenHash,
	); err != nil {
		return nil, err
	}

	switch purpose {
	case ChallengePurposeLogin:
		var meta LoginChallengeMetadata

		if err := json.Unmarshal(dbMetadata, &meta); err != nil {
			log.Println(err)
			return nil, err
		}

		token, err := s.completeLogin(ctx, meta)
		if err != nil {
			return nil, err
		}

		return &CompleteChallengeResult{
			Type:         CompleteChallengeLogin,
			SessionToken: token,
		}, nil

	case ChallengePurposeSchoolDeletion:
		var meta SchoolDeletionChallengeMetadata

		if err := json.Unmarshal(dbMetadata, &meta); err != nil {
			log.Println(err)
			return nil, err
		}

		return &CompleteChallengeResult{
			Type:           CompleteChallengeSchoolDeletion,
			SchoolDeletion: &meta,
		}, nil

	default:
		return nil, ErrUnsupportedChallengePurpose
	}
}

func (s *Service) verifyTFA(
	ctx context.Context,
	userID int64,
	code string,
	challengeID int64,
	tokenHash []byte,
) error {
	var secretKey []byte

	err := s.DB.QueryRowContext(ctx, `
        SELECT totp_secret
        FROM users
        WHERE id = $1
    `, userID).Scan(&secretKey)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		log.Println(err)
		return err
	}

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		return err
	}

	secret, err := helpers2.DecryptString(secretKey, appKey)
	if err != nil {
		log.Println(err)
		return err
	}

	if !totp.Validate(code, secret) {
		return ErrInvalidTwoFactorCode
	}

	var consumedID int64

	err = s.DB.QueryRowContext(ctx, `
        DELETE FROM two_factor_challenges
        WHERE id = $1
          AND token_hash = $2
          AND expires_at > NOW()
        RETURNING id
    `, challengeID, tokenHash).Scan(&consumedID)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
