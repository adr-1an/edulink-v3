package guardians

import (
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"time"
)

type Student struct {
	ID       int64
	Name     string
	LastName string
	DoB      time.Time
}

type ListGuardianStudentsInput struct {
	UserID     int64
	GuardianID int64
}

type ListGuardianStudentsResult struct {
	Students []Student
	SchoolID int64
}

func (s *Service) ListGuardianStudents(
	ctx context.Context,
	i *ListGuardianStudentsInput,
) (
	*ListGuardianStudentsResult,
	error,
) {
	schoolID, err := staffhelpers.GuardianToSchoolID(ctx, s.DB, i.GuardianID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !schools.Can(schools.PermissionGuardianStudentList, i.UserID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT s.id, s.name, s.last_name, s.date_of_birth
		FROM portal_users s
		JOIN guardian_students gs ON gs.student_id = s.id
		WHERE gs.guardian_id = $1
		AND s.school_id = $2
		AND s.account_type = $3
	`, i.GuardianID, schoolID, helpers.AccTypeStudent)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	students := make([]Student, 0)
	for rows.Next() {
		var s Student

		if err := rows.Scan(&s.ID, &s.Name, &s.LastName, &s.DoB); err != nil {
			log.Println(err)
			return nil, err
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListGuardianStudentsResult{
		Students: students,
		SchoolID: schoolID,
	}, nil
}
