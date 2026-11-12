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

type UpdateGradeInput struct {
	InvalidPayload bool
	UserID         int64
	GradeID        int64
	Grade          UpdateGrade
}

type UpdateGrade struct {
	Name  string
	Level int
}

type UpdateGradeResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) UpdateGrade(ctx context.Context, i *UpdateGradeInput) (*UpdateGradeResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	gradeID := i.GradeID
	p := i.Grade
	// Resolve school
	var schoolID int64
	if err := s.DB.QueryRowContext(ctx, `SELECT y.school_id FROM academic_years y JOIN grades g ON g.academic_year_id = y.id WHERE g.id = $1`, gradeID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check permissions
	if !permissions.Can(permissions.PermissionGradeUpdate, userID, schoolID, ctx, s.DB) {
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

	// Parse name level var
	p.Name = strings.Replace(p.Name, "{level}", strconv.Itoa(p.Level), 1)

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Get the grade's current data
	var yearID int64
	var oldName string
	var oldLevel int
	if err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id, name, level
		FROM grades
		WHERE id = $1
	`, gradeID).Scan(&yearID, &oldName, &oldLevel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check for name conflict
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM grades WHERE name = $1 AND academic_year_id = $2 AND id <> $3
		)
	`, p.Name, yearID, gradeID).Scan(&exists); err != nil {
		log.Println(err)
		return nil, err
	}

	if exists {
		return nil, ErrConflict
	}

	// Update grade
	res, err := tx.ExecContext(ctx, `
		UPDATE grades
		SET name = $1, level = $2
		WHERE id = $3
	`, p.Name, p.Level, gradeID)
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
	details := "{user} updated the grade name from '" + oldName + "' to '" + p.Name + "' and level from " + strconv.Itoa(oldLevel) + " to " + strconv.Itoa(p.Level) + "."
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionGradeEdit, schoollogs.TypeEdit, "Grade updated", "{user} updated the grade '"+p.Name+"'.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateGradeResult{NoRowsAffected: noRowsAffected}, nil
}
