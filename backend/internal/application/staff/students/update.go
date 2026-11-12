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

type UpdateStudentInput struct {
	InvalidPayload bool
	UserID         int64
	StudentID      int64
	Student        StudentPayload
}

func (s *Service) UpdateStudent(ctx context.Context, i *UpdateStudentInput) error {
	userID := i.UserID
	studentID := i.StudentID
	p := i.Student

	schoolID, err := staffhelpers.StudentToSchoolID(ctx, s.DB, studentID)
	if err != nil {
		log.Println(err)
		return err
	}

	if !schools.Can(schools.PermissionStudentUpdate, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	p, err = validate(p)
	if err != nil {
		return ErrInvalidData
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

	var accountActive bool
	if err := tx.QueryRowContext(ctx, `
		SELECT account_active
		FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
	`, studentID, schoolID, helpers.AccTypeStudent).Scan(&accountActive); err != nil {
		log.Println(err)
		return err
	}
	if permission, changed := schools.PortalUserActiveStatePermission(accountActive, p.AccountActive); changed &&
		!schools.Can(permission, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if p.Password != "" && len(p.Password) < 8 {
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

	var (
		name     string
		lastName string
		email    string
	)
	if err := s.DB.QueryRowContext(ctx, `
		SELECT name, last_name, email
		FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
	`, studentID, schoolID, helpers.AccTypeStudent).Scan(&name, &lastName, &email); err != nil {
		log.Println(err)
		return err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE portal_users
		SET
		    name = $1,
		    last_name = $2,
		    date_of_birth = $3,
		    email = $4,
		    phone = $5,
		    notes = $6,
		    account_enabled = $7,
		    account_active = $8,
		    password_hash = COALESCE($9, password_hash)
		WHERE id = $10
		AND school_id = $11
		AND account_type = $12
	`,
		p.Name,
		p.LastName,
		p.DoB,
		p.Email,
		p.Phone,
		p.Notes,
		p.AccountEnabled,
		p.AccountActive,
		passHash,
		studentID,
		schoolID,
		helpers.AccTypeStudent,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	if passHash != nil {
		// Password has been changed, delete all sessions
		if _, err := s.DB.ExecContext(ctx, `
			DELETE FROM portal_sessions
			WHERE portal_user_id = $1
		`, studentID); err != nil {
			log.Println(err)
			return err
		}
	}

	details := fmt.Sprintf(
		"{user} updated the student %s %s <%s> to %s %s <%s>.",
		name, lastName, email, p.Name, p.LastName, p.Email)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStudentEdit, schools.TypeEdit, "Student updated", "{user} updated a student.", tx, ctx, s.Sf, details); err != nil {
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
