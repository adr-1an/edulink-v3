package students

import (
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

type ViewStudentInput struct {
	UserID    int64
	StudentID int64
}

type grade struct {
	ID    int64
	Name  string
	Level int
}

type course struct {
	ID    int64
	Grade grade
	Name  string
	Color string
}

type assignment struct {
	ID          int64
	Course      course
	Title       string
	Description *string
}

type assignmentScore struct {
	ScorePercentage int
}

type submission struct {
	ID          int64
	Assignment  assignment
	Score       *assignmentScore
	SubmittedAt time.Time
	Notes       *string
}

type ViewStudentGuardian struct {
	ID       int64
	Name     string
	LastName string
	Email    string
	Phone    *string
}

type profilePicture struct {
	PresignedURL string
}

type student struct {
	ProfilePicture *profilePicture
	ID             int64
	Name           string
	LastName       string
	DoB            string
	Email          string
	Phone          *string
	Notes          *string
	AccountEnabled bool
	AccountActive  bool
	CreatedAt      time.Time
	Guardians      []ViewStudentGuardian
}

type ViewStudentResult struct {
	SchoolID    int64
	Submissions []submission
	Student     student
}

func (s *Service) ViewStudent(ctx context.Context, i *ViewStudentInput) (*ViewStudentResult, error) {
	schoolID, err := staff_helpers.StudentToSchoolID(ctx, s.DB, i.StudentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		}
		log.Println(err)
		return nil, err
	}

	if !schools.Can(schools.PermissionStudentView, i.UserID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	var st student

	// Get the student data & submissions
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
			s.id, s.created_at, s.notes,
			ss.score_percentage,
			a.id, a.title, a.description,
			c.id, c.name, c.color,
			g.id, g.name, g.level,
		
			st.id, st.name, st.last_name, st.date_of_birth, st.email, st.phone,
			st.notes, st.account_enabled, st.account_active, st.created_at,
			
			gu.id, gu.name, gu.last_name, gu.email, gu.phone
		FROM portal_users st
		LEFT JOIN assignment_submissions s
			ON s.submitted_by = st.id
			AND s.status = $1
		LEFT JOIN submission_scores ss ON ss.submission_id = s.id
		LEFT JOIN assignments a ON a.id = s.assignment_id
		LEFT JOIN courses c ON c.id = a.course_id
		LEFT JOIN grades g ON g.id = c.grade_id
		LEFT JOIN guardian_students gs ON gs.student_id = st.id
		LEFT JOIN portal_users gu ON gu.id = gs.guardian_id
			AND gu.account_type = $3
		WHERE st.id = $2
	`, helpers.SubmissionStatusSubmitted, i.StudentID, helpers.AccTypeGuardian)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var submissions []submission

	for rows.Next() {
		var (
			sub                           *submission
			subScore                      *int
			subID                         *int64
			submittedAt                   *time.Time
			subNotes                      *string
			subAssignmentID               *int64
			subAssignmentTitle            *string
			subAssignmentDescription      *string
			subAssignmentCourseID         *int64
			subAssignmentCourseName       *string
			subAssignmentCourseColor      *string
			subAssignmentCourseGradeID    *int64
			subAssignmentCourseGradeName  *string
			subAssignmentCourseGradeLevel *int

			gID       *int64
			gName     *string
			gLastName *string
			gEmail    *string
			gPhone    *string
		)

		if err := rows.Scan(
			&subID,
			&submittedAt,
			&subNotes,

			&subScore,

			&subAssignmentID,
			&subAssignmentTitle,
			&subAssignmentDescription,

			&subAssignmentCourseID,
			&subAssignmentCourseName,
			&subAssignmentCourseColor,

			&subAssignmentCourseGradeID,
			&subAssignmentCourseGradeName,
			&subAssignmentCourseGradeLevel,

			&st.ID,
			&st.Name,
			&st.LastName,
			&st.DoB,
			&st.Email,
			&st.Phone,
			&st.Notes,
			&st.AccountEnabled,
			&st.AccountActive,
			&st.CreatedAt,

			&gID,
			&gName,
			&gLastName,
			&gEmail,
			&gPhone,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if subID != nil && subAssignmentID != nil && subAssignmentCourseID != nil && subAssignmentCourseGradeID != nil {
			sub = &submission{
				ID: *subID,
				Assignment: assignment{
					ID: *subAssignmentID,
					Course: course{
						ID: *subAssignmentCourseID,
						Grade: grade{
							ID:    *subAssignmentCourseGradeID,
							Name:  *subAssignmentCourseGradeName,
							Level: *subAssignmentCourseGradeLevel,
						},
						Name:  *subAssignmentCourseName,
						Color: *subAssignmentCourseColor,
					},
					Title:       *subAssignmentTitle,
					Description: subAssignmentDescription,
				},
				Score:       nil,
				SubmittedAt: *submittedAt,
				Notes:       subNotes,
			}

			if subScore != nil {
				sub.Score = &assignmentScore{
					ScorePercentage: *subScore,
				}
			}
			submissions = append(submissions, *sub)
		}

		if gID != nil && gName != nil && gLastName != nil && gEmail != nil {
			st.Guardians = append(st.Guardians, ViewStudentGuardian{
				ID:       *gID,
				Name:     *gName,
				LastName: *gLastName,
				Email:    *gEmail,
				Phone:    gPhone,
			})
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	if err := rows.Close(); err != nil {
		log.Println(err)
		return nil, err
	}

	if st.Notes != nil && *st.Notes == "" {
		st.Notes = nil
	}

	var pfpBucketName string
	var pfpObjectKey string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT so.bucket_name, so.object_key
		FROM portal_user_profile_pictures p
		JOIN storage_objects so ON so.id = p.storage_object_id
		WHERE p.portal_user_id = $1
		  AND so.status = $2
	`, i.StudentID, helpers.StatusDone).Scan(&pfpBucketName, &pfpObjectKey); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
			return nil, err
		}
	} else {
		url, err := s.S3.PresignedGetObject(ctx, pfpBucketName, pfpObjectKey, 15*time.Minute, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		st.ProfilePicture = &profilePicture{
			PresignedURL: url.String(),
		}
	}

	return &ViewStudentResult{
		SchoolID:    schoolID,
		Submissions: submissions,
		Student:     st,
	}, err
}
