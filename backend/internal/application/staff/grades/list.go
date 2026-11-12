package grades

import (
	"context"
	"log"

	"app/internal/helpers/staff/schools"
)

type ListGrade struct {
	ID    string
	Name  string
	Level int
}
type ListGradesResult struct {
	Access   schools.AccessData
	Grades   []ListGrade
	SchoolID int64
}
type ListGradesInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) ListGrades(ctx context.Context, i *ListGradesInput) (*ListGradesResult, error) {
	userID := i.UserID
	schoolID := i.SchoolID
	// Check permissions
	if !schools.Can(schools.PermissionGradeList, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	var grades []ListGrade

	// Fetch from DB
	rows, err := s.DB.QueryContext(ctx, `
		SELECT g.id, g.name, g.level
		FROM grades g
		JOIN academic_years y
		ON y.id = g.academic_year_id
		WHERE y.school_id = $1
	`, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var g ListGrade

		if err := rows.Scan(&g.ID, &g.Name, &g.Level); err != nil {
			log.Println(err)
			return nil, err
		}

		grades = append(grades, g)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, s.DB, userID, schoolID)
	if err != nil {
		return nil, err
	}

	return &ListGradesResult{Access: access, Grades: grades, SchoolID: schoolID}, nil
}
