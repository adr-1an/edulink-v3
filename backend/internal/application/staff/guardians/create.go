package guardians

import (
	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
)

type CreateGuardian struct {
	Name            string
	LastName        string
	Email           string
	Phone           *string
	Notes           *string
	DateOfBirth     time.Time
	AccountEnabled  bool
	ActivateAccount bool
	Password        *string
}

type CreateGuardianInput struct {
	UserID   int64
	SchoolID int64
	Guardian CreateGuardian
}

type CreateGuardianResult struct {
	InsertedID int64
}

func (s *Service) CreateGuardian(ctx context.Context, i *CreateGuardianInput) (*CreateGuardianResult, error) {
	if !schools.Can(schools.PermissionGuardianCreate, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}
	if i.Guardian.ActivateAccount && !schools.Can(schools.PermissionPortalUserActivate, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
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
		return nil, ErrUnprocessableEntity
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var guardianPass string
	if i.Guardian.Password != nil {
		guardianPass, err = argon2id.CreateHash(*i.Guardian.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO portal_users (
		                          id,
		                          school_id,
		                          name,
		                          last_name,
		                          date_of_birth,
		                          email,
		                          phone,
		                          notes,
		                          account_enabled,
		                          account_active,
		                          password_hash,
		                          account_type
		                          )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		id,
		i.SchoolID,
		i.Guardian.Name,
		i.Guardian.LastName,
		i.Guardian.DateOfBirth,
		i.Guardian.Email,
		guardianPhone,
		guardianNotes,
		i.Guardian.AccountEnabled,
		i.Guardian.ActivateAccount,
		guardianPass,
		helpers.AccTypeGuardian,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if affected == 0 {
		log.Println("no rows inserted")
		return nil, ErrNoRowsInserted
	}

	return &CreateGuardianResult{
		InsertedID: id,
	}, nil
}
