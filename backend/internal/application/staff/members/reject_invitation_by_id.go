package members

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/schoollogs"
)

type RejectInvitationByIDInput struct {
	UserID       int64
	InvitationID int64
}

type RejectInvitationByIDResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) RejectInvitationByID(ctx context.Context, i *RejectInvitationByIDInput) (*RejectInvitationByIDResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	invitationID := i.InvitationID
	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get the invitation user ID and school ID
	var invUserID int64
	var schoolID int64
	if err := tx.QueryRowContext(ctx, `
		WITH inv_email AS (
		    SELECT user_email, school_id FROM staff_invitations WHERE id = $1 AND status <> 'rejected'
		)
		SELECT users.id, inv_email.school_id
		FROM users, inv_email
		WHERE users.email = inv_email.user_email
		FOR UPDATE
	`, invitationID).Scan(&invUserID, &schoolID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Compare IDs
	if userID != invUserID {
		// The user is trying to reject an invitation which is not meant for their account.
		return nil, ErrForbidden
	}

	// Update invitation status to rejected
	res, err := tx.ExecContext(ctx, `
		UPDATE staff_invitations
		SET status = 'rejected'
		WHERE id = $1
	`, invitationID)
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

	// Log action
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionStaffInvitationEdit, schoollogs.TypeOther, "Staff invitation rejected", "{user} rejected a staff invitation.", tx, ctx, s.Sf, "{user} rejected staff invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &RejectInvitationByIDResult{NoRowsAffected: noRowsAffected}, nil
}
