package coursepostattachments

import (
	"context"
	"database/sql"
	"errors"
	"log"

	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff/schools"

	"github.com/minio/minio-go/v7"
)

type DeletePostAttachmentInput struct {
	UserID       int64
	AttachmentID int64
}

func (svc *Service) DeletePostAttachment(ctx context.Context, i *DeletePostAttachmentInput) error {
	userID := i.UserID
	attachmentID := i.AttachmentID

	var schoolID int64
	var bucketName string
	var objKey string
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id, so.bucket_name, so.object_key
		FROM post_attachments pa
		JOIN course_posts p
		ON pa.post_id = p.id
		JOIN courses c
		ON p.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		JOIN storage_objects so
		ON so.id = pa.storage_object_id
		WHERE so.status = $1
		AND pa.id = $2
	`, helpers2.StatusDone, attachmentID).Scan(&schoolID, &bucketName, &objKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}
	}

	if !schools.Can(schools.PermissionPostAttachmentDelete, userID, schoolID, ctx, svc.DB) {
		return ErrForbidden
	}

	// delete from db
	if _, err := svc.DB.ExecContext(ctx, `
		WITH oid AS (
		    DELETE FROM post_attachments WHERE id = $1 RETURNING storage_object_id
		)
		DELETE FROM storage_objects so USING oid WHERE id = oid.storage_object_id
	`, attachmentID); err != nil {
		log.Println(err)
		return err
	}

	// delete from s3
	if err := svc.S3.RemoveObject(ctx, bucketName, objKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
