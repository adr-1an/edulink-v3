package submissionscores

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/application/staff/submissions"
	"app/internal/helpers/staff/schools"
)

type ClearSubmissionScoreInput struct {
	UserID       int64
	SubmissionID int64
}

func (svc *Service) ClearSubmissionScore(ctx context.Context, i *ClearSubmissionScoreInput) error {
	userID := i.UserID
	subID := i.SubmissionID

	schoolID, err := submissions.SubToSchoolID(subID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !schools.Can(schools.PermissionSubmissionRemoveGrade, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	if _, err := svc.DB.ExecContext(ctx, `
		DELETE FROM submission_scores WHERE submission_id = $1
	`, subID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
