package staff

import (
	"app/internal/helpers"
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
	"time"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sony/sonyflake/v2"
	"golang.org/x/text/language"
)

func SchoolListHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Search params

	// showDeleted := r.URL.Query().Get("showDeleted") == "true"
	// Not in use for now

	// School struct
	type School struct {
		ID         string `json:"id"`
		OwnerID    string `json:"ownerId"`
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}
	var schoolList []School

	// Get all schools the user has access to
	var rows *sql.Rows
	switch false { // Replaced showDeleted with false
	case true:
		rows, err = db.QueryContext(ctx, `
	SELECT DISTINCT s.id, s.owner_id, s.name, s.region_code
		FROM schools s
		WHERE s.owner_id = $1
	   	OR EXISTS (
			SELECT 1
			FROM school_staff ss
		  	WHERE ss.user_id = $1
	   )
`, userID, schools.PermissionSchoolView)
	default:
		rows, err = db.QueryContext(ctx, `
		SELECT DISTINCT s.id, s.owner_id, s.name, s.region_code
		FROM schools s
		WHERE s.owner_id = $1
   		AND s.deleted_at IS NULL
	   	OR EXISTS (
			SELECT 1
			FROM school_staff ss
		  	WHERE ss.user_id = $1
		  	AND deleted_at IS NULL
	   )
	`, userID)
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var school School

		if err := rows.Scan(&school.ID, &school.OwnerID, &school.Name, &school.RegionCode); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		schoolList = append(schoolList, school)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schools": schoolList,
	})
}

func CreateSchoolHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
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
	p.RegionCode = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(p.RegionCode, " ", "")))

	if p.Name == "" || len(p.Name) > 64 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidName,
		})
		return
	}

	if p.RegionCode != "" {
		region, err := language.ParseRegion(p.RegionCode)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !region.IsCountry() {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidRegionCode,
			})
			return
		}
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

	// Store
	res, err := tx.ExecContext(ctx, `
		INSERT INTO schools (id, owner_id, name, region_code)
		VALUES ($1, $2, $3, $4)
	`, id, userID, p.Name, p.RegionCode)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(id, userID, schools.ActionSchoolCreate, schools.TypeCreate, "School created", "{user} created the school '"+p.Name+"'.", tx, ctx, sf, "{user} created school '"+p.Name+"' in region '"+p.RegionCode+"'."); err != nil {
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

func UpdateSchoolHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	// Payload
	type Payload struct {
		Name                 string `json:"name"`
		RegionCode           string `json:"regionCode"`
		ActiveAcademicYearID *int64 `json:"activeAcademicYearId"`
	}
	var p Payload

	// School permission check
	if !schools.Can(schools.PermissionSchoolUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.ActiveAcademicYearID != nil &&
		// If p.ActiveAcademicYearID isn't null, it means the user is trying to change the currently active academic year.
		// To do that, they need the academicYear.toggle permission.
		!schools.Can(schools.PermissionAcademicYearToggleActive, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Parse & validate
	p.Name = strings.TrimSpace(p.Name)
	p.RegionCode = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(p.RegionCode, " ", "")))

	if p.Name == "" || len(p.Name) > 64 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidName,
		})
		return
	}

	if p.RegionCode != "" {
		region, err := language.ParseRegion(p.RegionCode)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !region.IsCountry() {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidRegionCode,
			})
			return
		}
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
	var oldRegionCode string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, region_code
		FROM schools
		WHERE id = $1
	`, schoolID).Scan(&oldName, &oldRegionCode); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	activeYearDetails := ""
	if p.ActiveAcademicYearID != nil {
		// Get the Academic Year's school ID
		var yearSchoolID int64
		var newStartYear int
		var newEndYear int
		if err := tx.QueryRowContext(ctx, `
		SELECT school_id, start_year, end_year FROM academic_years WHERE id = $1
	`, p.ActiveAcademicYearID).Scan(&yearSchoolID, &newStartYear, &newEndYear); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		if yearSchoolID != schoolID {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		activeYearDetails = " Active academic year was set to " + strconv.Itoa(newStartYear) + "-" + strconv.Itoa(newEndYear) + "."

		// Deactivate old year
		if _, err := tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = false
			WHERE school_id = $1
			AND is_active = true
		`, schoolID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Activate new year
		res, err := tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = true
			WHERE id = $1
			AND is_active = false
		`, p.ActiveAcademicYearID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
			return
		}
	}

	// Update
	res, err := tx.ExecContext(ctx, `
		UPDATE schools
		SET name = $1, region_code = $2, updated_at = NOW()
		WHERE id = $3
	`, p.Name, p.RegionCode, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	details := "{user} updated the school name from '" + oldName + "' to '" + p.Name + "' and region from '" + oldRegionCode + "' to '" + p.RegionCode + "'." + activeYearDetails
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolEdit, schools.TypeEdit, "School updated", "{user} updated the school '"+p.Name+"'.", tx, ctx, sf, details); err != nil {
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

func DeleteSchoolHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school id
	schoolIDStr := chi.URLParam(r, "schoolID")
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Permission check
	if !schools.Can(schools.SchoolOwner, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if the account has 2fa enabled
	var tfa string
	if err := db.QueryRowContext(ctx, `
		SELECT two_factor_status FROM users WHERE id = $1
	`, userID).Scan(&tfa); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if tfa == TwoFAStatusEnabled {
		// Generate challenge token
		token, err := gonanoid.New(128)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tokenHash := helpers.MakeHash256(token)

		id, err := sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		meta := SchoolDeletionChallengeMetadata{
			UserID:   userID,
			SchoolID: schoolID,
		}

		metadata, err := json.Marshal(meta)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(15 * time.Minute)
		if _, err := db.ExecContext(ctx, `
			INSERT INTO two_factor_challenges (id, user_id, purpose, token_hash, expires_at, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, id, userID, ChallengePurposeSchoolDeletion, tokenHash, expiresAt, metadata); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return challenge data
		_ = json.NewEncoder(w).Encode(map[string]any{
			"twoFactorChallenge": map[string]any{
				"token":     token,
				"purpose":   ChallengePurposeSchoolDeletion,
				"expiresAt": expiresAt,
			},
		})
		return
	}

	CompleteSchoolDeletion(CompleteSchoolDeletionPayload{
		W:   w,
		DB:  db,
		Sf:  sf,
		Ctx: ctx,
		Meta: SchoolDeletionChallengeMetadata{
			UserID:   userID,
			SchoolID: schoolID,
		},
	})
}

type CompleteSchoolDeletionPayload struct {
	W    http.ResponseWriter
	DB   *sql.DB
	Sf   *sonyflake.Sonyflake
	Ctx  context.Context
	Meta SchoolDeletionChallengeMetadata
}

func CompleteSchoolDeletion(p CompleteSchoolDeletionPayload) {
	// Start tx
	tx, err := p.DB.BeginTx(p.Ctx, nil)
	if err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var schoolName string
	if err := tx.QueryRowContext(p.Ctx, `
		SELECT name FROM schools WHERE id = $1
	`, p.Meta.SchoolID).Scan(&schoolName); err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete school (update deleted_at)
	res, err := tx.ExecContext(p.Ctx, `
		UPDATE schools
		SET deleted_at = NOW()
		WHERE id = $1
	`, p.Meta.SchoolID)
	if err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, p.W); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(p.Meta.SchoolID, p.Meta.UserID, schools.ActionSchoolDelete, schools.TypeDelete, "School deleted", "{user} deleted the school '"+schoolName+"'.", tx, p.Ctx, p.Sf, "{user} deleted school '"+schoolName+"' with ID "+strconv.FormatInt(p.Meta.SchoolID, 10)+"."); err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	p.W.WriteHeader(http.StatusNoContent)
}

func ViewSchoolDashboardHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
	if !schools.Can(schools.PermissionSchoolView, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Structs
	type Grade struct {
		ID             string `json:"id"`
		AcademicYearID string `json:"academicYearId"`
		Level          int    `json:"level"`
		Name           string `json:"name"`
		CreatedAt      string `json:"createdAt"`
	}

	type School struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		RegionCode string  `json:"regionCode"`
		Grades     []Grade `json:"grades"`
		CreatedAt  string  `json:"createdAt"`
		UpdatedAt  string  `json:"updatedAt"`
	}

	var s School
	var gs []Grade

	// Get school info
	if err := db.QueryRowContext(ctx, `
		SELECT
		    id, name, region_code, created_at, updated_at
		FROM schools
		WHERE id = $1
		AND deleted_at IS NULL
	`, schoolID).Scan(
		&s.ID,
		&s.Name,
		&s.RegionCode,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get grades
	rows, err := db.QueryContext(ctx, `
		SELECT g.id, g.academic_year_id, g.level, g.name, g.created_at
		FROM grades g
		JOIN academic_years y
		ON g.academic_year_id = y.id
		JOIN schools s
		ON y.school_id = s.id
		WHERE s.id = $1
		AND y.is_active = true
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

		if err := rows.Scan(&g.ID, &g.AcademicYearID, &g.Level, &g.Name, &g.CreatedAt); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		gs = append(gs, g)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return data
	s.Grades = gs

	_ = json.NewEncoder(w).Encode(map[string]any{
		"school": s,
	})
}
