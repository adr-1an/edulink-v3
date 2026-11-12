package schoollookup

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

// YearToSchoolID returns the ID of the school that the Academic Year belongs to.
func YearToSchoolID(db *sql.DB, w http.ResponseWriter, r *http.Request, ctx context.Context) (int64, int64, error) {
	yearIDStr := chi.URLParam(r, "yearID")
	yearID, err := strconv.ParseInt(yearIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 0, 0, errors.New("yearID is not a number")
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT
		    school_id
		FROM academic_years
		WHERE id = $1
	`, yearID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, 0, err
	}

	return schoolID, yearID, nil
}

// GradeToSchoolID returns the ID of the school that the Grade belongs to.
func GradeToSchoolID(db *sql.DB, w http.ResponseWriter, r *http.Request, ctx context.Context) (int64, int64, error) {
	gradeIDStr := chi.URLParam(r, "gradeID")
	gradeID, err := strconv.ParseInt(gradeIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 0, 0, errors.New("gradeID is not a number")
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM academic_years y
		JOIN grades g ON g.academic_year_id = y.id
		WHERE g.id = $1
	`, gradeID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, 0, err
	}

	return schoolID, gradeID, nil
}
