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
)

func GradeSubmissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	subID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := SubToSchoolID(subID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionSubmissionGrade, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type payload struct {
		Score int    `json:"score"`
		Notes string `json:"notes"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.Notes = strings.TrimSpace(p.Notes)
	if p.Score < 0 || p.Score > 100 || len(p.Notes) > 2048 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var exists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM assignment_submissions WHERE id = $1 AND (status = 'pending' OR status = 'returned')
		)
	`, subID).Scan(&exists); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO submission_scores (submission_id, graded_by, score_percentage, notes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (submission_id) DO UPDATE
		SET graded_by = EXCLUDED.graded_by,
		    score_percentage = EXCLUDED.score_percentage,
		    notes = EXCLUDED.notes,
		    graded_at = NOW()
	`, subID, userID, p.Score, p.Notes); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func ClearSubmissionScoreHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	subID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := SubToSchoolID(subID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionSubmissionRemoveGrade, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if _, err := db.ExecContext(ctx, `
		DELETE FROM submission_scores WHERE submission_id = $1
	`, subID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
