package students

import (
	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
)

type UploadProfilePictureFile struct {
	FileName            string
	DeclaredSize        int64
	DeclaredContentType string
}

type UploadProfilePictureInput struct {
	UserID       int64
	PortalUserID int64
	File         UploadProfilePictureFile
}

type UploadProfilePictureResult struct {
	ID              int64
	CompletionToken string
	URL             string
}

func (s *Service) UploadStudentProfilePicture(ctx context.Context, i *UploadProfilePictureInput) (*UploadProfilePictureResult, error) {
	// Get the portal user's school ID
	var portalUserSchoolID int64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id FROM portal_users WHERE id = $1
	`, i.PortalUserID).Scan(&portalUserSchoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	// Check perms
	if !schools.Can(schools.PermissionStudentUpdate, i.UserID, portalUserSchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Parse & validate
	if i.File.FileName == "" || len(i.File.FileName) > 255 || i.File.DeclaredSize == 0 || i.File.DeclaredContentType == "" {
		return nil, ErrInvalidData
	}

	if i.File.DeclaredSize <= 0 || i.File.DeclaredSize > 5*1024*1024 /* 5MB */ {
		return nil, ErrInvalidData
	}

	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
		"image/gif":  true,
	}

	if !allowed[i.File.DeclaredContentType] {
		return nil, ErrInvalidData
	}

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("Missing S3_BUCKET environment variable.")
		return nil, errors.New("missing s3_bucket env var")
	}

	objectID, err := s.Sf.NextID()
	if err != nil {
		return nil, err
	}

	pfpID, err := s.Sf.NextID()
	if err != nil {
		return nil, err
	}

	completionToken, err := gonanoid.New(128)
	if err != nil {
		return nil, err
	}
	tokenHash := helpers.MakeHash256(completionToken)

	objKey := fmt.Sprintf("%s/%d", helpers.UploadCategoryStudentProfilePics, objectID)

	url, err := s.S3.PresignedPutObject(ctx, bucketName, objKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	uploadService := helpers.UploadService{
		Db:         tx,
		S3:         s.S3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	if err := uploadService.StoreStorageObjectsRow(
		objectID,
		tokenHash,
		i.UserID,
		i.File.FileName,
		i.File.DeclaredSize,
		i.File.DeclaredContentType,
		objKey,
	); err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO portal_user_profile_pictures (id, portal_user_id, storage_object_id)
		VALUES ($1, $2, $3)
	`, pfpID, i.PortalUserID, objectID); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) &&
			pqErr.Code == "23505" &&
			pqErr.Constraint == "portal_user_profile_pictures_portal_user_id_unique" {
			return nil, ErrProfilePictureExists
		}
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UploadProfilePictureResult{
		ID:              objectID,
		CompletionToken: completionToken,
		URL:             url.String(),
	}, nil
}

type ClearStudentProfilePictureInput struct {
	UserID       int64
	PortalUserID int64
}

func (s *Service) ClearStudentProfilePicture(ctx context.Context, i *ClearStudentProfilePictureInput) error {
	// Get the student's school ID
	var studentSchoolID int64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id FROM portal_users WHERE id = $1
	`, i.PortalUserID).Scan(&studentSchoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		return err
	}

	// Check permissions
	if !schools.Can(schools.PermissionStudentUpdate, i.UserID, studentSchoolID, ctx, s.DB) {
		return ErrForbidden
	}

	// Delete the profile picture and its storage record.
	var objKey string
	var bucketName string
	if err := s.DB.QueryRowContext(ctx, `
		WITH deleted_picture AS (
			DELETE FROM portal_user_profile_pictures
			WHERE portal_user_id = $1
			RETURNING storage_object_id
		)
		DELETE FROM storage_objects so
		USING deleted_picture p
		WHERE so.id = p.storage_object_id
		RETURNING so.object_key, so.bucket_name
	`, i.PortalUserID).Scan(&objKey, &bucketName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		log.Println(err)
		return err
	}

	// Delete from S3
	if err := s.S3.RemoveObject(ctx, bucketName, objKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
		// No return on purpose
	}

	return nil
}
