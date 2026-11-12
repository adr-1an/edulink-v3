package students

import (
	"context"
	"fmt"
	"log"

	"app/internal/helpers"
	staffhelpers "app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"

	"github.com/alexedwards/argon2id"
)

type CreateStudentInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	Student        StudentPayload
}

func (s *Service) CreateStudent(ctx context.Context, i *CreateStudentInput) error {
	userID := i.UserID
	schoolID := i.SchoolID
	p := i.Student

	if !schools.Can(schools.PermissionStudentCreate, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	p, err := validate(p)
	if err != nil {
		return ErrInvalidData
	}
	if p.AccountActive && !schools.Can(schools.PermissionPortalUserActivate, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	conflict, err := checkSchoolAndEmailConflict(tx, ctx, p.Email)
	if err != nil {
		log.Println(err)
		return err
	}

	if conflict {
		return ErrConflict
	}

	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return err
	}

	if (p.AccountEnabled && p.Password == "") ||
		(p.Password != "" && len(p.Password) < 8) {
		return ErrInvalidData
	}

	var passHash *string
	if p.Password != "" {
		hash, err := argon2id.CreateHash(p.Password, argon2id.DefaultParams)
		if err != nil {
			log.Println(err)
			return err
		}

		passHash = &hash
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
		schoolID,
		p.Name,
		p.LastName,
		p.DoB,
		p.Email,
		p.Phone,
		p.Notes,
		p.AccountEnabled,
		p.AccountActive,
		passHash,
		helpers.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	details := fmt.Sprintf("{user} created a new student: %s <%s>", p.Name, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentCreate, schools.TypeCreate, "Student created", "{user} created a student.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	if err := staffhelpers.RowsAffected(res); err != nil {
		return err
	}

	return nil
}
