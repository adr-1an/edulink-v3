package coursestudents

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type ListCourseStudentsStudent struct {
	ID       string
	Name     string
	LastName string
	Email    string
	Assigned bool
}

type ListCourseStudentsInput struct {
	UserID   int64
	CourseID int64
}

type ListCourseStudentsResult struct {
	Access   schools.AccessData
	Students []ListCourseStudentsStudent
}

func (svc *Service) ListCourseStudents(ctx context.Context, i *ListCourseStudentsInput) (*ListCourseStudentsResult, error) {
	userID := i.UserID
	courseID := i.CourseID

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	schoolID, err := getCourseSchoolID(courseID, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !schools.Can(schools.PermissionCourseStudentList, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	var students []ListCourseStudentsStudent

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    s.id,
		    s.name,
		    s.last_name,
		    s.email,
		    acs.portal_user_id IS NOT NULL AS assigned
		FROM portal_users s
		LEFT JOIN assigned_course_students acs
			ON acs.portal_user_id = s.id
			AND acs.course_id = $1
		WHERE s.school_id = $2
		AND s.account_type = $3
		ORDER BY s.last_name, s.name
	`, courseID, schoolID, helpers.AccTypeStudent)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var s ListCourseStudentsStudent

		if err := rows.Scan(&s.ID, &s.Name, &s.LastName, &s.Email, &s.Assigned); err != nil {
			log.Println(err)
			return nil, err
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, svc.DB, userID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListCourseStudentsResult{Access: access, Students: students}, nil
}
