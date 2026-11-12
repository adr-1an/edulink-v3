package assignments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	schools2 "app/internal/helpers/staff/schools"
)

type DeleteAssignmentInput struct {
	UserID       int64
	AssignmentID int64
}

func (svc *Service) DeleteAssignment(ctx context.Context, i *DeleteAssignmentInput) error {
	userID := i.UserID
	assignmentID := i.AssignmentID

	schoolID, err := assignmentToSchoolID(assignmentID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentDelete, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var assignmentName string
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT title FROM assignments WHERE id = $1
	`, assignmentID).Scan(&assignmentName); err != nil {
		log.Println(err)
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM assignments
		WHERE id = $1
	`, assignmentID); err != nil {
		log.Println(err)
		return err
	}

	details := fmt.Sprintf("{user} deleted the assignment '%s'.", assignmentName)
	if err := schools2.StoreSchoolLog(schoolID, userID, schools2.ActionAssignmentDelete, schools2.TypeDelete, "Assignment deleted", "{user} deleted an assignment.", tx, ctx, svc.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
