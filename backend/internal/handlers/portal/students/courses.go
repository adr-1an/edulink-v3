package students

import (
	helpers2 "app/internal/helpers"
	portal2 "app/internal/helpers/portal"
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

func ListCoursesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	type grade struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type course struct {
		ID          string `json:"id"`
		Grade       grade  `json:"grade"`
		Name        string `json:"name"`
		Description string `json:"description"`
		AccentColor string `json:"accentColor"`
	}
	var courses []course

	rows, err := db.QueryContext(ctx, `
		SELECT
		    g.id,
		    g.name,
		    
		    c.id,
		    c.name,
		    c.description,
		    c.color
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN assigned_course_students acs
		ON acs.course_id = c.id
		JOIN academic_years y ON g.academic_year_id = y.id
		WHERE acs.portal_user_id = $1
		AND y.is_active = TRUE
	`, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var c course

		if err := rows.Scan(&c.Grade.ID, &c.Grade.Name, &c.ID, &c.Name, &c.Description, &c.AccentColor); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"courses": courses,
	})
}

func CourseDashboardHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := portal2.TokenToUID(w, r, db, ctx, helpers2.AccTypeStudent)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !portal2.CanAccessCourse(db, userID, courseID, ctx) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// get course posts
	type user struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Email  string  `json:"email"`
		PfpURL *string `json:"profilePictureURL"`
	}

	type attachment struct {
		ID           string `json:"id"`
		PresignedURL string `json:"presignedUrl"`
		Name         string `json:"fileName"`
		ContentType  string `json:"contentType"`
	}

	type post struct {
		ID          string       `json:"id"`
		Attachments []attachment `json:"attachments"`
		Author      user         `json:"author"`
		Title       string       `json:"title"`
		Body        string       `json:"body"`
		AccentColor string       `json:"accentColor"`
		CreatedAt   time.Time    `json:"createdAt"`
		EditedAt    *time.Time   `json:"editedAt"`
	}
	var posts []post

	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.id,
		    
		    u.id,
		    u.name,
		    u.email,
		    
		    p.title,
		    p.body,
		    p.accent_color,
		    p.created_at,
		    p.edited_at,
		    
		    pa.id, so.object_key, so.bucket_name, so.original_file_name, so.declared_content_type,
		    
		    pso.object_key,
		    pso.bucket_name,
		    pso.original_file_name,
		    pso.declared_content_type
		FROM course_posts p
		JOIN courses c ON p.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id AND is_active = TRUE
		JOIN users u ON p.author_id = u.id
		LEFT JOIN user_profile_pictures upp ON upp.user_id = u.id
		LEFT JOIN storage_objects pso ON pso.id = upp.storage_object_id AND pso.status = $2
		LEFT JOIN post_attachments pa ON p.id = pa.post_id
		LEFT JOIN storage_objects so ON pa.storage_object_id = so.id AND so.status = $2
		WHERE p.course_id = $1
		AND (p.show_until IS NULL OR p.show_until > NOW())
	`, courseID, helpers2.StatusDone)
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
			&p.ID,
			&p.Author.ID,
			&p.Author.Name,
			&p.Author.Email,

			&p.Title,
			&p.Body,
			&p.AccentColor,
			&p.CreatedAt,
			&p.EditedAt,

			&aID,
			&aKey,
			&aBucketName,
			&aFilename,
			&aContentType,

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
		if aID.Valid && aKey.Valid && aBucketName.Valid && aFilename.Valid {
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
					Name:         aFilename.String,
					ContentType:  aContentType.String,
				},
			)
		}

		// User pfp
		if pfpKey != nil && pfpBucketName != nil && pfpFilename != nil && pfpContentType != nil {
			url, err := s3.PresignedGetObject(ctx, *pfpBucketName, *pfpKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			pfpURL := url.String()
			posts[index].Author.PfpURL = &pfpURL
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get assignments
	type referencedPost struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	type assignment struct {
		ReferencedPost     *referencedPost       `json:"referencedPost"`
		Submission         *assignmentSubmission `json:"submission"`
		ID                 string                `json:"id"`
		Title              string                `json:"title"`
		Description        string                `json:"description"`
		DueDate            *time.Time            `json:"dueDate"`
		SubmissionsEnabled bool                  `json:"submissionsEnabled"`
		SubmissionsCloseAt *time.Time            `json:"submissionsCloseAt"`
		CreatedAt          time.Time             `json:"createdAt"`
	}
	var assignments []assignment

	rows, err = db.QueryContext(ctx, `
		SELECT
		    p.id, p.title,
		    
		    a.id, a.title, a.description, a.due_date, a.submissions_enabled, a.submissions_close_at, a.created_at,

		    s.id, s.status, s.notes, s.created_at
		FROM assignments a
		LEFT JOIN course_posts p
		ON a.referenced_post_id = p.id
		LEFT JOIN assignment_submissions s
		ON s.assignment_id = a.id
		AND s.submitted_by = $2
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id AND y.is_active = TRUE
		WHERE a.course_id = $1
	`, courseID, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	var submissionIDs []int64
	for rows.Next() {
		var (
			a                   assignment
			refPostID           sql.NullInt64
			refPostTitle        sql.NullString
			submissionID        sql.NullInt64
			submissionStatus    sql.NullString
			submissionNotes     sql.NullString
			submissionCreatedAt sql.NullTime
		)

		if err := rows.Scan(
			&refPostID,
			&refPostTitle,

			&a.ID,
			&a.Title,
			&a.Description,
			&a.DueDate,
			&a.SubmissionsEnabled,
			&a.SubmissionsCloseAt,
			&a.CreatedAt,

			&submissionID,
			&submissionStatus,
			&submissionNotes,
			&submissionCreatedAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if refPostID.Valid {
			a.ReferencedPost = &referencedPost{
				ID:    strconv.FormatInt(refPostID.Int64, 10),
				Title: refPostTitle.String,
			}
		}

		if submissionID.Valid {
			a.Submission = &assignmentSubmission{
				ID:          strconv.FormatInt(submissionID.Int64, 10),
				Status:      SubmissionStatus(submissionStatus.String),
				Notes:       submissionNotes.String,
				Attachments: []assignmentSubmissionAttachment{},
				CreatedAt:   submissionCreatedAt.Time,
			}
			submissionIDs = append(submissionIDs, submissionID.Int64)
		}

		assignments = append(assignments, a)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := rows.Close(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	attachments, err := listAssignmentSubmissionAttachments(ctx, db, submissionIDs, s3)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for index := range assignments {
		if assignments[index].Submission != nil {
			assignments[index].Submission.Attachments = attachments[assignments[index].Submission.ID]
			if assignments[index].Submission.Attachments == nil {
				assignments[index].Submission.Attachments = []assignmentSubmissionAttachment{}
			}
		}
	}

	// get current course data
	type grade struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}

	type course struct {
		Grade       grade  `json:"grade"`
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		AccentColor string `json:"accentColor"`
	}
	var c course

	if err := db.QueryRowContext(ctx, `
		SELECT
		    g.id,
		    g.name,
		    g.level,
		    
		    c.id,
		    c.name,
		    c.description,
		    c.color
		
		FROM courses c
		JOIN grades g
		ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id AND y.is_active = TRUE
		WHERE c.id = $1
	`, courseID).
		Scan(
			&c.Grade.ID, &c.Grade.Name, &c.Grade.Level,
			&c.ID, &c.Name, &c.Description, &c.AccentColor,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"course":      c,
		"assignments": assignments,
		"posts":       posts,
	})
}
