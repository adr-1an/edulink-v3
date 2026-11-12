package courses

import (
	"context"
	"log"
	"strings"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type CreateCourseData struct {
	Name        string
	Description string
	Color       string
}

type CreateCourseInput struct {
	InvalidPayload bool
	UserID         int64
	GradeID        int64
	Data           CreateCourseData
}

type CreateCourseResult struct{ AffectedRows int64 }

func (svc *Service) CreateCourse(ctx context.Context, i *CreateCourseInput) (*CreateCourseResult, error) {
	userID := i.UserID
	gradeID := i.GradeID
	p := i.Data

	// Get the grade's school ID
	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM academic_years y
		JOIN grades g
		ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		log.Println(err)
		return nil, err
	}

	// Check permissions
	if !permissions.Can(permissions.PermissionCourseCreate, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	// Parse
	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.Color = strings.TrimSpace(strings.TrimPrefix(p.Color, "#"))

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

	// Check for course name conflicts
	conflict, err := checkNameConflict(p.Name, gradeID, tx, ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if conflict {
		return nil, ErrConflict
	}

	// Generate ID
	id, err := svc.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Store course
	res, err := tx.ExecContext(ctx, `
		INSERT INTO courses (id, name, description, color, grade_id)
		VALUES ($1, $2, $3, $4, $5)
	`, id, p.Name, p.Description, p.Color, gradeID)
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
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionCourseCreate, schoollogs.TypeCreate, "Course created", "{user} created the course '"+p.Name+"'.", tx, ctx, svc.Sf, "{user} created course '"+p.Name+"' with color '#"+p.Color+"' and description '"+p.Description+"'."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &CreateCourseResult{AffectedRows: affectedRows}, nil
}
