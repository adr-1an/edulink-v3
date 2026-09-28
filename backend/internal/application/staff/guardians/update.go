package guardians

import (
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
)

type UpdateGuardian struct {
	Name           string
	LastName       string
	Email          string
	Phone          *string
	Notes          *string
	DateOfBirth    time.Time
	AccountEnabled bool
	AccountActive  bool
	Password       *string
}

type UpdateGuardianInput struct {
	UserID     int64
	GuardianID int64
	Guardian   UpdateGuardian
}

func (s *Service) UpdateGuardian(ctx context.Context, i *UpdateGuardianInput) error {
	var schoolID int64
	var accountActive bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id, account_active
		FROM portal_users
		WHERE id = $1
	`, i.GuardianID).Scan(&schoolID, &accountActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		log.Println(err)
		return err
	}

	if !schools.Can(schools.PermissionGuardianUpdate, i.UserID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}
	if permission, changed := schools.PortalUserActiveStatePermission(accountActive, i.Guardian.AccountActive); changed &&
		!schools.Can(permission, i.UserID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	// Parse & validate
	i.Guardian.Name = strings.TrimSpace(i.Guardian.Name)
	i.Guardian.LastName = strings.TrimSpace(i.Guardian.LastName)
	i.Guardian.Email = strings.TrimSpace(strings.ToLower(i.Guardian.Email))
	var guardianPhone string
	if i.Guardian.Phone != nil {
		guardianPhone = strings.TrimSpace(*i.Guardian.Phone)
	}
	var guardianNotes string
	if i.Guardian.Notes != nil {
		guardianNotes = strings.TrimSpace(*i.Guardian.Notes)
	}

	if len(i.Guardian.Name) < 1 ||
		len(i.Guardian.Name) > 32 ||
		len(i.Guardian.LastName) < 1 ||
		len(i.Guardian.LastName) > 32 ||
		len(i.Guardian.Email) < 5 ||
		len(i.Guardian.Email) > 254 ||
		(i.Guardian.Password != nil && len(*i.Guardian.Password) < 8) {
		return ErrUnprocessableEntity
	}

	var passHash *string
	if i.Guardian.Password != nil {
		hash, err := argon2id.CreateHash(*i.Guardian.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			return err
		}
		passHash = &hash
	}

	if _, err := s.DB.ExecContext(ctx, `
		UPDATE portal_users
		SET
		    name = $1,
		    last_name = $2,
		    email = $3,
		    phone = $4,
		    notes = $5,
		    date_of_birth = $6,
		    account_enabled = $7,
		    account_active = $8,
		    password_hash = COALESCE($9, password_hash)
		WHERE id = $10
		AND school_id = $11
	`,
		i.Guardian.Name,
		i.Guardian.LastName,
		i.Guardian.Email,
		guardianPhone,
		guardianNotes,
		i.Guardian.DateOfBirth,
		i.Guardian.AccountEnabled,
		i.Guardian.AccountActive,
		passHash,
		i.GuardianID,
		schoolID,
	); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
