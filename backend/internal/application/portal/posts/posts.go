package posts

import (
	helpers2 "app/internal/helpers"
	portal2 "app/internal/helpers/portal"
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
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

type ViewPostInput struct {
	UserID int64
	PostID int64
}
type ViewPostResult struct{ Post ViewPostPost }

func (s *Service) ViewPost(ctx context.Context, i *ViewPostInput) (*ViewPostResult, error) {
	userID := i.UserID
	postID := i.PostID

	courseID, err := postToCourseID(postID, s.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !portal2.CanAccessCourse(s.DB, userID, courseID, ctx) {
		return nil, ErrForbidden
	}

	var p ViewPostPost

	if err := s.DB.QueryRowContext(ctx, `
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
		return nil, err
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT pa.id, so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type
		FROM post_attachments pa
		JOIN storage_objects so
		ON pa.storage_object_id = so.id
		AND so.status = $2
		WHERE pa.post_id = $1
	`, postID, helpers2.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			aID          *int64
			aBucketName  *string
			aKey         *string
			aFilename    *string
			aContentType *string
		)

		if err := rows.Scan(
			&aID,
			&aBucketName,
			&aKey,
			&aFilename,
			&aContentType,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if aID != nil {
			url, err := s.S3.PresignedGetObject(ctx, *aBucketName, *aKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			p.Attachments = append(p.Attachments, ViewPostAttachment{
				ID:           strconv.FormatInt(*aID, 10),
				Filename:     *aFilename,
				ContentType:  *aContentType,
				PresignedURL: url.String(),
			})
		}
	}

	return &ViewPostResult{Post: p}, nil
}
