package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"app/internal/auth"
	"app/internal/cryptoutil"
	"app/internal/mailer"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type SendPasswordResetEmailInput struct{ Email string }

func (s *Service) SendPasswordResetEmail(ctx context.Context, i *SendPasswordResetEmailInput) error {

	// Parse
	i.Email = strings.TrimSpace(strings.ToLower(i.Email))

	// Validate
	if len(i.Email) < 5 || len(i.Email) > 254 {
		return ErrUnprocessableEntity
	}

	// Get ID from email
	var userID int64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT id FROM users WHERE email = $1
	`, i.Email).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Fake 204
			return nil
		} else {
			return err
		}
	}

	// Delete all previous password reset tokens
	_, err := s.DB.ExecContext(ctx, `
		DELETE FROM verification_tokens
		WHERE user_id = $1
		AND purpose = $2
	`, userID, auth.TokenPurposePasswordReset)
	if err != nil {
		return err
	}

	// Generate token & hash
	resetToken, err := gonanoid.New(128)
	if err != nil {
		return err
	}
	resetTokenHash := cryptoutil.MakeHash256(resetToken)

	// Store in DB
	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO verification_tokens (token_hash, user_id, purpose)
		VALUES ($1, $2, $3)
	`, resetTokenHash, userID, auth.TokenPurposePasswordReset)
	if err != nil {
		return err
	}

	// Send email
	frontendURL := os.Getenv("FRONTEND_URL")
	resetURL := fmt.Sprintf("%s/auth/reset/%s", frontendURL, resetToken)

	msgBody := fmt.Sprintf("Did you request a password reset? If so, here's your link: %s", resetURL)
	msg := mailer.Mail{
		To:          i.Email,
		Subject:     "Password reset request",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceHigh,
		Body:        msgBody,
	}
	go func() {
		if err := mailer.SendMail(msg); err != nil {
			log.Println(err)
			return
		}
	}()

	return nil
}
