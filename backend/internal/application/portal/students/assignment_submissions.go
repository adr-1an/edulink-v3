package students

import (
	"app/internal/cryptoutil"
	"app/internal/permissions"
	"app/internal/submissions"
	"app/internal/uploads"

	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"log"
	"os"
	"strconv"
	"time"
)

func listAssignmentSubmissionAttachments(ctx context.Context, db *sql.DB, submissionIDs []int64, s3 *minio.Client) (map[string][]AssignmentSubmissionAttachment, error) {
	attachments := make(map[string][]AssignmentSubmissionAttachment)
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
	`, pq.Array(submissionIDs), uploads.StatusDone)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var submissionID int64
		var attachment AssignmentSubmissionAttachment

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

type DeleteSubmissionAttachmentInput struct {
	UserID       int64
	SubmissionID int64
	AttachmentID int64
}

func (s *Service) DeleteSubmissionAttachment(ctx context.Context, i *DeleteSubmissionAttachmentInput) error {
	userID := i.UserID
	submissionID := i.SubmissionID
	attachmentID := i.AttachmentID

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var submittedBy int64
	var status submissions.SubmissionStatus
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
			return ErrNotFound
		} else {
			log.Println(err)
			return err
		}
	}

	if submittedBy != userID || !canAccessCourse {
		return ErrForbidden
	}

	if status != submissions.SubmissionStatusPending {
		return ErrConflict
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM portal_storage_objects
		WHERE id = $1
	`, storageObjectID); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	if err := s.S3.RemoveObject(context.WithoutCancel(ctx), bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
	}

	return nil
}

type CreateAssignmentSubmissionInput struct {
	UserID         int64
	AssignmentID   int64
	InvalidPayload bool
	Notes          string
}
type CreateAssignmentSubmissionResult struct {
	SubmissionID string
	Status       submissions.SubmissionStatus
}

func (s *Service) CreateAssignmentSubmission(ctx context.Context, i *CreateAssignmentSubmissionInput) (*CreateAssignmentSubmissionResult, error) {
	userID := i.UserID
	assignmentID := i.AssignmentID

	var courseID int64
	var submissionsOpen bool
	var archived bool
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
		    course_id,
		    submissions_enabled
		        AND (submissions_close_at IS NULL OR submissions_close_at > NOW()),
		    archived_at IS NOT NULL
		FROM assignments
		WHERE id = $1
	`, assignmentID).Scan(&courseID, &submissionsOpen, &archived); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !permissions.CanAccessCourse(s.DB, userID, courseID, ctx) {
		return nil, ErrForbidden
	}

	if !submissionsOpen {
		return nil, ErrConflict
	}

	if archived {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	if len(i.Notes) > 2048 {
		return nil, ErrUnprocessableEntity
	}

	newSubID, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Store a new pending submission, or resume the student's existing pending submission.
	var submissionID int64
	if err := s.DB.QueryRowContext(ctx, `
		INSERT INTO assignment_submissions (id, assignment_id, submitted_by, notes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (assignment_id, submitted_by)
		WHERE status <> $6
		DO UPDATE
		SET notes = EXCLUDED.notes
		WHERE assignment_submissions.status = $5
		RETURNING id
	`, newSubID, assignmentID, userID, i.Notes, submissions.SubmissionStatusPending, submissions.SubmissionStatusReturned).Scan(&submissionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrConflict
		} else {
			log.Println(err)
			return nil, err
		}
	}

	return &CreateAssignmentSubmissionResult{SubmissionID: strconv.FormatInt(submissionID, 10), Status: submissions.SubmissionStatusPending}, nil
}

type InitSubmissionAttachmentUploadInput struct {
	UserID              int64
	SubmissionID        int64
	InvalidPayload      bool
	Name                string
	DeclaredSize        int64
	DeclaredContentType string
}
type InitSubmissionAttachmentUploadResult struct {
	ID              string
	AttachmentID    string
	CompletionToken string
	URL             string
}

func (s *Service) InitSubmissionAttachmentUpload(ctx context.Context, i *InitSubmissionAttachmentUploadInput) (*InitSubmissionAttachmentUploadResult, error) {
	userID := i.UserID
	submissionID := i.SubmissionID

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	if i.DeclaredSize <= 0 || i.DeclaredSize > 50*1024*1024 /* 50mb */ {
		return nil, ErrUnprocessableEntity
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

	if !allowed[i.DeclaredContentType] {
		return nil, ErrUnprocessableEntity
	}

	// generate id & completion token
	storageObjID, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	subAttachmentID, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	completionToken, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tokenHash := cryptoutil.MakeHash256(completionToken)

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		return nil, ErrInternal
	}

	objectKey := fmt.Sprintf("%s/%d", uploads.UploadCategorySubmissionAttachments, storageObjID)

	// Generate the URL before opening the transaction so no database lock is held
	// while communicating with object storage.
	url, err := s.S3.PresignedPutObject(ctx, bucketName, objectKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var submittedBy int64
	var status submissions.SubmissionStatus
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
			return nil, ErrNotFound
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if submittedBy != userID || !canAccessCourse {
		return nil, ErrForbidden
	}

	if status != submissions.SubmissionStatusPending || !submissionsOpen {
		return nil, ErrConflict
	}

	upload := uploads.UploadService{
		Db:         tx,
		S3:         s.S3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	if err := upload.PortalStoreStorageObjectsRow(storageObjID, tokenHash, userID, i.Name, i.DeclaredSize, i.DeclaredContentType, objectKey); err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO submission_attachments (id, submission_id, storage_object_id)
		VALUES ($1, $2, $3)
	`, subAttachmentID, submissionID, storageObjID); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &InitSubmissionAttachmentUploadResult{ID: strconv.FormatInt(storageObjID, 10), AttachmentID: strconv.FormatInt(subAttachmentID, 10), CompletionToken: completionToken, URL: url.String()}, nil
}

type CompleteSubmissionCreationInput struct {
	UserID       int64
	SubmissionID int64
}

func (s *Service) CompleteSubmissionCreation(ctx context.Context, i *CompleteSubmissionCreationInput) error {
	userID := i.UserID
	submissionID := i.SubmissionID

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var assignmentID int64
	var submittedBy int64
	var status submissions.SubmissionStatus
	var submissionsOpen bool
	var canAccessCourse bool
	var assignmentArchived bool
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
		    ),
		    a.archived_at IS NOT NULL
		FROM assignment_submissions s
		JOIN assignments a
		ON s.assignment_id = a.id
		WHERE s.id = $1
		FOR UPDATE OF s
	`, submissionID, userID).
		Scan(
			&assignmentID,
			&submittedBy,
			&status,
			&submissionsOpen,
			&canAccessCourse,
			&assignmentArchived,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else {
			log.Println(err)
			return err
		}
	}

	if submittedBy != userID || !canAccessCourse {
		return ErrForbidden
	}

	// Finishing an already submitted submission is idempotent.
	if status == submissions.SubmissionStatusSubmitted {
		return nil
	}

	if status != submissions.SubmissionStatusPending || !submissionsOpen {
		return ErrConflict
	}

	if assignmentArchived {
		return ErrForbidden
	}

	// Delete any returned submissions for this assignment by this student
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM assignment_submissions
		WHERE assignment_id = $1
		AND submitted_by = $2
		AND status = $3
	`, assignmentID, userID, submissions.SubmissionStatusReturned); err != nil {
		log.Println(err)
		return err
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
	`, submissionID, uploads.StatusPending).Scan(&hasIncompleteAttachments); err != nil {
		log.Println(err)
		return err
	}

	if hasIncompleteAttachments {
		// Delete the unfinished attachment records
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM submission_attachments sa
			USING portal_storage_objects so
			WHERE so.status = $1
			AND sa.submission_id = $2
		`, uploads.StatusPending, submissionID); err != nil {
			log.Println(err)
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE assignment_submissions
		SET status = $1,
		created_at = NOW()
		WHERE id = $2
		AND status = $3
		AND submitted_by = $4
	`,
		submissions.SubmissionStatusSubmitted,
		submissionID,
		submissions.SubmissionStatusPending,
		userID,
	); err != nil {
		log.Println(err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
