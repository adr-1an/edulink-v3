package students

import (
	"app/internal/helpers"
	"app/internal/helpers/portal"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeStudent)
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

	rows, err := h.DB.QueryContext(ctx, `
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

func (h *Handler) CourseDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, h.DB, ctx, helpers.AccTypeStudent)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !portal.CanAccessCourse(h.DB, userID, courseID, ctx) {
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

	rows, err := h.DB.QueryContext(ctx, `
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
		if aID != nil && aKey != nil && aBucketName != nil && aFilename != nil {
			url, err := h.S3.PresignedGetObject(ctx, *aBucketName, *aKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			posts[index].Attachments = append(
				posts[index].Attachments,
				attachment{
					ID:           strconv.FormatInt(*aID, 10),
					PresignedURL: url.String(),
					Name:         *aFilename,
					ContentType:  *aContentType,
				},
			)
		}

		// User pfp
		if pfpKey != nil && pfpBucketName != nil && pfpFilename != nil && pfpContentType != nil {
			url, err := h.S3.PresignedGetObject(ctx, *pfpBucketName, *pfpKey, 15*time.Minute, nil)
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

	rows, err = h.DB.QueryContext(ctx, `
		SELECT
		    p.id, p.title,
		    
		    a.id, a.title, a.description, a.due_date, a.submissions_enabled, a.submissions_close_at, a.created_at,

		    s.id, s.status, s.notes, s.created_at,
		    
		    ssc.score_percentage, ssc.notes, ssc.graded_at,
		    
		    u.id,
		    u.name,
		    u.email,
		    u.phone
		FROM assignments a
		LEFT JOIN course_posts p
		ON a.referenced_post_id = p.id
		LEFT JOIN LATERAL (
		    SELECT sub.*
		    FROM assignment_submissions sub
		    WHERE sub.assignment_id = a.id
		    AND sub.submitted_by = $2
		    ORDER BY (sub.status <> $3) DESC,
		 	sub.created_at DESC
		    LIMIT 1
		) s ON TRUE
		LEFT JOIN submission_scores ssc
		ON s.id = ssc.submission_id
		LEFT JOIN users u
		ON ssc.graded_by = u.id
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id AND y.is_active = TRUE
		WHERE a.course_id = $1
		AND a.archived_at IS NULL
	`, courseID, userID, helpers.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	type optUser struct {
		ID    *string
		Name  *string
		Email *string
		Phone *string
	}

	type optGrade struct {
		Score    *int
		Notes    *string
		GradedAt *time.Time
		GradedBy optUser
	}

	var submissionIDs []int64
	for rows.Next() {
		var (
			a                   assignment
			refPostID           *int64
			refPostTitle        *string
			submissionID        *int64
			submissionStatus    *string
			submissionNotes     *string
			submissionCreatedAt sql.NullTime
			nullGrade           optGrade
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

			&nullGrade.Score,
			&nullGrade.Notes,
			&nullGrade.GradedAt,

			&nullGrade.GradedBy.ID,
			&nullGrade.GradedBy.Name,
			&nullGrade.GradedBy.Email,
			&nullGrade.GradedBy.Phone,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if refPostID != nil {
			a.ReferencedPost = &referencedPost{
				ID:    strconv.FormatInt(*refPostID, 10),
				Title: *refPostTitle,
			}
		}

		if submissionID != nil {
			a.Submission = &assignmentSubmission{
				ID:          strconv.FormatInt(*submissionID, 10),
				Status:      helpers.SubmissionStatus(*submissionStatus),
				Grade:       nil,
				Notes:       *submissionNotes,
				Attachments: []assignmentSubmissionAttachment{},
				CreatedAt:   submissionCreatedAt.Time,
			}

			if nullGrade.Score != nil {
				a.Submission.Grade = &grade{
					Score:    *nullGrade.Score,
					Notes:    nullGrade.Notes,
					GradedAt: *nullGrade.GradedAt,
					GradedBy: AssignmentGradeUser{
						ID:    *nullGrade.GradedBy.ID,
						Name:  *nullGrade.GradedBy.Name,
						Email: *nullGrade.GradedBy.Email,
						Phone: nullGrade.GradedBy.Phone,
					},
				}
			}

			submissionIDs = append(submissionIDs, *submissionID)
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

	attachments, err := listAssignmentSubmissionAttachments(ctx, h.DB, submissionIDs, h.S3)
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

	if err := h.DB.QueryRowContext(ctx, `
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
