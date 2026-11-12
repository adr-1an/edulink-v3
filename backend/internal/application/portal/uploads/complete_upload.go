package uploads

import (
	helpers2 "app/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
)

type CompleteUploadInput struct {
	UserID          int64
	ObjectID        int64
	InvalidPayload  bool
	CompletionToken string
}

func (s *Service) CompleteUpload(ctx context.Context, i *CompleteUploadInput) error {
	doCleanup := true
	userID := i.UserID
	objectID := i.ObjectID

	if i.InvalidPayload {
		return ErrBadRequest
	}

	tokenHash := helpers2.MakeHash256(i.CompletionToken)

	var objectUserID int64
	var declaredSize int64
	var declaredContentType string
	var objectKey string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT uploaded_by, declared_file_size, declared_content_type, object_key
		FROM portal_storage_objects WHERE id = $1 AND completion_token = $2 AND status = $3
	`, objectID, tokenHash, helpers2.StatusPending).Scan(&objectUserID, &declaredSize, &declaredContentType, &objectKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if userID != objectUserID {
		return ErrForbidden
	}

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		return ErrInternal
	}

	defer func() {
		if !doCleanup {
			return
		}

		if err := s.S3.RemoveObject(context.WithoutCancel(ctx), bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
			log.Println(err)
		}

		if _, err := s.DB.ExecContext(context.WithoutCancel(ctx), `
			UPDATE portal_storage_objects
			SET status = $1
			WHERE id = $2
			AND completion_token = $3
			AND status = $4
			AND uploaded_by = $5
		`, helpers2.StatusFailed, objectID, tokenHash, helpers2.StatusPending, userID); err != nil {
			log.Println(err)
		}
	}()

	// check if the declared object data matches the actual data
	info, err := s.S3.StatObject(ctx, bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	if info.Size != declaredSize {
		return ErrUnprocessableEntity
	}

	if info.ContentType != declaredContentType {
		return ErrUnprocessableEntity
	}
	// note: the content type stored in s3 is from the client. there's no guarantee it's correct, but
	// for this handler it really doesn't matter.

	if _, err := s.DB.ExecContext(ctx, `
		UPDATE portal_storage_objects
		SET status = $1, completion_token = '', completed_at = NOW()
		WHERE id = $2
		AND completion_token = $3
		AND status = $4
		AND uploaded_by = $5
	`, helpers2.StatusDone, objectID, tokenHash, helpers2.StatusPending, objectUserID); err != nil {
		log.Println(err)
		return err
	}

	doCleanup = false

	return nil
}
