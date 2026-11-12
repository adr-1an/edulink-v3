package assignments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	schools2 "app/internal/helpers/staff/schools"
)

type UpdateAssignmentInput struct {
	InvalidPayload bool
	UserID         int64
	AssignmentID   int64
	Data           Assignment
}

func (svc *Service) UpdateAssignment(ctx context.Context, i *UpdateAssignmentInput) error {
	userID := i.UserID
	assignmentID := i.AssignmentID
	p := i.Data

	schoolID, err := assignmentToSchoolID(assignmentID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentUpdate, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	p = parseAssignmentPayload(p)
	if err := validateAssignmentPayload(p, true); err != nil {
		return ErrUnprocessableEntity
	}

	if p.RefPostID != nil {
		courseID, err := assignmentToCourseID(assignmentID, svc.DB, ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrForbidden
			} else {
				log.Println(err)
				return err
			}
		}

		var postCourseID int64
		if err := svc.DB.QueryRowContext(ctx, `
			SELECT course_id FROM course_posts WHERE id = $1
		`, p.RefPostID).Scan(&postCourseID); err != nil {
			log.Println(err)
			return err
		}

		if postCourseID != courseID {
			return ErrForbidden
		}
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var archivedAt *time.Time
	if p.Archived {
		now := time.Now()
		archivedAt = &now
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE assignments
		SET
		    referenced_post_id = $1,
		    title = $2,
		    description = $3,
		    due_date = $4,
		    submissions_enabled = $5,
		    submissions_close_at = $6,
		    archived_at = $7
		WHERE id = $8
	`,
		p.RefPostID,
		p.Title,
		p.Description,
		p.DueDate,
		p.SubmissionsEnabled,
		p.SubmissionsCloseAt,
		archivedAt,
		assignmentID,
	); err != nil {
		log.Println(err)
		return err
	}

	details := fmt.Sprintf("{user} edited the assignment '%s'.", p.Title)
	if err := schools2.StoreSchoolLog(schoolID, userID, schools2.ActionAssignmentEdit, schools2.TypeEdit, "Assignment updated", "{user} edited an assignment.", tx, ctx, svc.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
