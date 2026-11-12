package academicyears

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"app/internal/helpers/staff/schools"
)

type PromoteSchoolInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	Promotion      PromoteSchool
}

type PromoteSchool struct {
	NewAcademicYear struct {
		From int
		To   int
	}
	Opts struct {
		ActivateAfterPromotion bool
		TransferGrades         bool
		PromoteGradeLevels     bool
	}
}

type PromoteSchoolResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) PromoteSchool(ctx context.Context, i *PromoteSchoolInput) (*PromoteSchoolResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	schoolID := i.SchoolID
	p := i.Promotion
	// Check permissions
	if !schools.Can(schools.PermissionSchoolPromote, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// -- 1. Get currently active academic year --
	var activeYearID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM academic_years
		WHERE school_id = $1
			AND is_active = true
		LIMIT 1
	`, schoolID).Scan(&activeYearID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveYear
		} else {
			log.Println(err)
			return nil, err
		}

	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse & validate
	if p.NewAcademicYear.From < 1900 || p.NewAcademicYear.From > 9999 {
		return nil, ErrUnprocessableEntity
	}

	if p.NewAcademicYear.To < p.NewAcademicYear.From {
		return nil, ErrUnprocessableEntity
	}

	// Check for start & end year conflicts
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM academic_years WHERE start_year = $1 AND end_year = $2 AND school_id = $3
		)
	`, p.NewAcademicYear.From, p.NewAcademicYear.To, schoolID).Scan(&exists); err != nil {
		log.Println(err)
		return nil, err
	}

	if exists {
		return nil, ErrAcademicYearConflict
	}

	// -- 3. Create new school year --
	// Generate ID
	newAcademicYearID, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO academic_years (id, school_id, start_year, end_year)
		VALUES ($1, $2, $3, $4)
	`, newAcademicYearID, schoolID, p.NewAcademicYear.From, p.NewAcademicYear.To)
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

	// Activate new academic year if requested
	if p.Opts.ActivateAfterPromotion {
		// Deactivate old
		if _, err := tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = false
			WHERE school_id = $1
			AND is_active = true
		`, schoolID); err != nil {
			log.Println(err)
			return nil, err
		}

		// Activate new
		res, err = tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = true
			WHERE id = $1
			AND is_active = false
		`, newAcademicYearID)
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
	}

	// -- 4. Transfer old data (on request) --

	// - Transfer grades -
	if p.Opts.TransferGrades {
		// Select all grades from the old year
		type grade struct {
			id    int64
			level int
			name  string
		}
		var grades []grade
		rows, err := tx.QueryContext(ctx, `
		SELECT id, level, name FROM grades WHERE academic_year_id = $1
	`, activeYearID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer func() { _ = rows.Close() }()

		// Scan rows
		for rows.Next() {
			var g grade

			if err := rows.Scan(&g.id, &g.level, &g.name); err != nil {
				log.Println(err)
				return nil, err
			}

			grades = append(grades, g)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
			return nil, err
		}

		if len(grades) > 0 {
			// For each grade, generate a new ID
			for i := range grades {
				newID, err := s.Sf.NextID()
				if err != nil {
					log.Println(err)
					return nil, err
				}

				// Increase level by 1 if requested
				if p.Opts.PromoteGradeLevels && p.Opts.TransferGrades {
					grades[i].name = strings.ReplaceAll(grades[i].name, strconv.Itoa(grades[i].level), strconv.Itoa(grades[i].level+1))
					grades[i].level++
				}

				grades[i].id = newID
			}

			// Insert grades
			var (
				values []string
				args   []any
			)
			for i, g := range grades {
				base := i * 4

				values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)",
					base+1,
					base+2,
					base+3,
					base+4,
				))

				args = append(args, g.id, newAcademicYearID, g.level, g.name)
			}

			// The "(0,0,0,0)" is a workaround. Putting %s inside the query would trigger a false SQL parsing error.
			query := "INSERT INTO grades (id, academic_year_id, level, name) VALUES (0,0,0,0)"
			query = strings.ReplaceAll(query, "(0,0,0,0)", strings.Join(values, ","))
			res, err = tx.ExecContext(ctx, query, args...)
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
		}
	}

	// TODO: - Transfer courses once implemented -

	// Log action
	promotionDetails := "{user} promoted the school to academic year " + strconv.Itoa(p.NewAcademicYear.From) + "-" + strconv.Itoa(p.NewAcademicYear.To) + "."
	if p.Opts.ActivateAfterPromotion {
		promotionDetails += " The new academic year was activated."
	}
	if p.Opts.TransferGrades {
		promotionDetails += " Grades were transferred."
	}
	if p.Opts.PromoteGradeLevels {
		promotionDetails += " Grade levels were promoted."
	}
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolPromote, schools.TypeOther, "School promoted", "{user} promoted the school to academic year "+strconv.Itoa(p.NewAcademicYear.From)+"-"+strconv.Itoa(p.NewAcademicYear.To)+".", tx, ctx, s.Sf, promotionDetails); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &PromoteSchoolResult{NoRowsAffected: noRowsAffected}, nil
}
