package submissions

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type ReturnSubmissionInput struct {
	UserID       int64
	SubmissionID int64
}

func (svc *Service) ReturnSubmission(ctx context.Context, i *ReturnSubmissionInput) error {
	userID := i.UserID
	subID := i.SubmissionID

	schoolID, err := SubToSchoolID(subID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !schools.Can(schools.PermissionSubmissionReturn, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	if _, err := svc.DB.ExecContext(ctx, `
		DELETE FROM submission_scores WHERE submission_id = $1
	`, subID); err != nil {
		log.Println(err)
		return err
	}

	result, err := svc.DB.ExecContext(ctx, `
		UPDATE assignment_submissions
		SET status = $1
		WHERE id = $2
		AND status = $3
	`, helpers.SubmissionStatusReturned, subID, helpers.SubmissionStatusSubmitted)
	if err != nil {
		log.Println(err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	if rowsAffected != 1 {
		return ErrConflict
	}

	return nil
}
