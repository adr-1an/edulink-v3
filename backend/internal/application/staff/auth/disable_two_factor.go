package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"app/internal/cryptoutil"

	"github.com/alexedwards/argon2id"
	"github.com/pquerna/otp/totp"
)

type DisableTwoFactorInput struct {
	UserID   int64
	TOTP     string
	Password string
}

func (s *Service) DisableTwoFactor(ctx context.Context, i *DisableTwoFactorInput) error {
	userID := i.UserID

	i.TOTP = strings.TrimSpace(i.TOTP)

	if len(i.TOTP) != 6 {
		return ErrUnprocessableEntity
	}

	// Get password hash
	var passHash string
	var totpSecret []byte
	if err := s.DB.QueryRowContext(ctx, `
		SELECT password_hash, totp_secret FROM users WHERE id = $1
	`, userID).Scan(&passHash, &totpSecret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	match, err := argon2id.ComparePasswordAndHash(i.Password, passHash)
	if err != nil {
		log.Println(err)
		return err
	}

	if !match {
		return ErrUnauthorized
	}

	// Check TOTP
	appKey, err := cryptoutil.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		return err
	}

	secret, err := cryptoutil.DecryptString(totpSecret, appKey)
	if err != nil {
		log.Println(err)
		return err
	}

	if !totp.Validate(i.TOTP, secret) {
		return ErrUnauthorized
	}

	// Disable 2FA
	if _, err := s.DB.ExecContext(ctx, `
		UPDATE users
		SET two_factor_status = $1,
		    totp_secret = NULL
		WHERE id = $2
	`, TwoFAStatusDisabled, userID); err != nil {
		log.Println(err)
		return err
	}

	// Delete recovery codes
	if _, err := s.DB.ExecContext(ctx, `
		DELETE FROM two_factor_recovery_codes
		WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
