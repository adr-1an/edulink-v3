package members

import (
	"context"
	"log"
	"time"

	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type UserSummary struct {
	ID     string
	Name   string
	Email  string
	PfpURL *string
}
type MemberRole struct {
	ID       string
	Position int
	Name     string
	Color    string
}
type Member struct {
	ID        string
	UserID    string
	User      UserSummary
	AddedBy   UserSummary
	CreatedAt time.Time
	Roles     []MemberRole
}
type ListStaffMembersResult struct {
	Access   schools.AccessData
	Staff    []Member
	SchoolID int64
}
type ListStaffMembersInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) ListStaffMembers(ctx context.Context, i *ListStaffMembersInput) (*ListStaffMembersResult, error) {
	userID := i.UserID
	schoolID := i.SchoolID
	// Check permissions
	if !schools.Can(schools.PermissionStaffView, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Get staff members
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		    s.id,
		    s.user_id,
		    u.id, u.name, u.email,
		    added_by.id, added_by.name, added_by.email,
		    s.created_at,
		    sr.id,
		    sr.position,
		    sr.name,
		    sr.color,
		    
		    so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type
		FROM school_staff s
		JOIN users u ON s.user_id = u.id
		JOIN users added_by ON s.added_by_user = added_by.id
		LEFT JOIN staff_role_members srm
		ON srm.staff_id = s.id
		LEFT JOIN staff_roles sr
		ON srm.role_id = sr.id
		LEFT JOIN user_profile_pictures upp ON upp.user_id = u.id
		LEFT JOIN storage_objects so ON so.id = upp.storage_object_id AND status = $2
		WHERE s.school_id = $1
		ORDER BY s.created_at, s.id, sr.position DESC
	`, schoolID, helpers2.StatusDone)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	staffByID := make(map[string]*Member)
	staffOrder := make([]string, 0)

	for rows.Next() {
		var (
			staffID   string
			userID    string
			user      UserSummary
			addedBy   UserSummary
			createdAt time.Time

			roleID       *string
			rolePosition *int64
			roleName     *string
			roleColor    *string

			bucketName  *string
			objectKey   *string
			filename    *string
			contentType *string
		)

		if err := rows.Scan(
			&staffID,
			&userID,
			&user.ID,
			&user.Name,
			&user.Email,
			&addedBy.ID,
			&addedBy.Name,
			&addedBy.Email,
			&createdAt,
			&roleID,
			&rolePosition,
			&roleName,
			&roleColor,
			&bucketName,
			&objectKey,
			&filename,
			&contentType,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		staff, exists := staffByID[staffID]
		if !exists {
			// Profile pic
			if bucketName != nil && objectKey != nil && filename != nil && contentType != nil {
				// Generate presigned URL
				url, err := s.S3.PresignedGetObject(ctx, *bucketName, *objectKey, 15*time.Minute, nil)
				if err != nil {
					log.Println(err)
					return nil, err
				}

				pfpURL := url.String()
				user.PfpURL = &pfpURL
			}

			staff = &Member{
				ID:        staffID,
				UserID:    userID,
				User:      user,
				AddedBy:   addedBy,
				CreatedAt: createdAt,
				Roles:     make([]MemberRole, 0),
			}

			staffByID[staffID] = staff
			staffOrder = append(staffOrder, staffID)
		}

		if roleID != nil {
			staff.Roles = append(staff.Roles, MemberRole{
				ID:       *roleID,
				Position: int(*rolePosition),
				Name:     *roleName,
				Color:    *roleColor,
			})
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	sMembers := make([]Member, 0, len(staffOrder))
	for _, id := range staffOrder {
		sMembers = append(sMembers, *staffByID[id])
	}

	access, err := schools.GetAllUserPermissions(ctx, s.DB, userID, schoolID)
	if err != nil {
		return nil, err
	}

	return &ListStaffMembersResult{Access: access, Staff: sMembers, SchoolID: schoolID}, nil
}
