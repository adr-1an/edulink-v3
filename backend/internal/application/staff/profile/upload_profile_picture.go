// Package profile
package profile

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"app/internal/cryptoutil"
	"app/internal/uploads"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type UploadProfilePictureInput struct {
	UserID              int64
	Name                string
	DeclaredSize        int64
	DeclaredContentType string
}
type UploadProfilePictureResult struct {
	ID              int64
	CompletionToken string
	URL             string
}

func (s *Service) UploadProfilePicture(ctx context.Context, i *UploadProfilePictureInput) (*UploadProfilePictureResult, error) {
	userID := i.UserID

	if i.DeclaredSize <= 0 || i.DeclaredSize > 5*1024*1024 /* 5MB */ {
		return nil, ErrInvalidData
	}

	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
		"image/gif":  true,
	}

	if !allowed[i.DeclaredContentType] {
		return nil, ErrInvalidData
	}

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		return nil, fmt.Errorf("S3_BUCKET environment variable not set")
	}

	upload := uploads.UploadService{
		DB:         s.DB,
		S3:         s.S3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pfpID, err := s.Sf.NextID()
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

	objKey := fmt.Sprintf("%s/%d", uploads.UploadCategoryUserProfilePics, id)

	if err := upload.StoreStorageObjectsRow(id, tokenHash, userID, i.Name, i.DeclaredSize, i.DeclaredContentType, objKey); err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err := s.DB.ExecContext(ctx, `
		INSERT INTO user_profile_pictures (id, user_id, storage_object_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET storage_object_id = EXCLUDED.storage_object_id
	`, pfpID, userID, id); err != nil {
		log.Println(err)
		return nil, err
	}

	url, err := s.S3.PresignedPutObject(ctx, bucketName, objKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UploadProfilePictureResult{ID: id, CompletionToken: completionToken, URL: url.String()}, nil
}
