package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"app/internal/helpers/staff/schools"
)

type AddStaffRoleInput struct {
	UserID  int64
	StaffID int64
	RoleID  int64
}

func (s *Service) AddStaffRole(ctx context.Context, i *AddStaffRoleInput) error {
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
	if !schools.Can(schools.PermissionStaffRoleAdd, userID, target.StaffSchoolID, ctx, s.DB) {
		return ErrForbidden
	}

	// Check if the user can assign this role
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, target.StaffSchoolID, target.RolePosition)
	if err != nil {
		log.Println(err)
		return err
	}

	if !allowed {
		return ErrForbidden
	}

	// Add role to staff member
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO staff_role_members (staff_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, staffID, roleID); err != nil {
		log.Println(err)
		return err
	}

	// Log action
	if err := schools.StoreSchoolLog(target.StaffSchoolID, userID, schools.ActionStaffRoleCreate, schools.TypeCreate, "Staff role added", "{user} added the role '"+target.RoleName+"' to "+target.StaffName+".", tx, ctx, s.Sf, "{user} added role '"+target.RoleName+"' to staff member "+target.StaffName+"."); err != nil {
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
