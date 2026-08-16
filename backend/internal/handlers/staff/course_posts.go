package staff

import (
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

type CoursePostPayload struct {
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	AccentColor string       `json:"accentColor"`
	ShowUntil   sql.NullTime `json:"showUntil"`
}

func parseAndValidate(p CoursePostPayload) (CoursePostPayload, error) {
	// Parse
	p.Title = strings.TrimSpace(p.Title)
	p.Body = strings.TrimSpace(p.Body)
	p.AccentColor = strings.TrimSpace(strings.TrimPrefix(p.AccentColor, "#"))

	// Validate
	if p.Title == "" || p.Body == "" || p.AccentColor == "" {
		return CoursePostPayload{}, errors.New("missing fields")
	}

	if len(p.Title) > 32 || len(p.Body) > 2048 || len(p.AccentColor) != 6 {
		return CoursePostPayload{}, errors.New("invalid lengths")
	}

	return p, nil
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON G.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionCoursePostCreate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p CoursePostPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err = parseAndValidate(p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO course_posts (id, course_id, author_id, title, body, accent_color, show_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, courseID, userID, p.Title, p.Body, p.AccentColor, p.ShowUntil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionPostCreate, schools.TypeCreate, "Post created", "{user} created the post '"+p.Title+"'.", tx, ctx, sf, "{user} created post '"+p.Title+"' with accent color '#"+p.AccentColor+"'."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": strconv.FormatInt(id, 10),
	})
}

func UpdatePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	if !schools.Can(schools.PermissionCoursePostUpdate, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var p CoursePostPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err = parseAndValidate(p)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var oldTitle string
	var oldBody string
	var oldAccentColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT title, body, accent_color
		FROM course_posts
		WHERE id = $1
	`, postID).Scan(&oldTitle, &oldBody, &oldAccentColor); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE course_posts
		SET title = $1, body = $2, show_until = $3, accent_color = $4, edited_at = NOW()
		WHERE id = $5
	`, p.Title, p.Body, p.ShowUntil, p.AccentColor, postID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := "{user} updated the post title from '" + oldTitle + "' to '" + p.Title + "', changed accent color from '#" + oldAccentColor + "' to '#" + p.AccentColor + "', and changed the body length from " + strconv.Itoa(len(oldBody)) + " to " + strconv.Itoa(len(p.Body)) + " characters."
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionPostEdit, schools.TypeEdit, "Post updated", "{user} updated the post '"+p.Title+"'.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListPostsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	if err := db.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y
		ON g.academic_year_id = y.id
		WHERE c.id = $1
	`, courseID).Scan(&schoolID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !schools.Can(schools.PermissionCoursePostList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type attachment struct {
		ID           string `json:"id"`
		PresignedURL string `json:"presignedUrl"`
		Filename     string `json:"fileName"`
		ContentType  string `json:"contentType"`
	}

	type post struct {
		Attachments  []attachment `json:"attachments"`
		ID           string       `json:"id"`
		AuthorName   string       `json:"authorName"`
		AuthorPfpURL *string      `json:"authorProfilePictureURL"`

		Title       string       `json:"title"`
		Body        string       `json:"body"`
		ShowUntil   sql.NullTime `json:"showUntil"`
		AccentColor string       `json:"accentColor"`

		EditedAt  sql.NullTime `json:"editedAt"`
		CreatedAt time.Time    `json:"createdAt"`
	}
	var posts []post

	rows, err := db.QueryContext(ctx, `
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
	`, courseID, helpers.StatusDone)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	postIndexes := make(map[string]int)

	for rows.Next() {
		var (
			p            post
			aID          sql.NullInt64
			aKey         sql.NullString
			aBucketName  sql.NullString
			aFilename    sql.NullString
			aContentType sql.NullString

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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		index, exists := postIndexes[p.ID]
		if !exists {
			p.Attachments = []attachment{}
			posts = append(posts, p)
			index = len(posts) - 1
			postIndexes[p.ID] = index
		}

		// Post attachments
		if aID.Valid && aKey.Valid && aBucketName.Valid && aFilename.Valid && aContentType.Valid {
			// generate presigned url
			url, err := s3.PresignedGetObject(ctx, aBucketName.String, aKey.String, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			posts[index].Attachments = append(
				posts[index].Attachments,
				attachment{
					ID:           strconv.FormatInt(aID.Int64, 10),
					PresignedURL: url.String(),
					Filename:     aFilename.String,
					ContentType:  aContentType.String,
				},
			)
		}

		// Author pfp
		if pfpKey != nil && pfpBucketName != nil && pfpFilename != nil && pfpContentType != nil {
			url, err := s3.PresignedGetObject(ctx, *pfpBucketName, *pfpKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			pfpURL := url.String()
			posts[index].AuthorPfpURL = &pfpURL
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"posts":  posts,
	})
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
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

	if !schools.Can(schools.PermissionCoursePostDelete, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var postTitle string
	if err := tx.QueryRowContext(ctx, `
		SELECT title FROM course_posts WHERE id = $1
	`, postID).Scan(&postTitle); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM course_posts WHERE id = $1
	`, postID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionPostDelete, schools.TypeDelete, "Post deleted", "{user} deleted the post '"+postTitle+"'.", tx, ctx, sf, "{user} deleted post '"+postTitle+"' with ID "+strconv.FormatInt(postID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ViewPostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
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

	if !schools.Can(schools.PermissionCoursePostView, userID, schoolID, ctx, db) {
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
		Color       string       `json:"accentColor"`
		ShowUntil   *time.Time   `json:"showUntil"`
		CreatedAt   time.Time    `json:"createdAt"`
		EditedAt    *time.Time   `json:"editedAt"`
	}
	var posts []post

	rows, err := db.QueryContext(ctx, `
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	var res *post

	for rows.Next() {
		var (
			p            post
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if res == nil {
			res = &p
		}

		if aID.Valid {
			// generate presigned url
			url, err := s3.PresignedGetObject(ctx, aBucketName.String, aKey.String, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res.Attachments = append(res.Attachments, attachment{
				ID:           strconv.FormatInt(aID.Int64, 10),
				PresignedURL: url.String(),
				Filename:     aFilename.String,
				ContentType:  aContentType.String,
			})
		}
	}

	if res != nil {
		posts = append(posts, *res)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"posts":  posts,
	})
}
