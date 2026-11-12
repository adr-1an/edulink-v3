package submissionscores

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"app/internal/application/staff/submissions"
	"app/internal/permissions"
)

type GradeSubmissionData struct {
	Score int
	Notes string
}

type GradeSubmissionInput struct {
	InvalidPayload bool
	UserID         int64
	SubmissionID   int64
	Data           GradeSubmissionData
}

func (svc *Service) GradeSubmission(ctx context.Context, i *GradeSubmissionInput) error {
	userID := i.UserID
	subID := i.SubmissionID
	p := i.Data

	schoolID, err := submissions.SubToSchoolID(subID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !permissions.Can(permissions.PermissionSubmissionGrade, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	p.Notes = strings.TrimSpace(p.Notes)
	if p.Score < 0 || p.Score > 100 || len(p.Notes) > 2048 {
		return ErrUnprocessableEntity
	}

	var exists bool
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM assignment_submissions WHERE id = $1 AND (status = 'pending' OR status = 'returned')
		)
	`, subID).Scan(&exists); err != nil {
		log.Println(err)
		return err
	}

	if exists {
		return ErrConflict
	}

	if _, err := svc.DB.ExecContext(ctx, `
		INSERT INTO submission_scores (submission_id, graded_by, score_percentage, notes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (submission_id) DO UPDATE
		SET graded_by = EXCLUDED.graded_by,
		    score_percentage = EXCLUDED.score_percentage,
		    notes = EXCLUDED.notes,
		    graded_at = NOW()
	`, subID, userID, p.Score, p.Notes); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
