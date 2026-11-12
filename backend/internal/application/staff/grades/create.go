package grades

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type CreateGradeInput struct {
	InvalidPayload bool
	UserID         int64
	YearID         int64
	Grade          CreateGrade
}

type CreateGrade struct {
	Name  string
	Level int
}

type CreateGradeResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) CreateGrade(ctx context.Context, i *CreateGradeInput) (*CreateGradeResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	yearID := i.YearID
	p := i.Grade
	// Resolve school
	var schoolID int64
	if err := s.DB.QueryRowContext(ctx, `SELECT school_id FROM academic_years WHERE id = $1`, yearID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check permissions
	if !permissions.Can(permissions.PermissionGradeCreate, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse & validate
	p.Name = strings.TrimSpace(p.Name)

	if !strings.Contains(p.Name, "{level}") {
		return nil, ErrMissingLevelVar
	}

	if p.Name == "" || len(p.Name) < 7 || len(p.Name) > 32 {
		return nil, ErrInvalidName
	}

	if p.Level < 0 || p.Level > 20 {
		return nil, ErrLevelOutOfRange
	}

	// Parse the {level} variable
	p.Name = strings.Replace(p.Name, "{level}", strconv.Itoa(p.Level), 1)

	// Check for grade level & name conflict
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM grades WHERE academic_year_id = $1 AND name = $2 AND level = $3
		)
	`, yearID, p.Name, p.Level).Scan(&exists); err != nil {
		log.Println(err)
		return nil, err
	}

	if exists {
		return nil, ErrConflict
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Store grade
	res, err := tx.ExecContext(ctx, `
		INSERT INTO grades (id, academic_year_id, level, name)
		VALUES ($1, $2, $3, $4)
	`, id, yearID, p.Level, p.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	// Log action
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionGradeCreate, schoollogs.TypeCreate, "Grade created", "{user} created the grade '"+p.Name+"'.", tx, ctx, s.Sf, "{user} created grade '"+p.Name+"' at level "+strconv.Itoa(p.Level)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &CreateGradeResult{NoRowsAffected: noRowsAffected}, nil
}
