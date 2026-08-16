package students

import (
	helpers2 "app/internal/helpers"
	portal2 "app/internal/helpers/portal"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"
	"github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusReturned  SubmissionStatus = "returned"
)

type user struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Phone *string `json:"phone"`
}

type grade struct {
	Score    int       `json:"score"`
	Notes    *string   `json:"notes"`
	GradedAt time.Time `json:"gradedAt"`
	GradedBy user      `json:"gradedBy"`
}

type assignmentSubmission struct {
	ID          string                           `json:"id"`
	Status      SubmissionStatus                 `json:"status"`
	Grade       *grade                           `json:"grade"`
	Notes       string                           `json:"notes"`
	Attachments []assignmentSubmissionAttachment `json:"attachments"`
	CreatedAt   time.Time                        `json:"createdAt"`
}

type assignmentSubmissionAttachment struct {
	ID           string `json:"id"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	ContentType  string `json:"contentType"`
	PresignedURL string `json:"presignedUrl"`
}

func listAssignmentSubmissionAttachments(ctx context.Context, db *sql.DB, submissionIDs []int64, s3 *minio.Client) (map[string][]assignmentSubmissionAttachment, error) {
	attachments := make(map[string][]assignmentSubmissionAttachment)
	if len(submissionIDs) == 0 {
		return attachments, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
		    sa.submission_id,
		    sa.id,
		    so.original_file_name,
		    so.declared_file_size,
		    so.declared_content_type,
		    so.bucket_name,
		    so.object_key
		FROM submission_attachments sa
		JOIN portal_storage_objects so
		ON so.id = sa.storage_object_id
		WHERE sa.submission_id = ANY($1)
		AND so.status = $2
		ORDER BY sa.created_at, sa.id
	`, pq.Array(submissionIDs), helpers2.StatusDone)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var submissionID int64
		var attachment assignmentSubmissionAttachment

		var bucketName string
		var objKey string

		if err := rows.Scan(
			&submissionID,
			&attachment.ID,
			&attachment.FileName,
			&attachment.FileSize,
			&attachment.ContentType,
			&bucketName,
			&objKey,
		); err != nil {
			return nil, err
		}

		url, err := s3.PresignedGetObject(ctx, bucketName, objKey, 15*time.Minute, nil)
		if err != nil {
			return nil, err
		}

		attachment.PresignedURL = url.String()

		id := strconv.FormatInt(submissionID, 10)
		attachments[id] = append(attachments[id], attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

func DeleteSubmissionAttachmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "attachmentID"), 10, 64)
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

	var submittedBy int64
	var status SubmissionStatus
	var storageObjectID int64
	var bucketName string
	var objectKey string
	var canAccessCourse bool
	if err := tx.QueryRowContext(ctx, `
		SELECT
		    s.submitted_by,
		    s.status,
		    so.id,
		    so.bucket_name,
		    so.object_key,
		    EXISTS (
		        SELECT 1
		        FROM assigned_course_students acs
		        WHERE acs.portal_user_id = $3
		        AND acs.course_id = a.course_id
		    )
		FROM assignment_submissions s
		JOIN assignments a
		ON a.id = s.assignment_id
		JOIN submission_attachments sa
		ON sa.submission_id = s.id
		JOIN portal_storage_objects so
		ON so.id = sa.storage_object_id
		WHERE s.id = $1
		AND sa.id = $2
		FOR UPDATE OF s, so
	`, submissionID, attachmentID, userID).Scan(
		&submittedBy,
		&status,
		&storageObjectID,
		&bucketName,
		&objectKey,
		&canAccessCourse,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if submittedBy != userID || !canAccessCourse {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if status != SubmissionStatusPending {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM portal_storage_objects
		WHERE id = $1
	`, storageObjectID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s3.RemoveObject(context.WithoutCancel(ctx), bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateAssignmentSubmissionHandler creates the submission with the "pending" status, to wait for attachment uploads.
func CreateAssignmentSubmissionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "assignmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var courseID int64
	var submissionsOpen bool
	if err := db.QueryRowContext(ctx, `
		SELECT
		    course_id,
		    submissions_enabled
		        AND (submissions_close_at IS NULL OR submissions_close_at > NOW())
		FROM assignments
		WHERE id = $1
	`, assignmentID).Scan(&courseID, &submissionsOpen); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !portal2.CanAccessCourse(db, userID, courseID, ctx) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if !submissionsOpen {
		w.WriteHeader(http.StatusConflict)
		return
	}

	type payload struct {
		Notes string `json:"notes"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(p.Notes) > 2048 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	newSubID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store a new pending submission, or resume the student's existing pending submission.
	var submissionID int64
	if err := db.QueryRowContext(ctx, `
		INSERT INTO assignment_submissions (id, assignment_id, submitted_by, notes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (assignment_id, submitted_by)
		WHERE status <> $6
		DO UPDATE
		SET notes = EXCLUDED.notes
		WHERE assignment_submissions.status = $5
		RETURNING id
	`, newSubID, assignmentID, userID, p.Notes, SubmissionStatusPending, SubmissionStatusReturned).Scan(&submissionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusConflict)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"submissionId": strconv.FormatInt(submissionID, 10),
		"status":       SubmissionStatusPending,
	})
}

func InitSubmissionAttachmentUpload(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type payload struct {
		Name                string `json:"fileName"`
		DeclaredSize        int64  `json:"declaredSize"`
		DeclaredContentType string `json:"declaredContentType"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.DeclaredSize <= 0 || p.DeclaredSize > 50*1024*1024 /* 50mb */ {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	allowed := map[string]bool{
		"image/jpeg":                   true,
		"image/png":                    true,
		"image/gif":                    true,
		"image/webp":                   true,
		"application/zip":              true,
		"application/x-zip-compressed": true,
		"application/pdf":              true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // MS Word
		"application/msword":          true,
		"application/gzip":            true,
		"application/x-gzip":          true,
		"audio/mp4":                   true,
		"audio/wav":                   true,
		"audio/mpeg":                  true,
		"video/mpeg":                  true,
		"video/webm":                  true,
		"application/x-7z-compressed": true,
	}

	if !allowed[p.DeclaredContentType] {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// generate id & completion token
	storageObjID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subAttachmentID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completionToken, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(completionToken)

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	objectKey := fmt.Sprintf("%s/%d", helpers2.UploadCategorySubmissionAttachments, storageObjID)

	// Generate the URL before opening the transaction so no database lock is held
	// while communicating with object storage.
	url, err := s3.PresignedPutObject(ctx, bucketName, objectKey, 5*time.Minute)
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

	var submittedBy int64
	var status SubmissionStatus
	var submissionsOpen bool
	var canAccessCourse bool
	if err := tx.QueryRowContext(ctx, `
		SELECT
		    s.submitted_by,
		    s.status,
		    a.submissions_enabled
		        AND (a.submissions_close_at IS NULL OR a.submissions_close_at > NOW()),
		    EXISTS (
		        SELECT 1
		        FROM assigned_course_students acs
		        WHERE acs.portal_user_id = $2
		        AND acs.course_id = a.course_id
		    )
		FROM assignment_submissions s
		JOIN assignments a
		ON s.assignment_id = a.id
		WHERE s.id = $1
		FOR UPDATE OF s
	`, submissionID, userID).Scan(&submittedBy, &status, &submissionsOpen, &canAccessCourse); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if submittedBy != userID || !canAccessCourse {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if status != SubmissionStatusPending || !submissionsOpen {
		w.WriteHeader(http.StatusConflict)
		return
	}

	s := helpers2.UploadService{
		Db:         tx,
		S3:         s3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	if err := s.PortalStoreStorageObjectsRow(storageObjID, tokenHash, userID, p.Name, p.DeclaredSize, p.DeclaredContentType, objectKey); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO submission_attachments (id, submission_id, storage_object_id)
		VALUES ($1, $2, $3)
	`, subAttachmentID, submissionID, storageObjID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":              strconv.FormatInt(storageObjID, 10),
		"attachmentId":    strconv.FormatInt(subAttachmentID, 10),
		"completionToken": completionToken,
		"url":             url.String(),
	})
}

func CompleteSubmissionCreationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	submissionID, err := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
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

	var assignmentID int64
	var submittedBy int64
	var status SubmissionStatus
	var submissionsOpen bool
	var canAccessCourse bool
	if err := tx.QueryRowContext(ctx, `
		SELECT
		    s.assignment_id,
		    s.submitted_by,
		    s.status,
		    a.submissions_enabled
		        AND (a.submissions_close_at IS NULL OR a.submissions_close_at > NOW()),
		    EXISTS (
		        SELECT 1
		        FROM assigned_course_students acs
		        WHERE acs.portal_user_id = $2
		        AND acs.course_id = a.course_id
		    )
		FROM assignment_submissions s
		JOIN assignments a
		ON s.assignment_id = a.id
		WHERE s.id = $1
		FOR UPDATE OF s
	`, submissionID, userID).Scan(&assignmentID, &submittedBy, &status, &submissionsOpen, &canAccessCourse); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if submittedBy != userID || !canAccessCourse {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Finishing an already submitted submission is idempotent.
	if status == SubmissionStatusSubmitted {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if status != SubmissionStatusPending || !submissionsOpen {
		w.WriteHeader(http.StatusConflict)
		return
	}

	// Delete any returned submissions for this assignment by this student
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM assignment_submissions
		WHERE assignment_id = $1
		AND submitted_by = $2
		AND status = $3
	`, assignmentID, userID, SubmissionStatusReturned); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var hasIncompleteAttachments bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1
		    FROM submission_attachments sa
		    JOIN portal_storage_objects so
		    ON so.id = sa.storage_object_id
		    WHERE sa.submission_id = $1
		    AND so.status = $2
		)
	`, submissionID, helpers2.StatusPending).Scan(&hasIncompleteAttachments); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hasIncompleteAttachments {
		w.WriteHeader(http.StatusConflict)
		return
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE assignment_submissions
		SET status = $1,
		created_at = NOW()
		WHERE id = $2
		AND status = $3
		AND submitted_by = $4
	`,
		SubmissionStatusSubmitted,
		submissionID,
		SubmissionStatusPending,
		userID,
	); err != nil {
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
