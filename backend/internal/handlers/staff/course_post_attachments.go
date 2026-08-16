package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
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
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func InitPostAttachmentUploadHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
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
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionPostAttachmentCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type payload struct {
		Name                string `json:"name"`
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

	// validate
	if p.DeclaredSize <= 0 || p.DeclaredSize > 5*1024*1024 { // 5mb
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if p.Name == "" || len(p.Name) > 255 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// generate id & completion token
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	postAttachmentID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// store row
	s := helpers2.UploadService{
		Db:         db,
		S3:         s3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	objectKey := fmt.Sprintf("%s/%d", helpers2.UploadCategoryPostAttachments, id)

	if err := s.StoreStorageObjectsRow(id, tokenHash, userID, p.Name, p.DeclaredSize, p.DeclaredContentType, objectKey); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO post_attachments (id, post_id, storage_object_id)
		VALUES ($1, $2, $3)
	`, postAttachmentID, postID, id); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// generate presigned url
	url, err := s3.PresignedPutObject(ctx, bucketName, objectKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":              strconv.FormatInt(id, 10),
		"completionToken": token,
		"url":             url.String(),
	})
}

func DeletePostAttachmentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "attachmentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	var bucketName string
	var objKey string
	if err := db.QueryRowContext(ctx, `
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
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionPostAttachmentDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// delete from db
	if _, err := db.ExecContext(ctx, `
		WITH oid AS (
		    DELETE FROM post_attachments WHERE id = $1 RETURNING storage_object_id
		)
		DELETE FROM storage_objects so USING oid WHERE id = oid.storage_object_id
	`, attachmentID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// delete from s3
	if err := s3.RemoveObject(ctx, bucketName, objKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
