package members

import (
	"context"
	"database/sql"
	"errors"
	"log"

	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff/schools"
)

type AcceptInvitationInput struct {
	UserID int64
	Token  string
}

type AcceptInvitationResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) AcceptInvitation(ctx context.Context, i *AcceptInvitationInput) (*AcceptInvitationResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	token := i.Token
	tokenHash := helpers2.MakeHash256(token)

	type StaffInvitation struct {
		SchoolID   int64
		UserEmail  string
		SentByUser int64
	}
	var inv StaffInvitation

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Get staff invitation from token
	if err := tx.QueryRowContext(ctx, `
		SELECT school_id, user_email, sent_by_user
		FROM staff_invitations
		WHERE token_hash = $1
		AND expires_at > NOW()
		AND status = 'pending'
		FOR UPDATE
	`, tokenHash).Scan(&inv.SchoolID, &inv.UserEmail, &inv.SentByUser); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Get user ID from email
	var staffUserID int64
	// Redundant ID check is intentional
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM users WHERE email = $1 AND id = $2
	`, inv.UserEmail, userID).Scan(&staffUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// The user is trying to accept an invitation which is not meant for their account.
			return nil, ErrForbidden
		} else {
			log.Println(err)
			return nil, err
		}

	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Insert staff member row
	res, err := tx.ExecContext(ctx, `
		INSERT INTO school_staff (id, school_id, user_id, added_by_user)
		VALUES ($1, $2, $3, $4)
	`, id, inv.SchoolID, staffUserID, inv.SentByUser)
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

	// Delete invitation
	res, err = tx.ExecContext(ctx, `
		DELETE FROM staff_invitations
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

	// Log invitation action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffInvitationDelete, schools.TypeDelete, "Staff invitation accepted", "{user} accepted a staff invitation.", tx, ctx, s.Sf, "{user} accepted the staff invitation sent to '"+inv.UserEmail+"'."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffCreate, schools.TypeCreate, "Staff member added", "{user} joined the school.", tx, ctx, s.Sf, "{user} was added as a staff member after accepting an invitation."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit TX
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &AcceptInvitationResult{NoRowsAffected: noRowsAffected}, nil
}
