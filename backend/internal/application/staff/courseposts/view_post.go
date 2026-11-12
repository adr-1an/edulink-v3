package courseposts

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type ViewPostUser struct {
	ID    string
	Name  string
	Email string
}

type ViewPostAttachment struct {
	ID           string
	PresignedURL string
	Filename     string
	ContentType  string
}

type ViewPostPost struct {
	Attachments []ViewPostAttachment
	Author      ViewPostUser
	ID          string
	Title       string
	Body        string
	Color       string
	ShowUntil   *time.Time
	CreatedAt   time.Time
	EditedAt    *time.Time
}

type ViewPostInput struct {
	UserID int64
	PostID int64
}

type ViewPostResult struct {
	Access schools.AccessData
	Posts  []ViewPostPost
}

func (svc *Service) ViewPost(ctx context.Context, i *ViewPostInput) (*ViewPostResult, error) {
	userID := i.UserID
	postID := i.PostID

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

	if !schools.Can(schools.PermissionCoursePostView, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	var posts []ViewPostPost

	rows, err := svc.DB.QueryContext(ctx, `
		SELECT
		    so.id, so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type,
		    author.id, author.name, author.email,
		    
		    p.id, p.title, p.body, COALESCE(p.accent_color, ''), p.show_until, p.created_at, p.edited_at
		FROM course_posts p
		JOIN users author
		ON p.author_id = author.id
		LEFT JOIN post_attachments pa
		ON p.id = pa.post_id
		LEFT JOIN storage_objects so
		ON pa.storage_object_id = so.id
		AND so.status = $2
		WHERE p.id = $1
	`, postID, helpers.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var res *ViewPostPost

	for rows.Next() {
		var (
			p            ViewPostPost
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

			&p.Author.ID,
			&p.Author.Name,
			&p.Author.Email,

			&p.ID,
			&p.Title,
			&p.Body,
			&p.Color,
			&p.ShowUntil,
			&p.CreatedAt,
			&p.EditedAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if res == nil {
			res = &p
		}

		if aID != nil {
			// generate presigned url
			url, err := svc.S3.PresignedGetObject(ctx, *aBucketName, *aKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			res.Attachments = append(res.Attachments, ViewPostAttachment{
				ID:           strconv.FormatInt(*aID, 10),
				PresignedURL: url.String(),
				Filename:     *aFilename,
				ContentType:  *aContentType,
			})
		}
	}

	if res != nil {
		posts = append(posts, *res)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := schools.GetAllUserPermissions(ctx, svc.DB, userID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ViewPostResult{Access: access, Posts: posts}, nil
}
