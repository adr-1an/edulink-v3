package academicyears

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/helpers/staff/schools"
)

type ClearAcademicYearInput struct {
	UserID   int64
	SchoolID int64
}

type ClearAcademicYearResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) ClearAcademicYear(ctx context.Context, i *ClearAcademicYearInput) (*ClearAcademicYearResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	schoolID := i.SchoolID
	// Check permissions
	if !schools.Can(schools.PermissionAcademicYearToggleActive, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var activeYearID int64
	var activeStartYear int
	var activeEndYear int
	if err := tx.QueryRowContext(ctx, `
		SELECT id, start_year, end_year
		FROM academic_years
		WHERE school_id = $1
		AND is_active = TRUE
	`, schoolID).Scan(&activeYearID, &activeStartYear, &activeEndYear); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	// Deactivate current year
	res, err := tx.ExecContext(ctx, `
		UPDATE academic_years
		SET is_active = FALSE
		WHERE school_id = $1
		AND is_active = TRUE
	`, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	details := "{user} cleared the active academic year."
	if activeYearID != 0 {
		details = "{user} cleared active academic year " + strconv.Itoa(activeStartYear) + "-" + strconv.Itoa(activeEndYear) + " with ID " + strconv.FormatInt(activeYearID, 10) + "."
	}
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolEdit, schools.TypeEdit, "Active academic year cleared", "{user} cleared the active academic year.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	return &ClearAcademicYearResult{NoRowsAffected: noRowsAffected}, nil
}
