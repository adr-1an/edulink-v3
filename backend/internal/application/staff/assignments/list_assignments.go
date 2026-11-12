package assignments

import (
	"context"
	"log"
	"strconv"
	"time"

	"app/internal/permissions"
)

type ListAssignmentsPost struct {
	ID    string
	Title string
}

type ListAssignmentsAssignment struct {
	RefPost            *ListAssignmentsPost
	ID                 string
	Title              string
	Description        string
	DueDate            *time.Time
	SubmissionsEnabled bool
	SubmissionsCloseAt *time.Time
	CreatedAt          time.Time
	ArchivedAt         *time.Time
}

type ListAssignmentsInput struct {
	UserID   int64
	CourseID int64
}

type ListAssignmentsResult struct {
	Access      permissions.AccessData
	Assignments []ListAssignmentsAssignment
}

func (svc *Service) ListAssignments(ctx context.Context, i *ListAssignmentsInput) (*ListAssignmentsResult, error) {
	userID := i.UserID
	courseID := i.CourseID

	schoolID, err := courseToSchoolID(courseID, svc.DB, ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !permissions.Can(permissions.PermissionCourseAssignmentList, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	var assignments []ListAssignmentsAssignment

	rows, err := svc.DB.QueryContext(ctx, `
		SELECT
		    p.id, p.title,
		    
		    a.id,
		    a.title,
		    a.description,
		    a.due_date,
		    a.submissions_enabled,
		    a.submissions_close_at,
		    a.created_at,
		    a.archived_at
		FROM assignments a
		LEFT JOIN course_posts p
		ON a.referenced_post_id = p.id
		WHERE a.course_id = $1
	`, courseID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			a            ListAssignmentsAssignment
			refPostID    *int64
			refPostTitle *string
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
			&a.ArchivedAt,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if refPostID != nil {
			a.RefPost = &ListAssignmentsPost{
				ID:    strconv.FormatInt(*refPostID, 10),
				Title: *refPostTitle,
			}
		}

		assignments = append(assignments, a)
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

	return &ListAssignmentsResult{Access: access, Assignments: assignments}, nil
}
