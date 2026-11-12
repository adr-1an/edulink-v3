package submissions

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"app/internal/permissions"
	"app/internal/portalusers"
	"app/internal/submissions"
)

type ListAssignmentSubmissionsPortalUser struct {
	ID       string
	Email    string
	Phone    string
	Name     string
	Lastname string
}

type ListAssignmentSubmissionsSubmission struct {
	ID          string
	Status      string
	SubmittedBy ListAssignmentSubmissionsPortalUser
	Notes       *string
	SubmittedAt time.Time
	Graded      bool
}

type listAssignmentSubmissionsSubmissionRow struct {
	SubmissionID string

	UserID   string
	Email    string
	Phone    string
	Name     string
	Lastname string

	Notes       *string
	SubmittedAt time.Time
	Status      string
	Graded      bool
}

type ListAssignmentSubmissionsInput struct {
	UserID       int64
	AssignmentID int64
}

type ListAssignmentSubmissionsResult struct {
	Access      permissions.AccessData
	Submissions []ListAssignmentSubmissionsSubmission
}

func (svc *Service) ListAssignmentSubmissions(ctx context.Context, i *ListAssignmentSubmissionsInput) (*ListAssignmentSubmissionsResult, error) {
	userID := i.UserID
	assignmentID := i.AssignmentID

	var schoolID int64
	if err := svc.DB.QueryRowContext(ctx, `
		SELECT y.school_id
		FROM assignments a
		JOIN courses c ON a.course_id = c.id
		JOIN grades g ON c.grade_id = g.id
		JOIN academic_years y ON g.academic_year_id = y.id
		WHERE  a.id = $1
	`, assignmentID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !permissions.Can(permissions.PermissionSubmissionList, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	var subs []ListAssignmentSubmissionsSubmission

	rows, err := svc.DB.QueryContext(ctx, `
		SELECT
		    s.id, u.id, u.email, u.phone, u.name, u.last_name,
		    s.notes, s.created_at, s.status,
		    EXISTS (SELECT 1 FROM submission_scores WHERE submission_id = s.id)
		FROM assignment_submissions s
		JOIN portal_users u ON s.submitted_by = u.id
		WHERE (s.status = $1
		  OR s.status = $4)
		AND s.assignment_id = $2
		AND u.account_type = $3
	`, submissions.SubmissionStatusSubmitted, assignmentID, portalusers.AccTypeStudent, submissions.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r listAssignmentSubmissionsSubmissionRow

		if err := rows.Scan(
			&r.SubmissionID,

			&r.UserID,
			&r.Email,
			&r.Phone,
			&r.Name,
			&r.Lastname,

			&r.Notes,
			&r.SubmittedAt,
			&r.Status,

			&r.Graded,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		subs = append(subs, ListAssignmentSubmissionsSubmission{
			ID: r.SubmissionID,

			SubmittedBy: ListAssignmentSubmissionsPortalUser{
				ID:       r.UserID,
				Email:    r.Email,
				Phone:    r.Phone,
				Name:     r.Name,
				Lastname: r.Lastname,
			},

			Notes:       r.Notes,
			SubmittedAt: r.SubmittedAt,
			Status:      r.Status,
			Graded:      r.Graded,
		})
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

	return &ListAssignmentSubmissionsResult{Access: access, Submissions: subs}, nil
}
