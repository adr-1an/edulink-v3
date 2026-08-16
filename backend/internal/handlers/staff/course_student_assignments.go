package staff

import (
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func getCourseSchoolID(courseID int64, tx *sql.Tx) (int64, error) {
	var schoolID int64
	if err := tx.QueryRow(`
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		return 0, err
	}

	return schoolID, nil
}

func AddOrRemoveCourseStudentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake, assign bool) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	schoolID, err := getCourseSchoolID(courseID, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if assign {
		if !schools.Can(schools.PermissionCourseStudentAssign, userID, schoolID, ctx, db) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	} else {
		if !schools.Can(schools.PermissionCourseStudentRemove, userID, schoolID, ctx, db) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	type Payload struct {
		StudentID string `json:"studentId"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.StudentID == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !belongsToSchool {
		w.WriteHeader(http.StatusForbidden)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if affected != 0 {
		if err := schools.StoreSchoolLog(schoolID, userID, logAction, logType, logTitle, logDesc, tx, ctx, sf, logDetails); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if assign {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListCourseStudentsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	schoolID, err := getCourseSchoolID(courseID, tx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionCourseStudentList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type student struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		LastName string `json:"lastName"`
		Email    string `json:"email"`
		Assigned bool   `json:"assigned"`
	}
	var students []student

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var s student

		if err := rows.Scan(&s.ID, &s.Name, &s.LastName, &s.Email, &s.Assigned); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   access,
		"students": students,
	})
}
