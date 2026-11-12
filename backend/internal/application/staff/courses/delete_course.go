package courses

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/helpers/staff/schools"
)

type DeleteCourseInput struct {
	UserID   int64
	CourseID int64
}

type DeleteCourseResult struct{ AffectedRows int64 }

func (svc *Service) DeleteCourse(ctx context.Context, i *DeleteCourseInput) (*DeleteCourseResult, error) {
	userID := i.UserID
	courseID := i.CourseID

	_, schoolID, err := getCourseGradeAndSchoolID(svc.DB, ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !schools.Can(schools.PermissionCourseDelete, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var courseName string
	if err := tx.QueryRowContext(ctx, `
		SELECT name FROM courses WHERE id = $1
	`, courseID).Scan(&courseName); err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM courses c
		USING grades g, academic_years y
		WHERE c.id = $1
		AND c.grade_id = g.id
		AND g.academic_year_id = y.id
		AND y.school_id = $2
	`, courseID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionCourseDelete, schools.TypeDelete, "Course deleted", "{user} deleted the course '"+courseName+"'.", tx, ctx, svc.Sf, "{user} deleted course '"+courseName+"' with ID "+strconv.FormatInt(courseID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteCourseResult{AffectedRows: affectedRows}, nil
}
