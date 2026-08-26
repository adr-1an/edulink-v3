package auth

import (
	helpers2 "app/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type SendRegistrationLinkInput struct {
	Email string
}

func (s *Service) SendRegistrationLink(ctx context.Context, i *SendRegistrationLinkInput) error {
	// Parse
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))

	// Validate
	if i.Email == "" || len(i.Email) < 5 || len(i.Email) > 254 {
		return ErrUnprocessableEntity
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Println(err)
		}
	}()

	// Check for email conflict
	var emailConflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM users u
		 	WHERE u.email = $1
		)
	`, i.Email).Scan(&emailConflict); err != nil {
		return err
	}
	if emailConflict {
		// Email is already taken, fake ok response
		return nil
	}

	// Check if an active token already exists
	var validTokenExists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM registration_tokens rt
		 	WHERE rt.email = $1
		 	AND rt.expires_at > NOW()
		)
	`, i.Email).Scan(&validTokenExists); err != nil {
		return err
	}
	if validTokenExists {
		// A valid token already exists, fake ok response
		return nil
	}

	// Generate token & hash
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return err
	}
	tokenHash := helpers2.MakeHash256(token)

	// Store registration token
	res, err := s.DB.ExecContext(ctx, `
		INSERT INTO registration_tokens (token_hash, email, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
		ON CONFLICT (email) DO UPDATE
		    SET
		        token_hash = EXCLUDED.token_hash,
		        expires_at = EXCLUDED.expires_at,
		        created_at = NOW()
		WHERE registration_tokens.expires_at <= NOW()
	`, tokenHash, i.Email)
	if err != nil {
		log.Println(err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}

	if affected == 0 {
		// Valid token with this email exists, fake ok response
		return nil
	}

	// Send email
	frontendURL := os.Getenv("FRONTEND_URL")
	appName := os.Getenv("APP_NAME")
	registrationURL := fmt.Sprintf("%s/auth/register/%s", frontendURL, token)
	msg := fmt.Sprintf("Hi! Let's complete your %s registration: %s", appName, registrationURL)
	email := helpers2.Mail{
		To:          i.Email,
		Subject:     "Complete your registration",
		Body:        msg,
		Importance:  mail.ImportanceHigh,
		ContentType: mail.TypeTextPlain,
	}

	// Send email
	go func() {
		if err := helpers2.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type CheckRegistrationTokenInput struct {
	Token string
}

type CheckRegistrationTokenResult struct {
	Email string
}

func (s *Service) CheckRegistrationToken(ctx context.Context, i *CheckRegistrationTokenInput) (*CheckRegistrationTokenResult, error) {
	tokenHash := helpers2.MakeHash256(i.Token)

	// Get the email from the token
	var email string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT email
		FROM registration_tokens
		WHERE token_hash = $1
		AND expires_at > NOW()
	`, tokenHash).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		log.Println(err)
		return nil, err
	}

	return &CheckRegistrationTokenResult{
		Email: email,
	}, nil
}

type RegisterInput struct {
	Name              string
	Phone             string
	Password          string
	RegistrationToken string
}

func (s *Service) Register(ctx context.Context, i *RegisterInput) error {
	// Parse
	i.Name = strings.TrimSpace(i.Name)
	i.Phone = strings.TrimSpace(i.Phone)

	// Validate

	// Name
	if i.Name == "" ||
		len(i.Name) < 1 ||
		len(i.Name) > 128 ||
		// Password
		i.Password == "" ||
		len(i.Password) < 8 {
		return ErrUnprocessableEntity
	}

	// Phone
	if i.Phone != "" {
		if len(i.Phone) < 3 || len(i.Phone) > 32 {
			return ErrUnprocessableEntity
		}
	}

	// Hash token
	tokenHash := helpers2.MakeHash256(i.RegistrationToken)

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Println(err)
		}
	}()

	// Get email from token
	var email string
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM registration_tokens
		WHERE token_hash = $1
		AND expires_at > NOW()
		RETURNING email
	`, tokenHash).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		log.Println(err)
		return err
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return err
	}

	// Hash password
	passwordHash, err := argon2id.CreateHash(i.Password, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		return err
	}

	// Store user
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO users (id, name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4, $5)
	`, id, i.Name, email, i.Phone, passwordHash)
	if err != nil {
		log.Println(err)
		return err
	}

	// Commit tx
	if err = tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
