package assignments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type CreateAssignmentInput struct {
	InvalidPayload bool
	UserID         int64
	CourseID       int64
	Data           Assignment
}

func (svc *Service) CreateAssignment(ctx context.Context, i *CreateAssignmentInput) error {
	userID := i.UserID
	courseID := i.CourseID
	p := i.Data

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	schoolID, err := courseToSchoolID(courseID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !permissions.Can(permissions.PermissionCourseAssignmentCreate, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	p = parseAssignmentPayload(p)
	if err := validateAssignmentPayload(p, false); err != nil {
		return ErrUnprocessableEntity
	}

	id, err := svc.Sf.NextID()
	if err != nil {
		log.Println(err)
		return err
	}

	var refPostSchoolID int64
	if p.RefPostID != nil {
		if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM course_posts p
		JOIN courses c
		ON p.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE p.id = $1
	`, p.RefPostID).Scan(&refPostSchoolID); err != nil {
			log.Println(err)
			return err
		}

		if refPostSchoolID != schoolID {
			return ErrForbidden
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO assignments (
		                         id,
		                         course_id,
		                         referenced_post_id,
		                         title,
		                         description,
		                         due_date,
		                         submissions_enabled,
		                         submissions_close_at
		                         )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		id,
		courseID,
		p.RefPostID,
		p.Title,
		p.Description,
		p.DueDate,
		p.SubmissionsEnabled,
		p.SubmissionsCloseAt,
	); err != nil {
		log.Println(err)
		return err
	}

	var submissions string
	if p.SubmissionsEnabled {
		submissions = "enabled"
	} else {
		submissions = "disabled"
	}

	details := fmt.Sprintf("{user} created the assignment '%s' with submissions %s.", p.Title, submissions)
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionAssignmentCreate, schoollogs.TypeCreate, "Assignment created", "{user} created an assignment", tx, ctx, svc.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
