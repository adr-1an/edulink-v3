package staff

import (
	"app/internal/handlers/portal/students"
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

// SubToSchoolID is public because it came in handy in another handler.
func SubToSchoolID(subID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var sID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignment_submissions s
		JOIN assignments a ON s.assignment_id = a.id
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id
		WHERE s.id = $1
	`, subID).Scan(&sID); err != nil {
		return 0, err
	}

	return sID, nil
}

func ListAssignmentSubmissionsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignments a
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id
		WHERE  a.id = $1
	`, assignmentID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionSubmissionList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type portalUser struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Name     string `json:"name"`
		Lastname string `json:"lastName"`
	}

	type submission struct {
		ID          string     `json:"id"`
		Status      string     `json:"status"`
		SubmittedBy portalUser `json:"submittedBy"`
		Notes       *string    `json:"notes"`
		SubmittedAt time.Time  `json:"submittedAt"`
	}
	var subs []submission

	rows, err := db.QueryContext(ctx, `
		SELECT
		    s.id, u.id, u.email, u.phone, u.name, u.last_name,
		    s.notes, s.created_at, s.status
		FROM assignment_submissions s
		JOIN portal_users u ON s.submitted_by = u.id
		WHERE (s.status = $1
		  OR s.status = $4)
		AND s.assignment_id = $2
		AND u.account_type = $3
	`, students.SubmissionStatusSubmitted, assignmentID, helpers2.AccTypeStudent, students.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	type submissionRow struct {
		SubmissionID string

		UserID   string
		Email    string
		Phone    string
		Name     string
		Lastname string

		Notes       *string
		SubmittedAt time.Time
		Status      string
	}

	for rows.Next() {
		var r submissionRow

		if err := rows.Scan(
			&r.SubmissionID,

			&r.UserID,
			&r.Email,
			&r.Phone,
			&r.Name,
			&r.Lastname,

			&r.Notes,
			&r.SubmittedAt,
			&r.Status,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		subs = append(subs, submission{
			ID: r.SubmissionID,

			SubmittedBy: portalUser{
				ID:       r.UserID,
				Email:    r.Email,
				Phone:    r.Phone,
				Name:     r.Name,
				Lastname: r.Lastname,
			},

			Notes:       r.Notes,
			SubmittedAt: r.SubmittedAt,
			Status:      r.Status,
		})
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
		"access":      access,
		"submissions": subs,
	})
}

func ViewSubmissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
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

	if !schools.Can(schools.PermissionSubmissionView, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	type grade struct {
		Score    int       `json:"score"`
		Notes    *string   `json:"notes"`
		GradedAt time.Time `json:"gradedAt"`
		GradedBy user      `json:"gradedBy"`
	}

	type portalUser struct {
		ID       string  `json:"id"`
		Email    string  `json:"email"`
		Phone    *string `json:"phone"`
		Name     string  `json:"name"`
		Lastname string  `json:"lastName"`
	}

	type submissionAttachment struct {
		ID           string `json:"id"`
		PresignedURL string `json:"presignedUrl"`
		Filename     string `json:"originalFilename"`
		ContentType  string `json:"declaredContentType"`
	}

	type submission struct {
		ID          string                 `json:"id"`
		Status      string                 `json:"status"`
		SubmittedBy portalUser             `json:"submittedBy"`
		Attachments []submissionAttachment `json:"attachments"`
		Grade       *grade                 `json:"grade"`
		Notes       *string                `json:"notes"`
		SubmittedAt time.Time              `json:"submittedAt"`
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
		    s.id,
		    s.status,
		    s.notes,
		    s.created_at,
		    
		    ss.score_percentage,
		    ss.notes,
		    ss.graded_at,
		    
		    gu.id,
		    gu.name,
		    gu.email,
		    gu.phone,
		    
		    u.id,
		    u.email,
		    u.phone,
		    u.name,
		    u.last_name,
		    
		    so.id,
		    so.original_file_name,
		    so.declared_content_type,
		    so.bucket_name
		FROM assignment_submissions s
		JOIN portal_users u ON s.submitted_by = u.id
		LEFT JOIN submission_attachments sa ON s.id = sa.submission_id
		LEFT JOIN portal_storage_objects so ON sa.storage_object_id = so.id
		LEFT JOIN submission_scores ss ON s.id = ss.submission_id
		LEFT JOIN users gu ON ss.graded_by = gu.id
		WHERE s.id = $1
		AND (s.status = $2 OR s.status = $3)
	`, subID, students.SubmissionStatusSubmitted, students.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	var (
		s           submission
		initialized bool
	)

	type subRow struct {
		ID        string
		Status    string
		Notes     *string
		CreatedAt time.Time

		GradeScore    *int
		GradeNotes    *string
		GradedAt      *time.Time
		GradedByID    *string
		GradedByName  *string
		GradedByEmail *string
		GradedByPhone *string

		UserID   string
		Email    string
		Phone    string
		Name     string
		Lastname string

		AttachmentID *string
		Filename     *string
		ContentType  *string
		BucketName   *string
	}

	for rows.Next() {
		var r subRow

		if err := rows.Scan(
			&r.ID,
			&r.Status,
			&r.Notes,
			&r.CreatedAt,

			&r.GradeScore,
			&r.GradeNotes,
			&r.GradedAt,

			&r.GradedByID,
			&r.GradedByName,
			&r.GradedByEmail,
			&r.GradedByPhone,

			&r.UserID,
			&r.Email,
			&r.Phone,
			&r.Name,
			&r.Lastname,

			&r.AttachmentID,
			&r.Filename,
			&r.ContentType,
			&r.BucketName,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !initialized {
			s = submission{
				ID:     r.ID,
				Status: r.Status,
				SubmittedBy: portalUser{
					ID:       r.UserID,
					Email:    r.Email,
					Phone:    &r.Phone,
					Name:     r.Name,
					Lastname: r.Lastname,
				},
				Grade:       nil,
				Attachments: make([]submissionAttachment, 0),
				SubmittedAt: r.CreatedAt,
			}

			if r.Notes != nil {
				s.Notes = r.Notes
			}

			initialized = true
		}

		if r.AttachmentID != nil {
			objKey := fmt.Sprintf("%s/%s", helpers2.UploadCategorySubmissionAttachments, *r.AttachmentID)

			url, err := s3.PresignedGetObject(ctx, *r.BucketName, objKey, 5*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			s.Attachments = append(s.Attachments, submissionAttachment{
				ID:           *r.AttachmentID,
				PresignedURL: url.String(),
				Filename:     *r.Filename,
				ContentType:  *r.ContentType,
			})
		}

		if r.GradeScore != nil && r.GradedByID != nil {
			s.Grade = &grade{
				Score:    *r.GradeScore,
				Notes:    r.GradeNotes,
				GradedAt: *r.GradedAt,
				GradedBy: user{
					ID:    *r.GradedByID,
					Name:  *r.GradedByName,
					Email: *r.GradedByEmail,
					Phone: *r.GradedByPhone,
				},
			}
		}
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
		"access":     access,
		"submission": s,
	})
}

func ReturnSubmissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

	if !schools.Can(schools.PermissionSubmissionReturn, userID, schoolID, ctx, db) {
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

	result, err := db.ExecContext(ctx, `
		UPDATE assignment_submissions
		SET status = $1
		WHERE id = $2
		AND status = $3
	`, students.SubmissionStatusReturned, subID, students.SubmissionStatusSubmitted)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rowsAffected != 1 {
		w.WriteHeader(http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteReturnedSubmissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
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

	if !schools.Can(schools.PermissionSubmissionDelete, userID, schoolID, ctx, db) {
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

	var status students.SubmissionStatus
	if err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM assignment_submissions
		WHERE id = $1
		FOR UPDATE
	`, subID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if status != students.SubmissionStatusReturned {
		w.WriteHeader(http.StatusConflict)
		return
	}

	type storageObject struct {
		BucketName string
		ObjectKey  string
	}
	var storageObjects []storageObject

	rows, err := tx.QueryContext(ctx, `
		DELETE FROM portal_storage_objects so
		USING submission_attachments sa
		WHERE sa.storage_object_id = so.id
		AND sa.submission_id = $1
		RETURNING so.bucket_name, so.object_key
	`, subID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var object storageObject
		if err := rows.Scan(&object.BucketName, &object.ObjectKey); err != nil {
			_ = rows.Close()
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		storageObjects = append(storageObjects, object)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := rows.Close(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM assignment_submissions
		WHERE id = $1
		AND status = $2
	`, subID, students.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rowsAffected != 1 {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, object := range storageObjects {
		if err := s3.RemoveObject(context.WithoutCancel(ctx), object.BucketName, object.ObjectKey, minio.RemoveObjectOptions{}); err != nil {
			log.Println(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
