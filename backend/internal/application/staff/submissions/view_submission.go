package submissions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"app/internal/permissions"
	"app/internal/submissions"
	"app/internal/uploads"
)

type ViewSubmissionUser struct {
	ID    string
	Name  string
	Email string
	Phone string
}

type ViewSubmissionGrade struct {
	Score    int
	Notes    *string
	GradedAt time.Time
	GradedBy ViewSubmissionUser
}

type ViewSubmissionPortalUser struct {
	ID       string
	Email    string
	Phone    *string
	Name     string
	Lastname string
}

type ViewSubmissionSubmissionAttachment struct {
	ID           string
	PresignedURL string
	Filename     string
	ContentType  string
}

type ViewSubmissionSubmission struct {
	ID          string
	Status      string
	SubmittedBy ViewSubmissionPortalUser
	Attachments []ViewSubmissionSubmissionAttachment
	Grade       *ViewSubmissionGrade
	Notes       *string
	SubmittedAt time.Time
}

type viewSubmissionSubRow struct {
	ID        string
	Status    string
	Notes     *string
	CreatedAt time.Time

	GradeScore    *int
	GradeNotes    *string
	GradedAt      *time.Time
	GradedByID    *string
	GradedByName  *string
	GradedByEmail *string
	GradedByPhone *string

	UserID   string
	Email    string
	Phone    string
	Name     string
	Lastname string

	AttachmentID *string
	Filename     *string
	ContentType  *string
	BucketName   *string
}

type ViewSubmissionInput struct {
	UserID       int64
	SubmissionID int64
}

type ViewSubmissionResult struct {
	Access     permissions.AccessData
	Submission ViewSubmissionSubmission
}

func (svc *Service) ViewSubmission(ctx context.Context, i *ViewSubmissionInput) (*ViewSubmissionResult, error) {
	userID := i.UserID
	subID := i.SubmissionID

	schoolID, err := SubToSchoolID(subID, svc.DB, ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}
	}

	if !permissions.Can(permissions.PermissionSubmissionView, userID, schoolID, ctx, svc.DB) {
		return nil, ErrForbidden
	}

	rows, err := svc.DB.QueryContext(ctx, `
		SELECT
		    s.id,
		    s.status,
		    s.notes,
		    s.created_at,
		    
		    ss.score_percentage,
		    ss.notes,
		    ss.graded_at,
		    
		    gu.id,
		    gu.name,
		    gu.email,
		    gu.phone,
		    
		    u.id,
		    u.email,
		    u.phone,
		    u.name,
		    u.last_name,
		    
		    so.id,
		    so.original_file_name,
		    so.declared_content_type,
		    so.bucket_name
		FROM assignment_submissions s
		JOIN portal_users u ON s.submitted_by = u.id
		LEFT JOIN submission_attachments sa ON s.id = sa.submission_id
		LEFT JOIN portal_storage_objects so ON sa.storage_object_id = so.id
		LEFT JOIN submission_scores ss ON s.id = ss.submission_id
		LEFT JOIN users gu ON ss.graded_by = gu.id
		WHERE s.id = $1
		AND (s.status = $2 OR s.status = $3)
	`, subID, submissions.SubmissionStatusSubmitted, submissions.SubmissionStatusReturned)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var (
		s           ViewSubmissionSubmission
		initialized bool
	)

	for rows.Next() {
		var r viewSubmissionSubRow

		if err := rows.Scan(
			&r.ID,
			&r.Status,
			&r.Notes,
			&r.CreatedAt,

			&r.GradeScore,
			&r.GradeNotes,
			&r.GradedAt,

			&r.GradedByID,
			&r.GradedByName,
			&r.GradedByEmail,
			&r.GradedByPhone,

			&r.UserID,
			&r.Email,
			&r.Phone,
			&r.Name,
			&r.Lastname,

			&r.AttachmentID,
			&r.Filename,
			&r.ContentType,
			&r.BucketName,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if !initialized {
			s = ViewSubmissionSubmission{
				ID:     r.ID,
				Status: r.Status,
				SubmittedBy: ViewSubmissionPortalUser{
					ID:       r.UserID,
					Email:    r.Email,
					Phone:    &r.Phone,
					Name:     r.Name,
					Lastname: r.Lastname,
				},
				Grade:       nil,
				Attachments: make([]ViewSubmissionSubmissionAttachment, 0),
				SubmittedAt: r.CreatedAt,
			}

			if r.Notes != nil {
				s.Notes = r.Notes
			}

			initialized = true
		}

		if r.AttachmentID != nil {
			objKey := fmt.Sprintf("%s/%s", uploads.UploadCategorySubmissionAttachments, *r.AttachmentID)

			url, err := svc.S3.PresignedGetObject(ctx, *r.BucketName, objKey, 5*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			s.Attachments = append(s.Attachments, ViewSubmissionSubmissionAttachment{
				ID:           *r.AttachmentID,
				PresignedURL: url.String(),
				Filename:     *r.Filename,
				ContentType:  *r.ContentType,
			})
		}

		if r.GradeScore != nil && r.GradedByID != nil {
			s.Grade = &ViewSubmissionGrade{
				Score:    *r.GradeScore,
				Notes:    r.GradeNotes,
				GradedAt: *r.GradedAt,
				GradedBy: ViewSubmissionUser{
					ID:    *r.GradedByID,
					Name:  *r.GradedByName,
					Email: *r.GradedByEmail,
					Phone: *r.GradedByPhone,
				},
			}
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

	return &ViewSubmissionResult{Access: access, Submission: s}, nil
}
