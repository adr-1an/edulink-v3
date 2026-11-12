package members

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"app/internal/helpers/staff/schools"
)

type CancelInvitationInput struct {
	UserID       int64
	InvitationID int64
}

type CancelInvitationResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) CancelInvitation(ctx context.Context, i *CancelInvitationInput) (*CancelInvitationResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	invID := i.InvitationID
	var schoolID int64
	var invEmail string
	if err := s.DB.QueryRowContext(ctx, `
		SELECT school_id, user_email FROM staff_invitations WHERE id = $1
	`, invID).Scan(&schoolID, &invEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	if !schools.Can(schools.PermissionSchoolInviteCancel, userID, schoolID, ctx, s.DB) {
		return nil, ErrForbidden
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE staff_invitations
		SET status = 'rejected'
		WHERE id = $1
	`, invID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	details := fmt.Sprintf("{user} canceled the staff invitation sent to %s.", invEmail)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationCancel, schools.TypeEdit, "Staff invitation canceled", "{user} canceled a staff invitation.", tx, ctx, s.Sf, details); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	affectedRows, err = res.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	noRowsAffected = noRowsAffected || affectedRows == 0

	return &CancelInvitationResult{NoRowsAffected: noRowsAffected}, nil
}
