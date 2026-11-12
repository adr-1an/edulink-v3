package courses

import (
	"context"
	"log"

	"app/internal/helpers/staff/schools"
)

type ListCoursesCourse struct {
	ID          string
	Name        string
	Description string
	Color       string
}

type ListCoursesInput struct {
	UserID  int64
	GradeID int64
}

type ListCoursesResult struct {
	Access  schools.AccessData
	Courses []ListCoursesCourse
}

func (svc *Service) ListCourses(ctx context.Context, i *ListCoursesInput) (*ListCoursesResult, error) {
	userID := i.UserID
	gradeID := i.GradeID

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM grades g
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		log.Println(err)
		return nil, err
	}

	if !schools.Can(schools.PermissionCourseList, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var courses []ListCoursesCourse

	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, description, color
		FROM courses c
		WHERE grade_id = $1
	`, gradeID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var c ListCoursesCourse

		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color); err != nil {
			log.Println(err)
			return nil, err
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	// Get user access
	access, err := schools.GetAllUserPermissions(ctx, svc.DB, userID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListCoursesResult{Access: access, Courses: courses}, nil
}
