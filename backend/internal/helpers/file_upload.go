package helpers

import (
	"context"
	"database/sql"

	"github.com/minio/minio-go/v7"
)

type UploadStatus string

const (
	StatusPending UploadStatus = "pending"
	StatusDone    UploadStatus = "done"
	StatusFailed  UploadStatus = "failed"
)

type databaseExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type UploadService struct {
	Db         databaseExecutor
	S3         *minio.Client
	Ctx        context.Context
	BucketName string
}

type UploadCategory string

const (
	UploadCategoryPostAttachments       UploadCategory = "post-attachments"
	UploadCategoryUserProfilePics       UploadCategory = "user-profile-pictures"
	UploadCategorySubmissionAttachments UploadCategory = "submission-attachments"
)

func (s UploadService) StoreStorageObjectsRow(
	id int64,
	completionToken []byte,
	uploadedBy int64,
	filename string,
	declaredSize int64,
	declaredContentType string,
	objectKey string,
) error {
	if _, err := s.Db.ExecContext(s.Ctx, `
		INSERT INTO storage_objects (
		                             id,
		                             completion_token,
		                             uploaded_by, bucket_name,
		                             object_key,
		                             original_file_name,
		                             declared_file_size,
		                             declared_content_type,
		                             status
		                             )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id,
		completionToken,
		uploadedBy,
		s.BucketName,
		objectKey,
		filename,
		declaredSize,
		declaredContentType,
		StatusPending,
	); err != nil {
		return err
	}

	return nil
}

func (s UploadService) PortalStoreStorageObjectsRow(
	id int64,
	completionToken []byte,
	uploadedBy int64,
	filename string,
	declaredSize int64,
	declaredContentType string,
	objectKey string,
) error {
	if _, err := s.Db.ExecContext(s.Ctx, `
		INSERT INTO portal_storage_objects (
		                                    id,
		                                    completion_token,
		                                    uploaded_by,
		                                    bucket_name,
		                                    object_key,
		                                    original_file_name,
		                                    declared_file_size,
		                                    declared_content_type,
		                                    status
		                                    )
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		id,
		completionToken,
		uploadedBy,
		s.BucketName,
		objectKey,
		filename,
		declaredSize,
		declaredContentType,
		StatusPending,
	); err != nil {
		return err
	}

	return nil
}
