package portal

import (
	helpers2 "app/internal/helpers"
	portal2 "app/internal/helpers/portal"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func postToCourseID(postID int64, db *sql.DB, ctx context.Context) (int64, error) {
	var courseID int64
	if err := db.QueryRowContext(ctx, `	
		SELECT course_id FROM course_posts WHERE id = $1
	`, postID).Scan(&courseID); err != nil {
		return 0, err
	}

	return courseID, nil
}

func ViewPostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeEither)
	if err != nil {
		return
	}

	postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	courseID, err := postToCourseID(postID, db, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !portal2.CanAccessCourse(db, userID, courseID, ctx) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type attachment struct {
		ID           string `json:"id"`
		PresignedURL string `json:"presignedUrl"`
		Filename     string `json:"fileName"`
		ContentType  string `json:"contentType"`
	}

	type post struct {
		Attachments []attachment `json:"attachments"`
		Author      user         `json:"author"`
		ID          string       `json:"id"`
		Title       string       `json:"title"`
		Body        string       `json:"body"`
		AccentColor string       `json:"accentColor"`
		ShowUntil   *time.Time   `json:"showUntil"`
		CreatedAt   time.Time    `json:"createdAt"`
		EditedAt    *time.Time   `json:"editedAt"`
	}
	var p post

	if err := db.QueryRowContext(ctx, `
		SELECT
		    u.id, u.name, u.email,
		    p.id, p.title, p.body, p.accent_color, p.show_until, p.created_at, p.edited_at
		FROM course_posts p
		JOIN users u ON p.author_id = u.id
		WHERE p.id = $1
	`, postID).Scan(
		&p.Author.ID,
		&p.Author.Name,
		&p.Author.Email,
		&p.ID,
		&p.Title,
		&p.Body,
		&p.AccentColor,
		&p.ShowUntil,
		&p.CreatedAt,
		&p.EditedAt,
	); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT pa.id, so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type
		FROM post_attachments pa
		JOIN storage_objects so
		ON pa.storage_object_id = so.id
		AND so.status = $2
		WHERE pa.post_id = $1
	`, postID, helpers2.StatusDone)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			aID          sql.NullInt64
			aBucketName  sql.NullString
			aKey         sql.NullString
			aFilename    sql.NullString
			aContentType sql.NullString
		)

		if err := rows.Scan(
			&aID,
			&aBucketName,
			&aKey,
			&aFilename,
			&aContentType,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if aID.Valid {
			url, err := s3.PresignedGetObject(ctx, aBucketName.String, aKey.String, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			p.Attachments = append(p.Attachments, attachment{
				ID:           strconv.FormatInt(aID.Int64, 10),
				Filename:     aFilename.String,
				ContentType:  aContentType.String,
				PresignedURL: url.String(),
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"post": p,
	})
}
