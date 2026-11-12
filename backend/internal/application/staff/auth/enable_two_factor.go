package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"

	helpers2 "app/internal/helpers"

	"github.com/pquerna/otp/totp"
)

type EnableTwoFactorInput struct{ UserID int64 }
type EnableTwoFactorResult struct{ URL string }

func (s *Service) EnableTwoFactor(ctx context.Context, i *EnableTwoFactorInput) (*EnableTwoFactorResult, error) {
	userID := i.UserID

	// Check if 2FA is already enabled
	var email string
	var status string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT email, two_factor_status FROM users WHERE id = $1
	`, userID).Scan(&email, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if status == TwoFAStatusEnabled {
		return nil, ErrConflict
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		log.Println("APP_NAME environment variable not set.")
		appName = "-"
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: email,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	secret := key.Secret()
	url := key.URL()

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Encrypt secret
	encSecret, err := helpers2.EncryptString(secret, appKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Store secret
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET totp_secret = $1,
		    two_factor_status = $2
		WHERE id = $3
	`, encSecret, TwoFAStatusPending, userID); err != nil {
		log.Println(err)
		return nil, err
	}

	return &EnableTwoFactorResult{URL: url}, nil
}
