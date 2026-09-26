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

type student struct {
	ID                int64
	Name              string
	LastName          string
	ProfilePictureURL *string
	DoB               *time.Time
	Email             *string
	Phone             *string
	Notes             *string
	AccountEnabled    bool
	CreatedAt         time.Time
}

type ListStudentsResult struct {
	Students []student
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

	var studentList []student

	rows, err := tx.QueryContext(ctx, `
		SELECT
		    pu.id, pu.name, pu.last_name, pu.date_of_birth, pu.email, pu.phone, pu.notes, pu.account_enabled, pu.created_at,
		    so.object_key, so.bucket_name
		FROM portal_users pu
		LEFT JOIN portal_user_profile_pictures pfp ON pu.id = pfp.portal_user_id
		LEFT JOIN storage_objects so ON so.id = pfp.storage_object_id AND status = $3
		WHERE school_id = $1
		AND account_type = $2
	`, i.SchoolID, helpers2.AccTypeStudent, helpers2.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var st student

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
