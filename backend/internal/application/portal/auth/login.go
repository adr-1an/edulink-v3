package auth

import (
	"app/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type LoginInput struct {
	Email    string
	Password string
}

func (i *LoginInput) validateInput() error {
	if i.Email == "" || len(i.Email) < 5 || len(i.Email) > 254 {
		return ErrUnprocessableEntity
	}

	return nil
}

type LoginResult struct {
	Token string
}

func (s *Service) Login(ctx context.Context, i *LoginInput) (*LoginResult, error) {
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))

	if err := i.validateInput(); err != nil {
		return nil, err
	}

	var userID int64
	var passwordHash string
	var loginAllowed bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT id, COALESCE(password_hash, ''), account_enabled
		FROM portal_users
		WHERE email = $1 AND account_active = true
	`, i.Email).Scan(
		&userID,
		&passwordHash,
		&loginAllowed,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		log.Println(err)
		return nil, err
	}

	if !loginAllowed {
		return nil, ErrUnauthorized
	}

	match, err := argon2id.ComparePasswordAndHash(i.Password, passwordHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !match {
		return nil, ErrUnauthorized
	}

	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tokenHash := helpers.MakeHash256(token)

	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO portal_sessions (token_hash, portal_user_id)
		VALUES ($1, $2)
	`, tokenHash, userID); err != nil {
		log.Println(err)
		return nil, err
	}

	return &LoginResult{
		Token: token,
	}, nil
}
