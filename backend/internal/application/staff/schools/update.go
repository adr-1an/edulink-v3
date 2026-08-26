package schools

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

type SchoolUpdate struct {
	Name                 string
	RegionCode           string
	ActiveAcademicYearID *int64
}

type UpdateSchoolInput struct {
	UserID   int64
	SchoolID int64
	School   SchoolUpdate
}

func (s *Service) UpdateSchool(ctx context.Context, i *UpdateSchoolInput) error {
	// School permission check
	if !schools.Can(schools.PermissionSchoolUpdate, i.UserID, i.SchoolID, ctx, s.DB) {
		return ErrNoPermission
	}

	if i.School.ActiveAcademicYearID != nil &&
		// If p.ActiveAcademicYearID isn't null, it means the user is trying to change the currently active academic year.
		// To do that, they need the academicYear.toggle permission.
		!schools.Can(schools.PermissionAcademicYearToggleActive, i.UserID, i.SchoolID, ctx, s.DB) {
		return ErrNoPermission
	}

	// Parse & validate
	i.School.Name = strings.TrimSpace(i.School.Name)
	i.School.RegionCode = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(i.School.RegionCode, " ", "")))

	if i.School.Name == "" || len(i.School.Name) > 64 {
		return ErrInvalidName
	}

	if i.School.RegionCode != "" {
		region, err := language.ParseRegion(i.School.RegionCode)
		if err != nil {
			return ErrBadRequest
		}

		if !region.IsCountry() {
			return ErrInvalidRegionCode
		}
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var oldName string
	var oldRegionCode string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, region_code
		FROM schools
		WHERE id = $1
	`, i.SchoolID).Scan(&oldName, &oldRegionCode); err != nil {
		log.Println(err)
		return err
	}

	activeYearDetails := ""
	if i.School.ActiveAcademicYearID != nil {
		// Get the Academic Year's school ID
		var yearSchoolID int64
		var newStartYear int
		var newEndYear int
		if err := tx.QueryRowContext(ctx, `
		SELECT school_id, start_year, end_year FROM academic_years WHERE id = $1
	`, i.School.ActiveAcademicYearID).Scan(&yearSchoolID, &newStartYear, &newEndYear); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrForbidden
			}
			log.Println(err)
			return err
		}

		if yearSchoolID != i.SchoolID {
			return ErrForbidden
		}

		activeYearDetails = " Active academic year was set to " + strconv.Itoa(newStartYear) + "-" + strconv.Itoa(newEndYear) + "."

		// Deactivate old year
		if _, err := tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = false
			WHERE school_id = $1
			AND is_active = true
		`, i.SchoolID); err != nil {
			log.Println(err)
			return err
		}

		// Activate new year
		res, err := tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = true
			WHERE id = $1
			AND is_active = false
		`, i.School.ActiveAcademicYearID)
		if err != nil {
			log.Println(err)
			return err
		}
		if err := staff_helpers.RowsAffected(res); err != nil {
			return err
		}
	}

	// Update
	res, err := tx.ExecContext(ctx, `
		UPDATE schools
		SET name = $1, region_code = $2, updated_at = NOW()
		WHERE id = $3
	`, i.School.Name, i.School.RegionCode, i.SchoolID)
	if err != nil {
		log.Println(err)
		return err
	}
	if err := staff_helpers.RowsAffected(res); err != nil {
		return err
	}

	// Log action
	details := "{user} updated the school name from '" + oldName + "' to '" + i.School.Name + "' and region from '" + oldRegionCode + "' to '" + i.School.RegionCode + "'." + activeYearDetails
	if err := schools.StoreSchoolLog(i.SchoolID, i.UserID, schools.ActionSchoolEdit, schools.TypeEdit, "School updated", "{user} updated the school '"+i.School.Name+"'.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
