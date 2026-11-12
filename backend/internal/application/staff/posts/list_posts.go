package posts

import (
	"context"
	"log"
	"strconv"
	"time"

	"app/internal/permissions"
	"app/internal/uploads"
)

type ListPostsAttachment struct {
	ID           string
	PresignedURL string
	Filename     string
	ContentType  string
}

type ListPostsPost struct {
	Attachments  []ListPostsAttachment
	ID           string
	AuthorName   string
	AuthorPfpURL *string

	Title       string
	Body        string
	ShowUntil   *time.Time
	AccentColor string

	EditedAt  *time.Time
	CreatedAt time.Time
}

type ListPostsInput struct {
	UserID   int64
	CourseID int64
}

type ListPostsResult struct {
	Access permissions.AccessData
	Posts  []ListPostsPost
}

func (svc *Service) ListPosts(ctx context.Context, i *ListPostsInput) (*ListPostsResult, error) {
	userID := i.UserID
	courseID := i.CourseID

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		log.Println(err)
		return nil, err
	}

	if !permissions.Can(permissions.PermissionCoursePostList, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	var posts []ListPostsPost

	rows, err := svc.DB.QueryContext(ctx, `
		SELECT
		    pa.id, so.object_key, so.bucket_name, so.original_file_name, so.declared_content_type,
		    p.id, u.name, p.title, p.body, p.show_until, p.accent_color, p.edited_at, p.created_at,
		    pso.object_key, pso.bucket_name, pso.original_file_name, pso.declared_content_type
		FROM course_posts p
		JOIN users u ON u.id = p.author_id
		LEFT JOIN post_attachments pa ON p.id = pa.post_id
		LEFT JOIN user_profile_pictures upp ON upp.user_id = u.id
		LEFT JOIN storage_objects pso ON pso.id = upp.storage_object_id AND pso.status = $2
		LEFT JOIN storage_objects so ON so.id = pa.storage_object_id AND so.status = $2
		WHERE p.course_id = $1
	`, courseID, uploads.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	postIndexes := make(map[string]int)

	for rows.Next() {
		var (
			p            ListPostsPost
			aID          *int64
			aKey         *string
			aBucketName  *string
			aFilename    *string
			aContentType *string

			pfpKey         *string
			pfpBucketName  *string
			pfpFilename    *string
			pfpContentType *string
		)

		if err := rows.Scan(
			&aID,
			&aKey,
			&aBucketName,
			&aFilename,
			&aContentType,
			&p.ID,
			&p.AuthorName,
			&p.Title,
			&p.Body,
			&p.ShowUntil,
			&p.AccentColor,
			&p.EditedAt,
			&p.CreatedAt,

			&pfpKey,
			&pfpBucketName,
			&pfpFilename,
			&pfpContentType,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		index, exists := postIndexes[p.ID]
		if !exists {
			p.Attachments = []ListPostsAttachment{}
			posts = append(posts, p)
			index = len(posts) - 1
			postIndexes[p.ID] = index
		}

		// Post attachments
		if aID != nil && aKey != nil && aBucketName != nil && aFilename != nil && aContentType != nil {
			// generate presigned url
			url, err := svc.S3.PresignedGetObject(ctx, *aBucketName, *aKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			posts[index].Attachments = append(
				posts[index].Attachments,
				ListPostsAttachment{
					ID:           strconv.FormatInt(*aID, 10),
					PresignedURL: url.String(),
					Filename:     *aFilename,
					ContentType:  *aContentType,
				},
			)
		}

		// Author pfp
		if pfpKey != nil && pfpBucketName != nil && pfpFilename != nil && pfpContentType != nil {
			url, err := svc.S3.PresignedGetObject(ctx, *pfpBucketName, *pfpKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			pfpURL := url.String()
			posts[index].AuthorPfpURL = &pfpURL
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := permissions.GetAllUserPermissions(ctx, svc.DB, userID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListPostsResult{Access: access, Posts: posts}, nil
}
