package students

import (
	"app/internal/helpers"
	"context"
	"log"
	"strconv"
	"time"
)

type ListAllAssignmentsInput struct{ UserID int64 }
type ListAllAssignmentsResult struct {
	Assignments []ListAllAssignmentsAssignment
}

func (s *Service) ListAllAssignments(ctx context.Context, i *ListAllAssignmentsInput) (*ListAllAssignmentsResult, error) {
	userID := i.UserID

	// get all assignments that belong to a ListAllAssignmentsCourse the student is assigned to

	var assignments []ListAllAssignmentsAssignment

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		    p.id,
		    p.title,
		    p.body,
		    p.accent_color,
		    
		    c.id,
		    c.name,
		    
		    a.id,
		    a.title,
		    COALESCE(a.description, ''),
		    a.due_date,
		    a.submissions_enabled,
		    a.submissions_close_at,
		    a.created_at,

		    s.id,
		    s.status,
		    s.notes,
		    s.created_at,
		    
		    ssc.score_percentage,
		    ssc.notes,
		    ssc.graded_at,
		    
		    gu.id,
		    gu.name,
		    gu.email,
		    gu.phone
		FROM assignments a
		LEFT JOIN course_posts p
			ON a.referenced_post_id = p.id
		JOIN courses c
			ON a.course_id = c.id
		JOIN assigned_course_students acs
			ON acs.course_id = a.course_id
			AND acs.portal_user_id = $1
		JOIN grades g
		    ON c.grade_id = g.id
		JOIN academic_years y
			ON g.academic_year_id = y.id
		   	AND y.is_active = TRUE
		LEFT JOIN LATERAL (
		    SELECT sub.*
		    FROM assignment_submissions sub
		    WHERE sub.assignment_id = a.id
		    AND sub.submitted_by = $1
		    ORDER BY
		        (sub.status <> $2) DESC,
		        sub.created_at DESC
		    LIMIT 1
		    ) s ON TRUE
		LEFT JOIN submission_scores ssc
		    ON s.id = ssc.submission_id
		LEFT JOIN users gu
		    ON ssc.graded_by = gu.id
		WHERE a.archived_at IS NULL
	`, userID, helpers.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type optRefPost struct {
		ID    *int64
		Title *string
		Body  *string
		Color *string
	}

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
			a                   ListAllAssignmentsAssignment
			nullRefPost         optRefPost
			submissionID        *int64
			submissionStatus    *string
			submissionNotes     *string
			submissionCreatedAt *time.Time
			nullGrade           optGrade
		)

		if err := rows.Scan(
			&nullRefPost.ID,
			&nullRefPost.Title,
			&nullRefPost.Body,
			&nullRefPost.Color,

			&a.Course.ID,
			&a.Course.Name,

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

		if nullRefPost.ID != nil {
			a.RefPost = &ListAllAssignmentsRefPost{
				ID:    strconv.FormatInt(*nullRefPost.ID, 10),
				Title: *nullRefPost.Title,
				Body:  *nullRefPost.Body,
				Color: *nullRefPost.Color,
			}
		}

		if submissionID != nil {
			a.Submission = &AssignmentSubmission{
				ID:          strconv.FormatInt(*submissionID, 10),
				Status:      helpers.SubmissionStatus(*submissionStatus),
				Notes:       *submissionNotes,
				Attachments: []AssignmentSubmissionAttachment{},
				CreatedAt:   *submissionCreatedAt,
				Grade:       nil,
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

	return &ListAllAssignmentsResult{Assignments: assignments}, nil
}
