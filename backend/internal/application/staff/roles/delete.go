package roles

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/helpers/staff/schools"
)

type DeleteRoleInput struct {
	UserID int64
	RoleID int64
}

type DeleteRoleResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) DeleteRole(ctx context.Context, i *DeleteRoleInput) (*DeleteRoleResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	roleID := i.RoleID
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
	if !schools.Can(schools.PermissionRoleDelete, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	order, err := lockSchoolRoleOrder(ctx, tx, schoolID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	deleteIndex := -1
	currentPosition := -1
	for i, role := range order {
		if role.ID == roleID {
			deleteIndex = i
			currentPosition = role.Position
			break
		}
	}
	if deleteIndex == -1 {
		return nil, ErrForbidden
	}

	// Check if the actor can delete this role, or if the actor is the school owner
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, currentPosition)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !allowed {
		return nil, ErrForbidden
	}

	// Delete role
	res, err := tx.ExecContext(ctx, `
		DELETE FROM staff_roles
		WHERE id = $1
		  AND school_id = $2
	`, roleID, schoolID)
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

	orderedRoleIDs := roleIDs(order)
	orderedRoleIDs = append(orderedRoleIDs[:deleteIndex], orderedRoleIDs[deleteIndex+1:]...)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleDelete, schools.TypeDelete, "Role deleted", "{user} deleted the role '"+roleName+"'.", tx, ctx, s.Sf, "{user} deleted role '"+roleName+"' with ID "+strconv.FormatInt(roleID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteRoleResult{NoRowsAffected: noRowsAffected}, nil
}
