package staff

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func parseCoursePayload(name, description, color string) (string, string, string) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	color = strings.TrimSpace(strings.TrimPrefix(color, "#"))

	return name, description, color
}

func validateCoursePayload(name, description, color string) bool {
	// Validate
	if name == "" || len(name) > 32 {
		return false
	}
	if len(description) > 128 {
		return false
	}
	if color == "" || len(color) != 6 { // Has to be hex
		return false
	}

	return true
}

func checkNameConflict(name string, gradeID int64, tx *sql.Tx, ctx context.Context) (bool, error) {
	var conflict bool

	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM courses WHERE lower(name) = $1 AND grade_id = $2
		)
	`, strings.ToLower(name), gradeID).Scan(&conflict); err != nil {
		return false, err
	}

	return conflict, nil
}

func checkNameConflictExcept(name string, gradeID, courseID int64, tx *sql.Tx, ctx context.Context) (bool, error) {
	var conflict bool

	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM courses WHERE lower(name) = $1 AND grade_id = $2 AND id <> $3
		)
	`, strings.ToLower(name), gradeID, courseID).Scan(&conflict); err != nil {
		return false, err
	}

	return conflict, nil
}

func getCourseGradeAndSchoolID(db *sql.DB, ctx context.Context, courseID int64) (int64, int64, error) {
	// Get the course's school ID and grade ID
	var schoolID int64
	var gradeID int64
	if err := db.QueryRowContext(ctx, `
		SELECT s.id, c.grade_id
		FROM grades g
		JOIN courses c
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		JOIN schools s
		ON y.school_id = s.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID, &gradeID); err != nil {
		return 0, 0, err
	}

	return gradeID, schoolID, nil
}

func CreateCourseHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get grade ID
	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the grade's school ID
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM academic_years y
		JOIN grades g
		ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionCourseCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)
	p.Description = strings.TrimSpace(p.Description)
	p.Color = strings.TrimSpace(strings.TrimPrefix(p.Color, "#"))

	// Validate
	if !validateCoursePayload(p.Name, p.Description, p.Color) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Check for course name conflicts
	conflict, err := checkNameConflict(p.Name, gradeID, tx, ctx)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store course
	res, err := tx.ExecContext(ctx, `
		INSERT INTO courses (id, name, description, color, grade_id)
		VALUES ($1, $2, $3, $4, $5)
	`, id, p.Name, p.Description, p.Color, gradeID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionCourseCreate, schools.TypeCreate, "Course created", "{user} created the course '"+p.Name+"'.", tx, ctx, sf, "{user} created course '"+p.Name+"' with color '#"+p.Color+"' and description '"+p.Description+"'."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateCourseHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get course ID
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get grade and school ID
	gradeID, schoolID, err := getCourseGradeAndSchoolID(db, ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionCourseUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name, p.Description, p.Color = parseCoursePayload(p.Name, p.Description, p.Color)

	// Validate
	if !validateCoursePayload(p.Name, p.Description, p.Color) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var oldName string
	var oldDescription string
	var oldColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, description, color
		FROM courses
		WHERE id = $1
	`, courseID).Scan(&oldName, &oldDescription, &oldColor); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check for course name conflicts
	conflict, err := checkNameConflictExcept(p.Name, gradeID, courseID, tx, ctx)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if conflict {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Update course
	res, err := tx.ExecContext(ctx, `
		UPDATE courses c
		SET name = $1, description = $2, color = $3
		WHERE id = $4
		AND grade_id = $5
	`, p.Name, p.Description, p.Color, courseID, gradeID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	details := "{user} updated the course name from '" + oldName + "' to '" + p.Name + "', description from '" + oldDescription + "' to '" + p.Description + "', and color from '#" + oldColor + "' to '#" + p.Color + "'."
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionCourseEdit, schools.TypeEdit, "Course updated", "{user} updated the course '"+p.Name+"'.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteCourseHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	_, schoolID, err := getCourseGradeAndSchoolID(db, ctx, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionCourseDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var courseName string
	if err := tx.QueryRowContext(ctx, `
		SELECT name FROM courses WHERE id = $1
	`, courseID).Scan(&courseName); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionCourseDelete, schools.TypeDelete, "Course deleted", "{user} deleted the course '"+courseName+"'.", tx, ctx, sf, "{user} deleted course '"+courseName+"' with ID "+strconv.FormatInt(courseID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListCoursesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM grades g
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionCourseList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	type Course struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}
	var courses []Course

	rows, err := tx.QueryContext(ctx, `
		SELECT id, name, description, color
		FROM courses c
		WHERE grade_id = $1
	`, gradeID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var c Course

		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user access
	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":  access,
		"courses": courses,
	})
}
