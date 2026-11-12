package grades

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type DeleteGradeInput struct {
	UserID  int64
	GradeID int64
}

type DeleteGradeResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) DeleteGrade(ctx context.Context, i *DeleteGradeInput) (*DeleteGradeResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	gradeID := i.GradeID
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
	if !permissions.Can(permissions.PermissionGradeDelete, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var gradeName string
	var gradeLevel int
	if err := tx.QueryRowContext(ctx, `
		SELECT name, level FROM grades WHERE id = $1
	`, gradeID).Scan(&gradeName, &gradeLevel); err != nil {
		log.Println(err)
		return nil, err
	}

	// Delete grade
	res, err := tx.ExecContext(ctx, `DELETE FROM grades WHERE id = $1`, gradeID)
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
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionGradeDelete, schoollogs.TypeDelete, "Grade deleted", "{user} deleted the grade '"+gradeName+"'.", tx, ctx, s.Sf, "{user} deleted grade '"+gradeName+"' at level "+strconv.Itoa(gradeLevel)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteGradeResult{NoRowsAffected: noRowsAffected}, nil
}
