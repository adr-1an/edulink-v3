package roles

import (
	"context"
	"log"
	"strconv"
	"strings"

	"app/internal/helpers/staff/schools"
)

type CreateRoleInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	Role           CreateRole
}

type CreateRole struct {
	Name     string
	Position int
	Color    *string
}

type CreateRoleResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) CreateRole(ctx context.Context, i *CreateRoleInput) (*CreateRoleResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	schoolID := i.SchoolID
	p := i.Role
	// Check permissions
	if !schools.Can(schools.PermissionRoleCreate, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)
	var color string
	if p.Color != nil {
		color = *p.Color
	}

	// Validate
	if p.Name == "" || len(p.Name) > 32 {
		return nil, ErrUnprocessableEntity
	}
	if len(color) != 6 { // invalid hex len
		return nil, ErrUnprocessableEntity
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
	if p.Position < 0 || p.Position > len(order) {
		return nil, ErrUnprocessableEntity
	}

	// Check if the actor can create a role at this position, or if the actor is the school owner.
	allowed, err := schools.OwnsSchoolOrHasRoleAbovePosition(ctx, tx, userID, schoolID, p.Position)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !allowed {
		return nil, ErrForbidden
	}

	// Check for a name conflict. Occupied positions are shifted automatically.
	var conflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM staff_roles
			WHERE name = $1
			AND school_id = $2
		)
	`, p.Name, schoolID).Scan(&conflict); err != nil {
		log.Println(err)
		return nil, err
	}

	if conflict {
		return nil, ErrConflict
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var temporaryPosition int
	if len(order) > 0 {
		temporaryPosition = order[len(order)-1].Position + 1
	}

	// Create the role in a free temporary slot, then rank every role contiguously.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO staff_roles (id, position, school_id, created_by, name, color)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, temporaryPosition, schoolID, userID, p.Name, color)
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

	orderedRoleIDs := insertRoleID(roleIDs(order), id, p.Position)
	if err := persistRoleOrder(ctx, tx, schoolID, orderedRoleIDs); err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionRoleCreate, schools.TypeCreate, "Role created", "{user} created the role '"+p.Name+"'.", tx, ctx, s.Sf, "{user} created role '"+p.Name+"' at position "+strconv.Itoa(p.Position)+" with color '#"+color+"'."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &CreateRoleResult{NoRowsAffected: noRowsAffected}, nil
}
