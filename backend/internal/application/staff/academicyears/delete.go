package academicyears

import (
	"context"
	"log"
	"strconv"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type DeleteAcademicYearInput struct {
	UserID         int64
	AcademicYearID int64
}

type DeleteAcademicYearResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) DeleteAcademicYear(ctx context.Context, i *DeleteAcademicYearInput) (*DeleteAcademicYearResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	acYearID := i.AcademicYearID
	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get the year's data
	var yearSchoolID int64
	var startYear int
	var endYear int
	if err := tx.QueryRowContext(ctx, `
		SELECT school_id, start_year, end_year FROM academic_years WHERE id = $1
	`, acYearID).Scan(&yearSchoolID, &startYear, &endYear); err != nil {
		log.Println(err)
		return nil, err
	}

	// Check permissions
	if !permissions.Can(permissions.PermissionAcademicYearDelete, userID, yearSchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Check if the current year is active
	var (
		active    bool
		hasGrades bool
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT
		    EXISTS (
		    SELECT 1 FROM academic_years
			WHERE id = $1
			AND is_active = TRUE
		),
		    EXISTS (
		        SELECT 1 FROM grades
				WHERE academic_year_id = $1
		    )
	`, acYearID).Scan(&active, &hasGrades); err != nil {
		log.Println(err)
		return nil, err
	}

	if active {
		return nil, ErrYearIsActive
	}

	if hasGrades {
		return nil, ErrYearHasGrades
	}

	// Delete the year
	res, err := tx.ExecContext(ctx, `
		DELETE FROM academic_years
		WHERE id = $1
	  	AND school_id = $2
		AND is_active = FALSE
	`, acYearID, yearSchoolID)
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

	yearLabel := strconv.Itoa(startYear) + "-" + strconv.Itoa(endYear)
	if err := schoollogs.StoreSchoolLog(yearSchoolID, userID, schoollogs.ActionAcademicYearDelete, schoollogs.TypeDelete, "Academic year deleted", "{user} deleted the academic year "+yearLabel+".", tx, ctx, s.Sf, "{user} deleted academic year "+yearLabel+" with ID "+strconv.FormatInt(acYearID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteAcademicYearResult{NoRowsAffected: noRowsAffected}, nil
}
