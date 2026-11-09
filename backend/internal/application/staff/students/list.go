package students

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"time"
)

type ListStudentsInput struct {
	UserID   int64
	SchoolID int64
}

type guardian struct {
	ID       int64
	Name     string
	LastName string
	Email    string
	Phone    *string
}

type studentType struct {
	ID                int64
	Name              string
	LastName          string
	ProfilePictureURL *string
	DoB               *time.Time
	Email             *string
	Phone             *string
	Notes             *string
	AccountEnabled    bool
	AccountActive     bool
	CreatedAt         time.Time
	Guardians         []guardian
}

type ListStudentsResult struct {
	Students []studentType
}

func (s *Service) ListStudents(ctx context.Context, i *ListStudentsInput) (*ListStudentsResult, error) {
	if !schools.Can(schools.PermissionStudentList, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var studentList []studentType

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    pu.id, pu.name, pu.last_name, pu.date_of_birth, pu.email, pu.phone, pu.notes, pu.account_enabled, pu.account_active,
		    pu.created_at, so.object_key, so.bucket_name
		FROM portal_users pu
		LEFT JOIN portal_user_profile_pictures pfp ON pu.id = pfp.portal_user_id
		LEFT JOIN storage_objects so ON so.id = pfp.storage_object_id AND status = $3
		LEFT JOIN guardian_students gs ON gs.student_id = pu.id
		LEFT JOIN portal_users g ON gs.guardian_id = g.id
			AND g.account_type = $4
		WHERE pu.school_id = $1
		AND pu.account_type = $2
	`, i.SchoolID, helpers2.AccTypeStudent, helpers2.StatusDone, helpers2.AccTypeGuardian)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var st studentType

		var objKey *string
		var bucketName *string

		if err := rows.Scan(
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
			&objKey,
			&bucketName,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		if objKey != nil && bucketName != nil {
			presignedUrl, err := s.S3.PresignedGetObject(ctx, *bucketName, *objKey, 15*time.Minute, nil)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			url := presignedUrl.String()

			st.ProfilePictureURL = &url
		}

		studentList = append(studentList, st)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &ListStudentsResult{
		Students: studentList,
	}, nil
}
