package students

import (
	"app/internal/permissions"
	"app/internal/submissions"
	"app/internal/uploads"

	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"
)

type ListCoursesInput struct{ UserID int64 }
type ListCoursesResult struct{ Courses []ListCoursesCourse }

func (s *Service) ListCourses(ctx context.Context, i *ListCoursesInput) (*ListCoursesResult, error) {
	userID := i.UserID

	var courses []ListCoursesCourse

	rows, err := s.DB.QueryContext(ctx, `
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
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var c ListCoursesCourse

		if err := rows.Scan(&c.Grade.ID, &c.Grade.Name, &c.ID, &c.Name, &c.Description, &c.AccentColor); err != nil {
			log.Println(err)
			return nil, err
		}

		courses = append(courses, c)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListCoursesResult{Courses: courses}, nil
}

type CourseDashboardInput struct {
	UserID   int64
	CourseID int64
}
type CourseDashboardResult struct {
	Course      CourseDashboardCourse
	Assignments []CourseDashboardAssignment
	Posts       []CourseDashboardPost
}

func (s *Service) CourseDashboard(ctx context.Context, i *CourseDashboardInput) (*CourseDashboardResult, error) {
	userID := i.UserID
	courseID := i.CourseID

	if !permissions.CanAccessCourse(s.DB, userID, courseID, ctx) {
		return nil, ErrForbidden
	}

	// get CourseDashboardCourse posts

	var posts []CourseDashboardPost

	rows, err := s.DB.QueryContext(ctx, `
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
	`, courseID, uploads.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	postIndexes := make(map[string]int)

	for rows.Next() {
		var (
			p            CourseDashboardPost
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
			return nil, err
		}

		index, exists := postIndexes[p.ID]
		if !exists {
			p.Attachments = []CourseDashboardAttachment{}
			posts = append(posts, p)
			index = len(posts) - 1
			postIndexes[p.ID] = index
		}

		// Post attachments
		if aID != nil && aKey != nil && aBucketName != nil && aFilename != nil {
			url, err := s.S3.PresignedGetObject(ctx, *aBucketName, *aKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			posts[index].Attachments = append(
				posts[index].Attachments,
				CourseDashboardAttachment{
					ID:           strconv.FormatInt(*aID, 10),
					PresignedURL: url.String(),
					Name:         *aFilename,
					ContentType:  *aContentType,
				},
			)
		}

		// User pfp
		if pfpKey != nil && pfpBucketName != nil && pfpFilename != nil && pfpContentType != nil {
			url, err := s.S3.PresignedGetObject(ctx, *pfpBucketName, *pfpKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			pfpURL := url.String()
			posts[index].Author.PfpURL = &pfpURL
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	// get assignments

	var assignments []CourseDashboardAssignment

	rows, err = s.DB.QueryContext(ctx, `
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
	`, courseID, userID, submissions.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		return nil, err
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
			a                   CourseDashboardAssignment
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
			return nil, err
		}

		if refPostID != nil {
			a.ReferencedPost = &CourseDashboardReferencedPost{
				ID:    strconv.FormatInt(*refPostID, 10),
				Title: *refPostTitle,
			}
		}

		if submissionID != nil {
			a.Submission = &AssignmentSubmission{
				ID:          strconv.FormatInt(*submissionID, 10),
				Status:      submissions.SubmissionStatus(*submissionStatus),
				Grade:       nil,
				Notes:       *submissionNotes,
				Attachments: []AssignmentSubmissionAttachment{},
				CreatedAt:   submissionCreatedAt.Time,
			}

			if nullGrade.Score != nil {
				a.Submission.Grade = &SubmissionGrade{
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
		return nil, err
	}

	if err := rows.Close(); err != nil {
		log.Println(err)
		return nil, err
	}

	attachments, err := listAssignmentSubmissionAttachments(ctx, s.DB, submissionIDs, s.S3)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for index := range assignments {
		if assignments[index].Submission != nil {
			assignments[index].Submission.Attachments = attachments[assignments[index].Submission.ID]
			if assignments[index].Submission.Attachments == nil {
				assignments[index].Submission.Attachments = []AssignmentSubmissionAttachment{}
			}
		}
	}

	// get current CourseDashboardCourse data

	var c CourseDashboardCourse

	if err := s.DB.QueryRowContext(ctx, `
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
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	return &CourseDashboardResult{Course: c, Assignments: assignments, Posts: posts}, nil
}
