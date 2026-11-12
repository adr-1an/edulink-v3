package coursestudents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type AddOrRemoveCourseStudentData struct {
	StudentID string
}

type AddOrRemoveCourseStudentInput struct {
	InvalidPayload bool
	UserID         int64
	CourseID       int64
	Data           AddOrRemoveCourseStudentData
	Assign         bool
}

func (svc *Service) AddOrRemoveCourseStudent(ctx context.Context, i *AddOrRemoveCourseStudentInput) error {
	userID := i.UserID
	courseID := i.CourseID
	p := i.Data
	assign := i.Assign

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	schoolID, err := getCourseSchoolID(courseID, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if assign {
		if !schools.Can(schools.PermissionCourseStudentAssign, userID, schoolID, ctx, svc.DB) {
			return ErrForbidden
		}
	} else {
		if !schools.Can(schools.PermissionCourseStudentRemove, userID, schoolID, ctx, svc.DB) {
			return ErrForbidden
		}
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	if p.StudentID == "" {
		return ErrUnprocessableEntity
	}

	var belongsToSchool bool
	var courseName string
	var studentName string
	if err := tx.QueryRowContext(ctx, `
		WITH exists AS (
		    SELECT EXISTS (
		        SELECT 1 FROM portal_users WHERE id = $1 AND school_id = $2 AND account_type = $4
		    )
		),
		course_name AS (
		    SELECT name FROM courses WHERE id = $3
		),
		student_name AS (
		    SELECT name FROM portal_users WHERE id = $1 AND account_type = $4
		)
		SELECT
		    (SELECT exists FROM exists),
		    (SELECT name FROM course_name),
		    (SELECT name FROM student_name)
	`, p.StudentID, schoolID, courseID, helpers.AccTypeStudent).Scan(&belongsToSchool, &courseName, &studentName); err != nil {
		log.Println(err)
		return err
	}

	if !belongsToSchool {
		return ErrForbidden
	}

	var query string
	var logAction schools.Action
	var logType schools.Type
	var logTitle string
	var logDesc string
	var logDetails string
	switch assign {
	case true:
		logAction = schools.ActionCourseStudentAssign
		logType = schools.TypeCreate
		logTitle = "Course student assigned"
		logDesc = fmt.Sprintf("A student was assigned to the course %s.", courseName)
		logDetails = fmt.Sprintf("{user} assigned %s to the course %s.", studentName, courseName)
		query = `
INSERT INTO assigned_course_students (portal_user_id, course_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`
	default:
		logAction = schools.ActionCourseStudentRemove
		logType = schools.TypeDelete
		logTitle = "Course student removed"
		logDesc = fmt.Sprintf("A student was removed from the course %s.", courseName)
		logDetails = fmt.Sprintf("{user} removed %s from the course %s.", studentName, courseName)
		query = `
DELETE FROM assigned_course_students
WHERE portal_user_id = $1 AND course_id = $2
`
	}

	res, err := tx.ExecContext(ctx, query, p.StudentID, courseID)
	if err != nil {
		log.Println(err)
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}

	if affected != 0 {
		if err := schools.StoreSchoolLog(schoolID, userID, logAction, logType, logTitle, logDesc, tx, ctx, svc.Sf, logDetails); err != nil {
			log.Println(err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
