package auth

import (
	"app/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"log"
	"net/netip"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
	ErrBadRequest   = errors.New("bad request")
)

type TokenCheckInput struct{ UserID int64 }
type TokenCheckResult struct{ AccountType string }

func (s *Service) TokenCheck(ctx context.Context, i *TokenCheckInput) (*TokenCheckResult, error) {
	var res TokenCheckResult
	if err := s.DB.QueryRowContext(ctx, `SELECT account_type FROM portal_users WHERE id = $1`, i.UserID).Scan(&res.AccountType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	return &res, nil
}

type LogoutInput struct{ Token string }

func (s *Service) Logout(ctx context.Context, i *LogoutInput) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM portal_sessions WHERE token_hash = $1`, helpers.MakeHash256(i.Token))
	return err
}

type SessionInput struct {
	ForwardedFor string
	SourceIP     string
	UserAgent    string
}
type ActivateInput struct {
	Token          string
	NewPassword    string
	InvalidPayload bool
	Session        SessionInput
}

func (s *Service) Activate(ctx context.Context, i *ActivateInput) (*LoginResult, error) {
	tokenHash := helpers.MakeHash256(i.Token)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get user ID by token hash
	var userID int64
	var expiresAt time.Time
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM portal_account_activation_tokens
		WHERE token_hash = $1
		RETURNING portal_user_id, expires_at
	`, tokenHash).Scan(&userID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM portal_account_activation_tokens
		WHERE portal_user_id = $1
	`, userID); err != nil {
		log.Println(err)
		return nil, err
	}

	if expiresAt.Before(time.Now()) || expiresAt.Equal(time.Now()) {
		return nil, ErrExpiredToken
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	if len(i.NewPassword) < 8 {
		return nil, ErrUnprocessableEntity
	}

	// Hash new password
	passHash, err := argon2id.CreateHash(i.NewPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Update user
	if _, err := tx.ExecContext(ctx, `
		UPDATE portal_users
		SET
		    password_hash = $1,
		    account_active = true
		WHERE id = $2
	`, passHash, userID); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return s.completeLogin(ctx, userID, i.Session)
}
func (s *Service) completeLogin(ctx context.Context, userID int64, i SessionInput) (*LoginResult, error) {
	token, err := gonanoid.New(128)
	if err != nil {
		return nil, err
	}
	var forwarded, source *string
	if _, err := netip.ParseAddr(i.ForwardedFor); err == nil {
		forwarded = &i.ForwardedFor
	}
	if _, err := netip.ParseAddr(i.SourceIP); err == nil {
		source = &i.SourceIP
	}
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO portal_sessions (token_hash, portal_user_id, x_forwarded_for, source_ip, user_agent) VALUES ($1, $2, $3, $4, $5)`, helpers.MakeHash256(token), userID, forwarded, source, i.UserAgent); err != nil {
		return nil, err
	}
	return &LoginResult{Token: token}, nil
}
