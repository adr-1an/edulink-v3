package submissions

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/permissions"
	"app/internal/submissions"

	"github.com/minio/minio-go/v7"
)

type deleteReturnedSubmissionStorageObject struct {
	BucketName string
	ObjectKey  string
}

type DeleteReturnedSubmissionInput struct {
	UserID       int64
	SubmissionID int64
}

func (svc *Service) DeleteReturnedSubmission(ctx context.Context, i *DeleteReturnedSubmissionInput) error {
	userID := i.UserID
	subID := i.SubmissionID

	schoolID, err := SubToSchoolID(subID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !permissions.Can(permissions.PermissionSubmissionDelete, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	tx, err := svc.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status submissions.SubmissionStatus
	if err := tx.QueryRowContext(ctx, `
		SELECT status
		FROM assignment_submissions
		WHERE id = $1
		FOR UPDATE
	`, subID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if status != submissions.SubmissionStatusReturned {
		return ErrConflict
	}

	var storageObjects []deleteReturnedSubmissionStorageObject

	rows, err := tx.QueryContext(ctx, `
		DELETE FROM portal_storage_objects so
		USING submission_attachments sa
		WHERE sa.storage_object_id = so.id
		AND sa.submission_id = $1
		RETURNING so.bucket_name, so.object_key
	`, subID)
	if err != nil {
		log.Println(err)
		return err
	}

	for rows.Next() {
		var object deleteReturnedSubmissionStorageObject
		if err := rows.Scan(&object.BucketName, &object.ObjectKey); err != nil {
			_ = rows.Close()
			log.Println(err)
			return err
		}
		storageObjects = append(storageObjects, object)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		log.Println(err)
		return err
	}
	if err := rows.Close(); err != nil {
		log.Println(err)
		return err
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM assignment_submissions
		WHERE id = $1
		AND status = $2
	`, subID, submissions.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return err
	}
	if rowsAffected != 1 {
		return ErrConflict
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	for _, object := range storageObjects {
		if err := svc.S3.RemoveObject(context.WithoutCancel(ctx), object.BucketName, object.ObjectKey, minio.RemoveObjectOptions{}); err != nil {
			log.Println(err)
		}
	}

	return nil
}
