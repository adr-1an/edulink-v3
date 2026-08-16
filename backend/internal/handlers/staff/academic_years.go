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

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func ListAcademicYearsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionAcademicYearList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// AC Year struct
	type AcYear struct {
		ID        string `json:"id"`
		StartYear int    `json:"startYear"`
		EndYear   int    `json:"endYear"`
		IsActive  bool   `json:"isActive"`
	}
	var years []AcYear

	// Get rows
	rows, err := db.QueryContext(ctx, `
		SELECT id, start_year, end_year, is_active
		FROM academic_years
		WHERE school_id = $1
	`, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var y AcYear

		if err := rows.Scan(&y.ID, &y.StartYear, &y.EndYear, &y.IsActive); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		years = append(years, y)
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

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":        access,
		"academicYears": years,
	})
}

func CreateAcademicYearHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionAcademicYearCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		AcademicYear struct {
			From int `json:"from"`
			To   int `json:"to"`
		} `json:"academicYear"`
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

	// If you're here in/after the year 10,000, add an extra 9 to the 9999 below
	// Or if you've gone back to before 1900, IDK anymore maybe change it to like 100 or smth
	// Also I had to write "IDK" in all caps or GoLand would cry
	// I know it's corny
	if p.AcademicYear.From < 1900 || p.AcademicYear.From > 9999 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if p.AcademicYear.To < p.AcademicYear.From {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Check for conflicts
	var exists bool
	if err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM academic_years
		 	WHERE school_id = $1
		 	AND start_year = $2
		 	AND end_year = $3
		)
	`, schoolID, p.AcademicYear.From, p.AcademicYear.To).Scan(&exists); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeAcademicYearConflict,
		})
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Insert academic year
	res, err := tx.ExecContext(ctx, `
		INSERT INTO academic_years (id, school_id, start_year, end_year)
		VALUES ($1, $2, $3, $4)
	`, id, schoolID, p.AcademicYear.From, p.AcademicYear.To)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	yearLabel := strconv.Itoa(p.AcademicYear.From) + "-" + strconv.Itoa(p.AcademicYear.To)
	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionAcademicYearCreate, schools.TypeCreate, "Academic year created", "{user} created the academic year "+yearLabel+".", tx, ctx, sf, "{user} created academic year "+yearLabel+" with ID "+strconv.FormatInt(id, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func DeleteAcademicYearHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get year ID
	acYearIDStr := chi.URLParam(r, "yearID")
	acYearID, err := strconv.ParseInt(acYearIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

	// Get the year's data
	var yearSchoolID int64
	var startYear int
	var endYear int
	if err := tx.QueryRowContext(ctx, `
		SELECT school_id, start_year, end_year FROM academic_years WHERE id = $1
	`, acYearID).Scan(&yearSchoolID, &startYear, &endYear); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionAcademicYearDelete, userID, yearSchoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the current year is active
	var active bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM academic_years
			WHERE id = $1
			AND is_active = TRUE
		)
	`, acYearID).Scan(&active); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if active {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Delete the year
	res, err := tx.ExecContext(ctx, `
		DELETE FROM academic_years
		WHERE id = $1
	  	AND school_id = $2
		AND is_active = FALSE
	`, acYearID, yearSchoolID)
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

	yearLabel := strconv.Itoa(startYear) + "-" + strconv.Itoa(endYear)
	if err := schools.StoreSchoolLog(yearSchoolID, userID, schools.ActionAcademicYearDelete, schools.TypeDelete, "Academic year deleted", "{user} deleted the academic year "+yearLabel+".", tx, ctx, sf, "{user} deleted academic year "+yearLabel+" with ID "+strconv.FormatInt(acYearID, 10)+"."); err != nil {
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

func ClearAcademicYearHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionAcademicYearToggleActive, userID, schoolID, ctx, db) {
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

	var activeYearID int64
	var activeStartYear int
	var activeEndYear int
	if err := tx.QueryRowContext(ctx, `
		SELECT id, start_year, end_year
		FROM academic_years
		WHERE school_id = $1
		AND is_active = TRUE
	`, schoolID).Scan(&activeYearID, &activeStartYear, &activeEndYear); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Deactivate current year
	res, err := tx.ExecContext(ctx, `
		UPDATE academic_years
		SET is_active = FALSE
		WHERE school_id = $1
		AND is_active = TRUE
	`, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := "{user} cleared the active academic year."
	if activeYearID != 0 {
		details = "{user} cleared active academic year " + strconv.Itoa(activeStartYear) + "-" + strconv.Itoa(activeEndYear) + " with ID " + strconv.FormatInt(activeYearID, 10) + "."
	}
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolEdit, schools.TypeEdit, "Active academic year cleared", "{user} cleared the active academic year.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
