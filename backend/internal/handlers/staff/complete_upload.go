package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func CompleteUploadHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	doCleanup := true

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	objectID, err := strconv.ParseInt(chi.URLParam(r, "objectID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type payload struct {
		CompletionToken string `json:"completionToken"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenHash := helpers2.MakeHash256(p.CompletionToken)

	var objectUserID int64
	var declaredSize int64
	var declaredContentType string
	var objectKey string
	if err := db.QueryRowContext(ctx, `
		SELECT uploaded_by, declared_file_size, declared_content_type, object_key
		FROM storage_objects WHERE id = $1 AND completion_token = $2 AND status = $3
	`, objectID, tokenHash, helpers2.StatusPending).Scan(&objectUserID, &declaredSize, &declaredContentType, &objectKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if userID != objectUserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		if !doCleanup {
			return
		}

		if err := s3.RemoveObject(context.WithoutCancel(ctx), bucketName, objectKey, minio.RemoveObjectOptions{}); err != nil {
			log.Println(err)
		}

		if _, err := db.ExecContext(context.WithoutCancel(ctx), `
			UPDATE storage_objects
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
	info, err := s3.StatObject(ctx, bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if info.Size != declaredSize {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if info.ContentType != declaredContentType {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	// note: the content type stored in s3 is from the client. there's no guarantee it's correct, but
	// for this handler it really doesn't matter.

	if _, err := db.ExecContext(ctx, `
		UPDATE storage_objects
		SET status = $1, completion_token = '', completed_at = NOW()
		WHERE id = $2
		AND completion_token = $3
		AND status = $4
		AND uploaded_by = $5
	`, helpers2.StatusDone, objectID, tokenHash, helpers2.StatusPending, objectUserID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	doCleanup = false

	w.WriteHeader(http.StatusNoContent)
}
