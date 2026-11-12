package academicyears

import (
	"context"
	"log"
	"strconv"

	"app/internal/helpers/staff/schools"
)

type CreateAcademicYear struct {
	From int
	To   int
}

type CreateAcademicYearInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	AcademicYear   CreateAcademicYear
}

type CreateAcademicYearResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) CreateAcademicYear(ctx context.Context, i *CreateAcademicYearInput) (*CreateAcademicYearResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	if !schools.Can(schools.PermissionAcademicYearCreate, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse & validate

	// If you're here in/after the year 10,000, add an extra 9 to the 9999 below
	// Or if you've gone back to before 1900, IDK anymore maybe change it to like 100 or smth
	// Also I had to write "IDK" in all caps or GoLand would cry
	// I know it's corny
	if i.AcademicYear.From < 1900 || i.AcademicYear.From > 9999 {
		return nil, ErrUnprocessableEntity
	}

	if i.AcademicYear.To < i.AcademicYear.From {
		return nil, ErrUnprocessableEntity
	}

	// Check for conflicts
	var exists bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM academic_years
		 	WHERE school_id = $1
		 	AND start_year = $2
		 	AND end_year = $3
		)
	`, i.SchoolID, i.AcademicYear.From, i.AcademicYear.To).Scan(&exists); err != nil {
		log.Println(err)
		return nil, err
	}

	if exists {
		return nil, ErrAcademicYearConflict
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Insert academic year
	res, err := tx.ExecContext(ctx, `
		INSERT INTO academic_years (id, school_id, start_year, end_year)
		VALUES ($1, $2, $3, $4)
	`, id, i.SchoolID, i.AcademicYear.From, i.AcademicYear.To)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	yearLabel := strconv.Itoa(i.AcademicYear.From) + "-" + strconv.Itoa(i.AcademicYear.To)
	// Log action
	if err := schools.StoreSchoolLog(i.SchoolID, i.UserID, schools.ActionAcademicYearCreate, schools.TypeCreate, "Academic year created", "{user} created the academic year "+yearLabel+".", tx, ctx, s.Sf, "{user} created academic year "+yearLabel+" with ID "+strconv.FormatInt(id, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &CreateAcademicYearResult{NoRowsAffected: noRowsAffected}, nil
}
