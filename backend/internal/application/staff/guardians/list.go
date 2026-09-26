package guardians

import (
	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"time"
)

type ListGuardian struct {
	ID             int64
	Name           string
	LastName       string
	Email          string
	Phone          *string
	Notes          *string
	DateOfBirth    time.Time
	AccountEnabled bool
	AccountActive  bool
}

type ListGuardiansResult struct {
	Guardians []ListGuardian
}

type ListGuardiansInput struct {
	UserID   int64
	SchoolID int64
}

func (s *Service) ListGuardians(ctx context.Context, i *ListGuardiansInput) (*ListGuardiansResult, error) {
	if !schools.Can(schools.PermissionGuardianList, i.UserID, i.SchoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		    id,
		    name,
		    last_name,
		    email,
		    phone,
		    notes,
		    date_of_birth,
		    account_enabled,
		    account_active
		FROM portal_users
		WHERE school_id = $1
		AND account_type = $2
	`, i.SchoolID, helpers.AccTypeGuardian)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var guardians []ListGuardian
	for rows.Next() {
		var g ListGuardian

		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.LastName,
			&g.Email,
			&g.Phone,
			&g.Notes,
			&g.DateOfBirth,
			&g.AccountEnabled,
			&g.AccountActive,
		); err != nil {
			log.Println(err)
			return nil, err
		}

		guardians = append(guardians, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListGuardiansResult{
		Guardians: guardians,
	}, nil
}
