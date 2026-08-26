package schools

import (
	"app/internal/helpers/staff/schools"
	"context"
	"log"
)

type SchoolViewInput struct {
	UserID   int64
	SchoolID int64
}

type SchoolViewResult struct {
	School schoolView
}

type gradeView struct {
	ID             string `json:"id"`
	AcademicYearID string `json:"academicYearId"`
	Level          int    `json:"level"`
	Name           string `json:"name"`
	CreatedAt      string `json:"createdAt"`
}

type schoolView struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	RegionCode string      `json:"regionCode"`
	Grades     []gradeView `json:"grades"`
	CreatedAt  string      `json:"createdAt"`
	UpdatedAt  string      `json:"updatedAt"`
}

func (s *Service) ViewSchool(ctx context.Context, i *SchoolViewInput) (*SchoolViewResult, error) {
	// Check permissions
	if !schools.Can(schools.PermissionSchoolView, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrNoPermission
	}

	var sc schoolView
	var gs []gradeView

	// Get school info
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
		    id, name, region_code, created_at, updated_at
		FROM schools
		WHERE id = $1
		AND deleted_at IS NULL
	`, i.SchoolID).Scan(
		&sc.ID,
		&sc.Name,
		&sc.RegionCode,
		&sc.CreatedAt,
		&sc.UpdatedAt,
	); err != nil {
		log.Println(err)
		return nil, err
	}

	// Get grades
	rows, err := s.DB.QueryContext(ctx, `
		SELECT g.id, g.academic_year_id, g.level, g.name, g.created_at
		FROM grades g
		JOIN academic_years y
		ON g.academic_year_id = y.id
		JOIN schools s
		ON y.school_id = s.id
		WHERE s.id = $1
		AND y.is_active = true
	`, i.SchoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var g gradeView

		if err := rows.Scan(&g.ID, &g.AcademicYearID, &g.Level, &g.Name, &g.CreatedAt); err != nil {
			log.Println(err)
			return nil, err
		}

		gs = append(gs, g)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	// Return data
	sc.Grades = gs

	return &SchoolViewResult{
		School: sc,
	}, nil
}
