package courses

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type UpdateCourseData struct {
	Name        string
	Description string
	Color       string
}

type UpdateCourseInput struct {
	InvalidPayload bool
	UserID         int64
	CourseID       int64
	Data           UpdateCourseData
}

type UpdateCourseResult struct{ AffectedRows int64 }

func (svc *Service) UpdateCourse(ctx context.Context, i *UpdateCourseInput) (*UpdateCourseResult, error) {
	userID := i.UserID
	courseID := i.CourseID
	p := i.Data

	// Get grade and school ID
	gradeID, schoolID, err := getCourseGradeAndSchoolID(svc.DB, ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	// Check permissions
	if !permissions.Can(permissions.PermissionCourseUpdate, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	// Parse
	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	p.Name, p.Description, p.Color = parseCoursePayload(p.Name, p.Description, p.Color)

	// Validate
	if !validateCoursePayload(p.Name, p.Description, p.Color) {
		return nil, ErrUnprocessableEntity
	}

	// Start tx
	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var oldName string
	var oldDescription string
	var oldColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, description, color
		FROM courses
		WHERE id = $1
	`, courseID).Scan(&oldName, &oldDescription, &oldColor); err != nil {
		log.Println(err)
		return nil, err
	}

	// Check for course name conflicts
	conflict, err := checkNameConflictExcept(p.Name, gradeID, courseID, tx, ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if conflict {
		return nil, ErrConflict
	}

	// Update course
	res, err := tx.ExecContext(ctx, `
		UPDATE courses c
		SET name = $1, description = $2, color = $3
		WHERE id = $4
		AND grade_id = $5
	`, p.Name, p.Description, p.Color, courseID, gradeID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	details := "{user} updated the course name from '" + oldName + "' to '" + p.Name + "', description from '" + oldDescription + "' to '" + p.Description + "', and color from '#" + oldColor + "' to '#" + p.Color + "'."
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionCourseEdit, schoollogs.TypeEdit, "Course updated", "{user} updated the course '"+p.Name+"'.", tx, ctx, svc.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateCourseResult{AffectedRows: affectedRows}, nil
}
