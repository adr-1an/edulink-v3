package profile

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/minio/minio-go/v7"
)

type ClearProfilePictureInput struct{ UserID int64 }

func (s *Service) ClearProfilePicture(ctx context.Context, i *ClearProfilePictureInput) error {
	userID := i.UserID
	var err error

	// delete from db
	var objKey string
	var bucketName string
	if err = s.DB.QueryRowContext(ctx, `
		WITH oid AS (
		    DELETE FROM user_profile_pictures WHERE user_id = $1 RETURNING storage_object_id
		)
		DELETE FROM storage_objects so USING oid WHERE id = oid.storage_object_id RETURNING so.object_key, so.bucket_name
	`, userID).Scan(&objKey, &bucketName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else {
			log.Println(err)
			return err
		}
	}

	// delete from s3
	if err := s.S3.RemoveObject(ctx, bucketName, objKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
