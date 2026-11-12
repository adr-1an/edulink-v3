package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/helpers/staff/schools"
)

type RemoveStaffRoleInput struct {
	UserID  int64
	StaffID int64
	RoleID  int64
}

func (s *Service) RemoveStaffRole(ctx context.Context, i *RemoveStaffRoleInput) error {
	userID := i.UserID
	staffID := i.StaffID
	roleID := i.RoleID
	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Get staff and role data
	target, err := getStaffRoleTarget(ctx, tx, staffID, roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		} else {
			log.Println(err)
			return err
		}

	}

	if target.StaffSchoolID != target.RoleSchoolID {
		return ErrForbidden
	}

	// Check permissions
	if !schools.Can(schools.PermissionStaffRoleRemove, userID, target.StaffSchoolID, ctx, s.DB) {
		return ErrForbidden
	}

	// Check if the user can remove this role
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, target.StaffSchoolID, target.RolePosition)
	if err != nil {
		log.Println(err)
		return err
	}

	if !allowed {
		return ErrForbidden
	}

	// Remove role from staff member
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM staff_role_members
		WHERE staff_id = $1
		AND role_id = $2
	`, staffID, roleID); err != nil {
		log.Println(err)
		return err
	}

	// Log action
	if err := schools.StoreSchoolLog(target.StaffSchoolID, userID, schools.ActionStaffRoleDelete, schools.TypeDelete, "Staff role removed", "{user} removed the role '"+target.RoleName+"' from "+target.StaffName+".", tx, ctx, s.Sf, "{user} removed role '"+target.RoleName+"' from staff member "+target.StaffName+"."); err != nil {
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
