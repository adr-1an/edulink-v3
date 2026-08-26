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

	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type LoginInput struct {
	Email        string
	Password     string
	StayLoggedIn bool
	IP           *string
	UserAgent    string
}

type LoginResultType int

const (
	LoginResultSuccess LoginResultType = iota
	LoginResultTwoFactor
)

type LoginResult struct {
	Type LoginResultType

	Token              string
	TwoFactorChallenge *TwoFactorResult
}

type TwoFactorResult struct {
	Token     string
	Purpose   TwoFactorChallengePurpose
	ExpiresAt time.Time
}

type LoginChallengeMetadata struct {
	UserID       int64
	StayLoggedIn bool
	IP           *string
	UserAgent    string
}

func (s *Service) Login(
	ctx context.Context,
	i *LoginInput,
) (*LoginResult, error) {
	i.Email = strings.TrimSpace(strings.ToLower(i.Email))

	if i.Email == "" || len(i.Email) < 5 || len(i.Email) > 254 {
		return nil, ErrUnprocessableEntity
	}

	var userID int64
	var passwordHash string

	err := s.DB.QueryRowContext(ctx, `
        SELECT id, password_hash
        FROM users
        WHERE email = $1
    `, i.Email).Scan(&userID, &passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnauthorized
		}

		log.Println(err)
		return nil, err
	}

	match, err := argon2id.ComparePasswordAndHash(i.Password, passwordHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !match {
		return nil, ErrUnauthorized
	}

	var tfa string

	err = s.DB.QueryRowContext(ctx, `
        SELECT two_factor_status
        FROM users
        WHERE id = $1
    `, userID).Scan(&tfa)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	meta := LoginChallengeMetadata{
		UserID:       userID,
		StayLoggedIn: i.StayLoggedIn,
		IP:           i.IP,
		UserAgent:    i.UserAgent,
	}

	if tfa == TwoFAStatusEnabled {
		token, err := gonanoid.New(128)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		tokenHash := helpers2.MakeHash256(token)

		id, err := s.Sf.NextID()
		if err != nil {
			log.Println(err)
			return nil, err
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
			ChallengePurposeLogin,
			tokenHash,
			expiresAt,
			metadata,
		)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		return &LoginResult{
			Type: LoginResultTwoFactor,
			TwoFactorChallenge: &TwoFactorResult{
				Token:     token,
				Purpose:   ChallengePurposeLogin,
				ExpiresAt: expiresAt,
			},
		}, nil
	}

	token, err := s.completeLogin(ctx, meta)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Type:  LoginResultSuccess,
		Token: token,
	}, nil
}

func (s *Service) completeLogin(
	ctx context.Context,
	meta LoginChallengeMetadata,
) (string, error) {
	sessionToken, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return "", err
	}

	sessionTokenHash := helpers2.MakeHash256(sessionToken)

	now := time.Now()

	var expiresAt time.Time
	if meta.StayLoggedIn {
		expiresAt = now.AddDate(0, 3, 0)
	} else {
		expiresAt = now.AddDate(0, 0, 1)
	}

	_, err = s.DB.ExecContext(ctx, `
        INSERT INTO sessions (
            token_hash,
            user_id,
            created_from_ip,
            user_agent,
            expires_at
        )
        VALUES ($1, $2, $3, $4, $5)
    `,
		sessionTokenHash,
		meta.UserID,
		meta.IP,
		meta.UserAgent,
		expiresAt,
	)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return sessionToken, nil
}
