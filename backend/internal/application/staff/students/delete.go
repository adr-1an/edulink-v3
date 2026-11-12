package students

import (
	"context"
	"fmt"
	"log"

	"app/internal/permissions"
	"app/internal/portalusers"
	"app/internal/schoollogs"
)

type DeleteStudentInput struct {
	UserID    int64
	StudentID int64
}

func (s *Service) DeleteStudent(ctx context.Context, i *DeleteStudentInput) error {
	userID := i.UserID
	studentID := i.StudentID

	schoolID, err := portalusers.StudentToSchoolID(ctx, s.DB, studentID)
	if err != nil {
		log.Println(err)
		return err
	}

	if !permissions.Can(permissions.PermissionStudentDelete, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		name  string
		email string
	)
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM portal_users
		WHERE id = $1
		AND school_id = $2
		AND account_type = $3
		RETURNING name, email
	`, studentID, schoolID, portalusers.AccTypeStudent).Scan(&name, &email); err != nil {
		log.Println(err)
		return err
	}

	details := fmt.Sprintf("{user} deleted the student %s <%s>.", name, email)
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionStudentDelete, schoollogs.TypeDelete, "Student deleted", "{user} deleted a student.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
