package guardians

import (
	"app/internal/helpers"
	"app/internal/helpers/staff/schools"
	"context"
	"database/sql"
	"errors"
	"log"
)

type DeleteGuardianInput struct {
	UserID     int64
	GuardianID int64
}

func (s *Service) DeleteGuardian(ctx context.Context, i *DeleteGuardianInput) error {
	var schoolID int64
	if err := s.DB.QueryRowContext(ctx, `
		SELECT
		    school_id
		FROM portal_users
		WHERE id = $1
	  	AND account_type = $2
	`, i.GuardianID, helpers.AccTypeGuardian).Scan(&schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		log.Println(err)
		return err
	}

	if !schools.Can(schools.PermissionGuardianDelete, i.UserID, schoolID, ctx, s.DB) {
		return ErrForbidden
	}

	if _, err := s.DB.ExecContext(ctx, `
		DELETE FROM portal_users WHERE id = $1
	`, i.GuardianID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
