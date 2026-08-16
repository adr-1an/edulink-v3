package staff

import (
	"app/internal/helpers/staff"
	schools2 "app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

type payload struct {
	RefPostID          *string    `json:"referencedPostId"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	DueDate            *time.Time `json:"dueDate"`
	SubmissionsEnabled bool       `json:"submissionsEnabled"`
	SubmissionsCloseAt *time.Time `json:"submissionsCloseAt"`
}

func parseAssignmentPayload(p payload) payload {
	p.Title = strings.TrimSpace(p.Title)
	p.Description = strings.TrimSpace(p.Description)

	return p
}

func validateAssignmentPayload(p payload, editing bool) error {
	if len(p.Title) < 3 || len(p.Title) > 64 {
		return errors.New("invalid title")
	}

	if len(p.Description) > 4096 {
		return errors.New("invalid description")
	}

	if !editing {
		if p.DueDate != nil {
			if p.DueDate.Before(time.Now()) {
				return errors.New("invalid due date")
			}
		}

		if p.SubmissionsCloseAt != nil {
			if p.SubmissionsCloseAt.Before(time.Now()) {
				return errors.New("invalid submissions close date")
			}

			if p.DueDate != nil {
				if p.SubmissionsCloseAt.Before(*p.DueDate) {
					return errors.New("invalid submissions close date")
				}
			}
		}
	}

	return nil
}

func courseToSchoolID(courseID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
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

func assignmentToSchoolID(assignmentID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignments a
		JOIN courses c
		ON a.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE a.id = $1
	`, assignmentID).Scan(&schoolID); err != nil {
		return 0, err
	}

	return schoolID, nil
}

func assignmentToCourseID(assignmentID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var courseID int64
	if err := db.QueryRowContext(ctx, `
		SELECT course_id FROM assignments WHERE id = $1
	`, assignmentID).Scan(&courseID); err != nil {
		return 0, err
	}

	return courseID, nil
}

func CreateAssignmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	schoolID, err := courseToSchoolID(courseID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p = parseAssignmentPayload(p)
	if err := validateAssignmentPayload(p, false); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var refPostSchoolID int64
	if p.RefPostID != nil {
		if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM course_posts p
		JOIN courses c
		ON p.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE p.id = $1
	`, p.RefPostID).Scan(&refPostSchoolID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if refPostSchoolID != schoolID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO assignments (
		                         id,
		                         course_id,
		                         referenced_post_id,
		                         title,
		                         description,
		                         due_date,
		                         submissions_enabled,
		                         submissions_close_at
		                         )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		id,
		courseID,
		p.RefPostID,
		p.Title,
		p.Description,
		p.DueDate,
		p.SubmissionsEnabled,
		p.SubmissionsCloseAt,
	); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var submissions string
	if p.SubmissionsEnabled {
		submissions = "enabled"
	} else {
		submissions = "disabled"
	}

	details := fmt.Sprintf("{user} created the assignment '%s' with submissions %s.", p.Title, submissions)
	if err := schools2.StoreSchoolLog(schoolID, userID, schools2.ActionAssignmentCreate, schools2.TypeCreate, "Assignment created", "{user} created an assignment", tx, ctx, sf, details); err != nil {
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

func ListAssignmentsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	schoolID, err := courseToSchoolID(courseID, db, ctx)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type post struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	type assignment struct {
		RefPost            *post      `json:"referencedPost"`
		ID                 string     `json:"id"`
		Title              string     `json:"title"`
		Description        string     `json:"description"`
		DueDate            *time.Time `json:"dueDate"`
		SubmissionsEnabled bool       `json:"submissionsEnabled"`
		SubmissionsCloseAt *time.Time `json:"submissionsCloseAt"`
		CreatedAt          time.Time  `json:"createdAt"`
	}
	var assignments []assignment

	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.id, p.title,
		    
		    a.id,
		    a.title,
		    a.description,
		    a.due_date,
		    a.submissions_enabled,
		    a.submissions_close_at,
		    a.created_at
		FROM assignments a
		LEFT JOIN course_posts p
		ON a.referenced_post_id = p.id
		WHERE a.course_id = $1
	`, courseID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			a            assignment
			refPostID    sql.NullInt64
			refPostTitle sql.NullString
		)

		if err := rows.Scan(
			&refPostID,
			&refPostTitle,
			&a.ID,
			&a.Title,
			&a.Description,
			&a.DueDate,
			&a.SubmissionsEnabled,
			&a.SubmissionsCloseAt,
			&a.CreatedAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if refPostID.Valid {
			a.RefPost = &post{
				ID:    strconv.FormatInt(refPostID.Int64, 10),
				Title: refPostTitle.String,
			}
		}

		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools2.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":      access,
		"assignments": assignments,
	})
}

func UpdateAssignmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "assignmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := assignmentToSchoolID(assignmentID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p = parseAssignmentPayload(p)
	if err = validateAssignmentPayload(p, true); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if p.RefPostID != nil {
		courseID, err := assignmentToCourseID(assignmentID, db, ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		var postCourseID int64
		if err := db.QueryRowContext(ctx, `
			SELECT course_id FROM course_posts WHERE id = $1
		`, p.RefPostID).Scan(&postCourseID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if postCourseID != courseID {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		UPDATE assignments
		SET
		    referenced_post_id = $1,
		    title = $2,
		    description = $3,
		    due_date = $4,
		    submissions_enabled = $5,
		    submissions_close_at = $6
		WHERE id = $7
	`,
		p.RefPostID,
		p.Title,
		p.Description,
		p.DueDate,
		p.SubmissionsEnabled,
		p.SubmissionsCloseAt,
		assignmentID,
	); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} edited the assignment '%s'.", p.Title)
	if err := schools2.StoreSchoolLog(schoolID, userID, schools2.ActionAssignmentEdit, schools2.TypeEdit, "Assignment updated", "{user} edited an assignment.", tx, ctx, sf, details); err != nil {
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

func DeleteAssignmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "assignmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schoolID, err := assignmentToSchoolID(assignmentID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools2.Can(schools2.PermissionCourseAssignmentDelete, userID, schoolID, ctx, db) {
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

	var assignmentName string
	if err := db.QueryRowContext(ctx, `
		SELECT title FROM assignments WHERE id = $1
	`, assignmentID).Scan(&assignmentName); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM assignments
		WHERE id = $1
	`, assignmentID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} deleted the assignment '%s'.", assignmentName)
	if err := schools2.StoreSchoolLog(schoolID, userID, schools2.ActionAssignmentDelete, schools2.TypeDelete, "Assignment deleted", "{user} deleted an assignment.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
