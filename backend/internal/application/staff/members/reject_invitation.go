package members

import (
	"context"
	"database/sql"
	"errors"
	"log"

	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type RejectInvitationInput struct {
	UserID int64
	Token  string
}

type RejectInvitationResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) RejectInvitation(ctx context.Context, i *RejectInvitationInput) (*RejectInvitationResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	token := i.Token
	tokenHash := helpers2.MakeHash256(token)

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
		    SELECT user_email, school_id FROM staff_invitations WHERE token_hash = $1 AND status <> 'rejected'
		)
		SELECT users.id, inv_email.school_id
		FROM users, inv_email
		WHERE users.email = inv_email.user_email 
		FOR UPDATE
	`, tokenHash).Scan(&invUserID, &schoolID); err != nil {
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
		WHERE token_hash = $1
	`, tokenHash)
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
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationEdit, schools.TypeEdit, "Staff invitation rejected", "{user} rejected a staff invitation.", tx, ctx, s.Sf, "{user} rejected a staff invitation by token."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &RejectInvitationResult{NoRowsAffected: noRowsAffected}, nil
}
