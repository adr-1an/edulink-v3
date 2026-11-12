package attachments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"app/internal/cryptoutil"
	"app/internal/permissions"
	"app/internal/uploads"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type InitPostAttachmentUploadData struct {
	Name                string
	DeclaredSize        int64
	DeclaredContentType string
}

type InitPostAttachmentUploadInput struct {
	InvalidPayload bool
	UserID         int64
	PostID         int64
	Data           InitPostAttachmentUploadData
}

type InitPostAttachmentUploadResult struct {
	ID              int64
	CompletionToken string
	URL             string
}

func (svc *Service) InitPostAttachmentUpload(ctx context.Context, i *InitPostAttachmentUploadInput) (*InitPostAttachmentUploadResult, error) {
	userID := i.UserID
	postID := i.PostID
	p := i.Data

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM course_posts p
		JOIN courses c
		ON p.course_id = c.id
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE p.id = $1
	`, postID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !permissions.Can(permissions.PermissionPostAttachmentCreate, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	// validate
	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	if p.DeclaredSize <= 0 || p.DeclaredSize > 5*1024*1024 { // 5mb
		return nil, ErrUnprocessableEntity
	}
	if p.Name == "" || len(p.Name) > 255 {
		return nil, ErrUnprocessableEntity
	}

	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/zip": true,
		"application/pdf": true,
	}

	if !allowed[p.DeclaredContentType] {
		return nil, ErrUnprocessableEntity
	}

	// generate id & completion token
	id, err := svc.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	postAttachmentID, err := svc.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tokenHash := cryptoutil.MakeHash256(token)

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		return nil, ErrInternal
	}

	// store row
	s := uploads.UploadService{
		Db:         svc.DB,
		S3:         svc.S3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	objectKey := fmt.Sprintf("%s/%d", uploads.UploadCategoryPostAttachments, id)

	if err := s.StoreStorageObjectsRow(id, tokenHash, userID, p.Name, p.DeclaredSize, p.DeclaredContentType, objectKey); err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err := svc.DB.ExecContext(ctx, `
		INSERT INTO post_attachments (id, post_id, storage_object_id)
		VALUES ($1, $2, $3)
	`, postAttachmentID, postID, id); err != nil {
		log.Println(err)
		return nil, err
	}

	// generate presigned url
	url, err := svc.S3.PresignedPutObject(ctx, bucketName, objectKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &InitPostAttachmentUploadResult{ID: id, CompletionToken: token, URL: url.String()}, nil
}
