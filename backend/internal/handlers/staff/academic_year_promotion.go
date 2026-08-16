package staff

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

// SchoolPromotionHandler promotes all the grades in the currently active Academic Year.
// The promotion process is roughly:
// - Creating a new Academic Year
// - Optionally copying grades, courses, student/teacher/parent accounts etc.
// - If grades are selected for copying, their level value can optionally be bumped by 1.
func SchoolPromotionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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
	if !schools.Can(schools.PermissionSchoolPromote, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// -- 1. Get currently active academic year --
	var activeYearID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM academic_years
		WHERE school_id = $1
			AND is_active = true
		LIMIT 1
	`, schoolID).Scan(&activeYearID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeNoActiveYear,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// -- 2. Payload & options --
	type Payload struct {
		NewAcademicYear struct {
			From int `json:"from"`
			To   int `json:"to"`
		} `json:"newAcademicYear"`
		Opts struct {
			ActivateAfterPromotion bool `json:"activateAfterPromotion"`
			TransferGrades         bool `json:"transferGrades"`
			PromoteGradeLevels     bool `json:"promoteGradeLevels"`
			// TODO: Uncomment features below once they're added
			// TransferCourses bool `json:"transferCourses"`
			// TransferStudents       bool `json:"transferStudents"`
			// TransferTeachers bool `json:"transferTeachers"`
			// TransferParents bool `json:"transferParents"`
		} `json:"options"`
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
	if p.NewAcademicYear.From < 1900 || p.NewAcademicYear.From > 9999 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if p.NewAcademicYear.To < p.NewAcademicYear.From {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Check for start & end year conflicts
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM academic_years WHERE start_year = $1 AND end_year = $2 AND school_id = $3
		)
	`, p.NewAcademicYear.From, p.NewAcademicYear.To, schoolID).Scan(&exists); err != nil {
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

	// -- 3. Create new school year --
	// Generate ID
	newAcademicYearID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO academic_years (id, school_id, start_year, end_year)
		VALUES ($1, $2, $3, $4)
	`, newAcademicYearID, schoolID, p.NewAcademicYear.From, p.NewAcademicYear.To)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Activate new academic year if requested
	if p.Opts.ActivateAfterPromotion {
		// Deactivate old
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

		// Activate new
		res, err = tx.ExecContext(ctx, `
			UPDATE academic_years
			SET is_active = true
			WHERE id = $1
			AND is_active = false
		`, newAcademicYearID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
			return
		}
	}

	// -- 4. Transfer old data (on request) --

	// - Transfer grades -
	if p.Opts.TransferGrades {
		// Select all grades from the old year
		type grade struct {
			id    int64
			level int
			name  string
		}
		var grades []grade
		rows, err := tx.QueryContext(ctx, `
		SELECT id, level, name FROM grades WHERE academic_year_id = $1
	`, activeYearID)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() { _ = rows.Close() }()

		// Scan rows
		for rows.Next() {
			var g grade

			if err := rows.Scan(&g.id, &g.level, &g.name); err != nil {
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

		if len(grades) > 0 {
			// For each grade, generate a new ID
			for i := range grades {
				newID, err := sf.NextID()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Increase level by 1 if requested
				if p.Opts.PromoteGradeLevels && p.Opts.TransferGrades {
					grades[i].name = strings.ReplaceAll(grades[i].name, strconv.Itoa(grades[i].level), strconv.Itoa(grades[i].level+1))
					grades[i].level++
				}

				grades[i].id = newID
			}

			// Insert grades
			var (
				values []string
				args   []any
			)
			for i, g := range grades {
				base := i * 4

				values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)",
					base+1,
					base+2,
					base+3,
					base+4,
				))

				args = append(args, g.id, newAcademicYearID, g.level, g.name)
			}

			// The "(0,0,0,0)" is a workaround. Putting %s inside the query would trigger a false SQL parsing error.
			query := "INSERT INTO grades (id, academic_year_id, level, name) VALUES (0,0,0,0)"
			query = strings.ReplaceAll(query, "(0,0,0,0)", strings.Join(values, ","))
			res, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
				return
			}
		}
	}

	// TODO: - Transfer courses once implemented -

	// Log action
	promotionDetails := "{user} promoted the school to academic year " + strconv.Itoa(p.NewAcademicYear.From) + "-" + strconv.Itoa(p.NewAcademicYear.To) + "."
	if p.Opts.ActivateAfterPromotion {
		promotionDetails += " The new academic year was activated."
	}
	if p.Opts.TransferGrades {
		promotionDetails += " Grades were transferred."
	}
	if p.Opts.PromoteGradeLevels {
		promotionDetails += " Grade levels were promoted."
	}
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolPromote, schools.TypeOther, "School promoted", "{user} promoted the school to academic year "+strconv.Itoa(p.NewAcademicYear.From)+"-"+strconv.Itoa(p.NewAcademicYear.To)+".", tx, ctx, sf, promotionDetails); err != nil {
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
