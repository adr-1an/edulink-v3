package assignments

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type Assignment struct {
	RefPostID          *string
	Title              string
	Description        string
	DueDate            *time.Time
	SubmissionsEnabled bool
	SubmissionsCloseAt *time.Time
	Archived           bool
}

func parseAssignmentPayload(p Assignment) Assignment {
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)

	return p
}

func validateAssignmentPayload(p Assignment, editing bool) error {
	if len(p.Title) < 3 || len(p.Title) > 64 {
		return errors.New("invalid title")
	}

	if len(p.Description) > 4096 {
		return errors.New("invalid description")
	}

	if !editing {
		if p.DueDate != nil {
			if p.DueDate.Before(time.Now()) {
				return errors.New("invalid due date")
			}
		}

		if p.SubmissionsCloseAt != nil {
			if p.SubmissionsCloseAt.Before(time.Now()) {
				return errors.New("invalid submissions close date")
			}

			if p.DueDate != nil {
				if p.SubmissionsCloseAt.Before(*p.DueDate) {
					return errors.New("invalid submissions close date")
				}
			}
		}
	}

	return nil
}

func courseToSchoolID(courseID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		return 0, err
	}

	return schoolID, nil
}

func assignmentToSchoolID(assignmentID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignments a
		JOIN courses c
		ON a.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE a.id = $1
	`, assignmentID).Scan(&schoolID); err != nil {
		return 0, err
	}

	return schoolID, nil
}

func assignmentToCourseID(assignmentID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var courseID int64
	if err := db.QueryRowContext(ctx, `
		SELECT course_id FROM assignments WHERE id = $1
	`, assignmentID).Scan(&courseID); err != nil {
		return 0, err
	}

	return courseID, nil
}
