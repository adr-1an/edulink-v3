package students

import (
	"context"
	"log"

	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
)

type ListGuardiansInput struct {
	UserID    int64
	StudentID int64
}

type Guardian struct {
	ID          int64
	Name        string
	LastName    string
	DateOfBirth string
	Email       string
	Phone       *string
}

type ListGuardiansResult struct {
	Access    schools.AccessData
	SchoolID  int64
	Guardians []Guardian
}

func (s *Service) ListGuardians(ctx context.Context, i *ListGuardiansInput) (*ListGuardiansResult, error) {
	sSchoolID, err := staffhelpers.StudentToSchoolID(ctx, s.DB, i.StudentID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !schools.Can(schools.PermissionStudentGuardianList, i.UserID, sSchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		    g.id,
		    g.name,
		    g.last_name,
		    g.date_of_birth,
		    g.email,
		    g.phone
		FROM portal_users g
		JOIN guardian_students gs
		ON gs.guardian_id = g.id
		WHERE gs.student_id = $1
		AND g.account_type = $2
	`, i.StudentID, helpers.AccTypeGuardian)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	guardians := make([]Guardian, 0)
	for rows.Next() {
		var g Guardian

		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.LastName,
			&g.DateOfBirth,
			&g.Email,
			&g.Phone,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		guardians = append(guardians, g)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, s.DB, i.UserID, sSchoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListGuardiansResult{
		Access:    access,
		SchoolID:  sSchoolID,
		Guardians: guardians,
	}, nil
}
