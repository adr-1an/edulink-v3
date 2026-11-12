package profile

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"app/internal/auth"
	"app/internal/cryptoutil"
	"app/internal/mailer"

	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type SendEmailChangeInput struct {
	UserID   int64
	NewEmail string
	Password string
}

func (s *Service) SendEmailChange(ctx context.Context, i *SendEmailChangeInput) error {
	userID := i.UserID
	var err error

	// Parse
	i.NewEmail = strings.TrimSpace(strings.ToLower(i.NewEmail))

	// Validate
	if i.NewEmail == "" || len(i.NewEmail) < 5 || len(i.NewEmail) > 254 {
		return ErrInvalidData
	}

	// Get current password hash & check for email conflict
	var passwordHash string
	var emailConflict bool
	err = s.DB.QueryRowContext(ctx, `
		WITH conflict AS (
		    SELECT EXISTS (SELECT 1 FROM users WHERE email = $2) AS email_exists
		)
		SELECT u.password_hash, c.email_exists
		FROM users u, conflict c
		WHERE u.id = $1
	`, userID, i.NewEmail).Scan(&passwordHash, &emailConflict)
	if err != nil {
		log.Println(err)
		return err
	}

	if emailConflict {
		// Fake 204
		return nil
	}

	// Compare passwords
	match, err := argon2id.ComparePasswordAndHash(i.Password, passwordHash)
	if err != nil {
		log.Println(err)
		return err
	}

	if !match {
		return ErrIncorrectPassword
	}

	// Generate verification token
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return err
	}
	tokenHash := cryptoutil.MakeHash256(token)

	// Store token
	var userEmail string
	if err := s.DB.QueryRowContext(ctx, `
		WITH inserted AS (
		    INSERT INTO verification_tokens (token_hash, user_id, purpose, email_change_new_email)
		   	VALUES ($1, $2, $3, $4)
		    RETURNING user_id
		)
		SELECT u.email
		FROM users u
		JOIN inserted i ON u.id = i.user_id
	`, tokenHash, userID, auth.TokenPurposeEmailChange, i.NewEmail).Scan(&userEmail); err != nil {
		log.Println(err)
		return err
	}

	// Send verification email
	verURL := fmt.Sprintf("%s/auth/verify-change/%s", os.Getenv("FRONTEND_URL"), token)
	msg := fmt.Sprintf("Click this link to finish changing your email: %s", verURL)
	email := mailer.Mail{
		To:          i.NewEmail,
		Subject:     "Verify your Email",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceHigh,
		Body:        msg,
	}
	if err := mailer.SendMail(email); err != nil {
		log.Println(err)
		// Don't return - token was already stored
	}

	return nil
}
