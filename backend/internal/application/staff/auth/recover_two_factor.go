package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"

	helpers2 "app/internal/helpers"
)

type RecoverTwoFactorInput struct{ RecoveryCode string }

func (s *Service) RecoverTwoFactor(ctx context.Context, i *RecoverTwoFactorInput) error {

	codeHash := helpers2.MakeHash256(i.RecoveryCode)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT user_id FROM two_factor_recovery_codes WHERE recovery_code_hash = $1
	`, codeHash).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else {
			log.Println(err)
			return err
		}
	}

	// Disable 2FA for the user
	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET totp_secret = NULL, two_factor_status = $1
		WHERE id = $2
		AND two_factor_status = $3
	`, TwoFAStatusDisabled, userID, TwoFAStatusEnabled); err != nil {
		log.Println(err)
		return err
	}

	// Delete all the user's recovery codes
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM two_factor_recovery_codes WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
