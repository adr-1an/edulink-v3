package auth

import (
	"app/internal/helpers/staff"
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/alexedwards/argon2id"
)

type UpdatePasswordInput struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

func (s *Service) UpdatePassword(ctx context.Context, i *UpdatePasswordInput) error {
	// Parse & validate
	if i.OldPassword == "" || i.NewPassword == "" || len(i.NewPassword) < 8 {
		return ErrUnprocessableEntity
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var oldPassHash string
	if err := tx.QueryRowContext(ctx, `
		SELECT password_hash FROM portal_users WHERE id = $1
		AND account_active = true
		AND account_enabled = true
	`, i.UserID).Scan(&oldPassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnauthorized
		}
		log.Println(err)
		return err
	}

	// Compare passwords
	match, err := argon2id.ComparePasswordAndHash(i.OldPassword, oldPassHash)
	if err != nil {
		log.Println(err)
		return err
	}

	if !match {
		return ErrUnauthorized
	}

	// Hash new password
	newPassHash, err := argon2id.CreateHash(i.NewPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		return err
	}

	// Update password
	res, err := tx.ExecContext(ctx, `
		UPDATE portal_users SET password_hash = $1
		WHERE id = $2
	`, newPassHash, i.UserID)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := staff_helpers.RowsAffected(res); err != nil {
		log.Println(err)
		return err
	}

	return tx.Commit()
}
