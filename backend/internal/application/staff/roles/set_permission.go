package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"app/internal/permissions"
	"app/internal/schoollogs"
)

type SetStaffRolePermissionInput struct {
	InvalidPayload bool
	UserID         int64
	RoleID         int64
	Role           SetStaffRolePermission
}

type SetStaffRolePermission struct {
	Permission string
	Allow      bool
}

func (s *Service) SetStaffRolePermission(ctx context.Context, i *SetStaffRolePermissionInput) error {
	userID := i.UserID
	roleID := i.RoleID
	p := i.Role
	// Get the role's school ID and position
	var position int
	var schoolID int64
	var roleName string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT position, school_id, name
		FROM staff_roles
		WHERE id = $1
	`, roleID).Scan(&position, &schoolID, &roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}

	}

	// Check permissions
	if !permissions.Can(permissions.PermissionRolePermissionUpdate, userID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if i.InvalidPayload {
		return ErrBadRequest
	}

	// Parse
	p.Permission = strings.TrimSpace(p.Permission)

	// Validate
	if !permissions.ValidatePermissionString(p.Permission) {
		return ErrUnprocessableEntity
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Check if the user owns the school or can edit the role
	allowed, err := permissions.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, position)
	if err != nil {
		log.Println(err)
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	if p.Allow {
		// insert the permission
		if _, err := tx.ExecContext(ctx, `
		INSERT INTO staff_role_permissions (role_id, permission)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission) DO NOTHING
	`, roleID, p.Permission); err != nil {
			log.Println(err)
			return err
		}
	} else {
		// delete the permission
		if _, err := tx.ExecContext(ctx, `
			DELETE FROM staff_role_permissions rp
			USING staff_roles r
			WHERE rp.role_id = r.id
			AND rp.permission = $1
			AND r.school_id = $2
			AND rp.role_id = $3
		`, p.Permission, schoolID, roleID); err != nil {
			log.Println(err)
			return err
		}
	}

	title := "Role permission revoked"
	message := "{user} revoked the permission '" + p.Permission + "' from the role '" + roleName + "'."
	if p.Allow {
		title = "Role permission granted"
		message = "{user} granted the permission '" + p.Permission + "' to the role '" + roleName + "'."
	}

	// Log action
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionRolePermissionEdit, schoollogs.TypeEdit, title, message, tx, ctx, s.Sf); err != nil {
		log.Println(err)
		return err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
