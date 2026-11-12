package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/permissions"
)

type StaffRole struct {
	ID       string
	Name     string
	Color    string
	Position int
}
type ListStaffRolesResult struct {
	Access   permissions.AccessData
	Roles    []StaffRole
	SchoolID int64
}
type ListStaffRolesInput struct {
	UserID  int64
	StaffID int64
}

func (s *Service) ListStaffRoles(ctx context.Context, i *ListStaffRolesInput) (*ListStaffRolesResult, error) {
	userID := i.UserID
	staffID := i.StaffID
	// Get the staff member's school ID
	var schoolID int64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id FROM school_staff WHERE id = $1
	`, staffID).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check permissions
	if !permissions.Can(permissions.PermissionStaffRoleList, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// StaffRole struct
	var roles []StaffRole

	// Fetch roles
	rows, err := s.DB.QueryContext(ctx, `
		SELECT sr.id, sr.name, sr.color, sr.position
		FROM staff_roles sr
		JOIN staff_role_members srm
		ON srm.role_id = sr.id
		WHERE srm.staff_id = $1
		AND sr.school_id = $2
		ORDER BY sr.position DESC
	`, staffID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r StaffRole

		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.Color,
			&r.Position,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		roles = append(roles, r)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	access, err := permissions.GetAllUserPermissions(ctx, s.DB, userID, schoolID)
	if err != nil {
		return nil, err
	}

	return &ListStaffRolesResult{Access: access, Roles: roles, SchoolID: schoolID}, nil
}
