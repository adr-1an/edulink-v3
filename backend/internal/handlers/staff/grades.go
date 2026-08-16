package staff

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
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

func ListGradesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolIDStr := chi.URLParam(r, "schoolID")
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionGradeList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Grade struct
	type Grade struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	var grades []Grade

	// Fetch from DB
	rows, err := db.QueryContext(ctx, `
		SELECT g.id, g.name, g.level
		FROM grades g
		JOIN academic_years y
		ON y.id = g.academic_year_id
		WHERE y.school_id = $1
	`, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var g Grade

		if err := rows.Scan(&g.ID, &g.Name, &g.Level); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		grades = append(grades, g)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user permissions
	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return grades
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"grades": grades,
	})
}

func CreateGradeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, yearID, err := schools.YearToSchoolID(db, w, r, ctx)
	if err != nil {
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionGradeCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse & validate
	p.Name = strings.TrimSpace(p.Name)

	if !strings.Contains(p.Name, "{level}") {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeMissingLevelVar,
		})
		return
	}

	if p.Name == "" || len(p.Name) < 7 || len(p.Name) > 32 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidName,
		})
		return
	}

	if p.Level < 0 || p.Level > 20 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeLevelOutOfRange,
		})
		return
	}

	// Parse the {level} variable
	p.Name = strings.Replace(p.Name, "{level}", strconv.Itoa(p.Level), 1)

	// Check for grade level & name conflict
	var exists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM grades WHERE academic_year_id = $1 AND name = $2 AND level = $3
		)
	`, yearID, p.Name, p.Level).Scan(&exists); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
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

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Store grade
	res, err := tx.ExecContext(ctx, `
		INSERT INTO grades (id, academic_year_id, level, name)
		VALUES ($1, $2, $3, $4)
	`, id, yearID, p.Level, p.Name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionGradeCreate, schools.TypeCreate, "Grade created", "{user} created the grade '"+p.Name+"'.", tx, ctx, sf, "{user} created grade '"+p.Name+"' at level "+strconv.Itoa(p.Level)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func UpdateGradeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, gradeID, err := schools.GradeToSchoolID(db, w, r, ctx)
	if err != nil {
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionGradeUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse & validate
	p.Name = strings.TrimSpace(p.Name)
	if !strings.Contains(p.Name, "{level}") {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeMissingLevelVar,
		})
		return
	}

	if p.Name == "" || len(p.Name) < 7 || len(p.Name) > 32 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidName,
		})
		return
	}

	if p.Level < 0 || p.Level > 20 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeLevelOutOfRange,
		})
		return
	}

	// Parse name level var
	p.Name = strings.Replace(p.Name, "{level}", strconv.Itoa(p.Level), 1)

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Get the grade's current data
	var yearID int64
	var oldName string
	var oldLevel int
	if err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id, name, level
		FROM grades
		WHERE id = $1
	`, gradeID).Scan(&yearID, &oldName, &oldLevel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check for name conflict
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM grades WHERE name = $1 AND academic_year_id = $2 AND id <> $3
		)
	`, p.Name, yearID, gradeID).Scan(&exists); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Update grade
	res, err := tx.ExecContext(ctx, `
		UPDATE grades
		SET name = $1, level = $2
		WHERE id = $3
	`, p.Name, p.Level, gradeID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	details := "{user} updated the grade name from '" + oldName + "' to '" + p.Name + "' and level from " + strconv.Itoa(oldLevel) + " to " + strconv.Itoa(p.Level) + "."
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionGradeEdit, schools.TypeEdit, "Grade updated", "{user} updated the grade '"+p.Name+"'.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteGradeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, gradeID, err := schools.GradeToSchoolID(db, w, r, ctx)
	if err != nil {
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionGradeDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
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

	var gradeName string
	var gradeLevel int
	if err := tx.QueryRowContext(ctx, `
		SELECT name, level FROM grades WHERE id = $1
	`, gradeID).Scan(&gradeName, &gradeLevel); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete grade
	res, err := tx.ExecContext(ctx, `DELETE FROM grades WHERE id = $1`, gradeID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionGradeDelete, schools.TypeDelete, "Grade deleted", "{user} deleted the grade '"+gradeName+"'.", tx, ctx, sf, "{user} deleted grade '"+gradeName+"' at level "+strconv.Itoa(gradeLevel)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
