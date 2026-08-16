package students

import (
	"app/internal/helpers"
	"app/internal/helpers/portal"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
)

func ListAllAssignmentsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, db, ctx, helpers.AccTypeStudent)
	if err != nil {
		return
	}

	// get all assignments that belong to a course the student is assigned to
	type refPost struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Body  string `json:"body"`
		Color string `json:"accentColor"`
	}

	type course struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type assignment struct {
		RefPost            *refPost              `json:"referencedPost"`
		Course             course                `json:"course"`
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

	rows, err := db.QueryContext(ctx, `
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
	`, userID, SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	type optRefPost struct {
		ID    sql.NullInt64
		Title sql.NullString
		Body  sql.NullString
		Color sql.NullString
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
			a                   assignment
			nullRefPost         optRefPost
			submissionID        sql.NullInt64
			submissionStatus    sql.NullString
			submissionNotes     sql.NullString
			submissionCreatedAt sql.NullTime
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if nullRefPost.ID.Valid {
			a.RefPost = &refPost{
				ID:    strconv.FormatInt(nullRefPost.ID.Int64, 10),
				Title: nullRefPost.Title.String,
				Body:  nullRefPost.Body.String,
				Color: nullRefPost.Color.String,
			}
		}

		if submissionID.Valid {
			a.Submission = &assignmentSubmission{
				ID:          strconv.FormatInt(submissionID.Int64, 10),
				Status:      SubmissionStatus(submissionStatus.String),
				Notes:       submissionNotes.String,
				Attachments: []assignmentSubmissionAttachment{},
				CreatedAt:   submissionCreatedAt.Time,
				Grade:       nil,
			}

			if nullGrade.Score != nil {
				a.Submission.Grade = &grade{
					Score:    *nullGrade.Score,
					Notes:    nullGrade.Notes,
					GradedAt: *nullGrade.GradedAt,
					GradedBy: user{
						ID:    *nullGrade.GradedBy.ID,
						Name:  *nullGrade.GradedBy.Name,
						Email: *nullGrade.GradedBy.Email,
						Phone: nullGrade.GradedBy.Phone,
					},
				}
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

	_ = json.NewEncoder(w).Encode(map[string]any{
		"assignments": assignments,
	})
}
