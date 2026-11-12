package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"

	"app/internal/helpers/staff/schools"
)

type UpdateRoleInput struct {
	InvalidPayload bool
	UserID         int64
	RoleID         int64
	Role           UpdateRole
}

type UpdateRole struct {
	Name     string
	Position int
	Color    string
}

type UpdateRoleResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) UpdateRole(ctx context.Context, i *UpdateRoleInput) (*UpdateRoleResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	roleID := i.RoleID
	p := i.Role
	// Get the role's school ID
	var schoolID int64
	var roleName string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id, name FROM staff_roles WHERE id = $1
	`, roleID).Scan(&schoolID, &roleName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Check permissions
	if !schools.Can(schools.PermissionRoleUpdate, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)

	// Validate
	if p.Name == "" || len(p.Name) > 32 {
		return nil, ErrUnprocessableEntity
	}
	if len(p.Color) != 6 {
		return nil, ErrUnprocessableEntity
	}

	order, err := lockSchoolRoleOrder(ctx, tx, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if p.Position < 0 || p.Position >= len(order) {
		return nil, ErrUnprocessableEntity
	}

	currentIndex := -1
	currentPosition := -1
	for i, role := range order {
		if role.ID == roleID {
			currentIndex = i
			currentPosition = role.Position
			break
		}
	}
	if currentIndex == -1 {
		return nil, ErrForbidden
	}

	var oldName string
	var oldColor string
	if err := tx.QueryRowContext(ctx, `
		SELECT name, color
		FROM staff_roles
		WHERE id = $1
		AND school_id = $2
	`, roleID, schoolID).Scan(&oldName, &oldColor); err != nil {
		log.Println(err)
		return nil, err
	}

	// Check if the actor can update the target role and move it to this position, or if the actor is the school owner
	canUpdateCurrent, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, currentPosition)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	canSetPosition, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, p.Position)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !canUpdateCurrent || !canSetPosition {
		return nil, ErrForbidden
	}

	// Check for a name conflict. Moving to an occupied rank shifts the roles in between.
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM staff_roles
			WHERE school_id = $1
			  AND id <> $2
			  AND name = $3
		)
	`, schoolID, roleID, p.Name).Scan(&conflict); err != nil {
		log.Println(err)
		return nil, err
	}

	if conflict {
		return nil, ErrConflict
	}

	// Update the role fields, then atomically rebuild the contiguous role order.
	res, err := tx.ExecContext(ctx, `
		UPDATE staff_roles
		SET name = $1,
		    color = $2,
		    updated_at = NOW()
		WHERE id = $3
		  AND school_id = $4
	`, p.Name, p.Color, roleID, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	orderedRoleIDs := moveRoleID(roleIDs(order), currentIndex, p.Position)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	details := "{user} updated the role name from '" + oldName + "' to '" + p.Name + "', color from '#" + oldColor + "' to '#" + p.Color + "', and position from " + strconv.Itoa(currentPosition) + " to " + strconv.Itoa(p.Position) + "."
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleEdit, schools.TypeEdit, "Role updated", "{user} updated the role '"+p.Name+"'.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateRoleResult{NoRowsAffected: noRowsAffected}, nil
}
