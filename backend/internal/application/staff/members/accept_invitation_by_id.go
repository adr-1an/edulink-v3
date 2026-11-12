package members

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"

	"app/internal/schoollogs"
)

type AcceptInvitationByIDInput struct {
	UserID       int64
	InvitationID int64
}

type AcceptInvitationByIDResult struct {
	// NoRowsAffected preserves the result of the original write without changing transaction flow.
	NoRowsAffected bool
}

func (s *Service) AcceptInvitationByID(ctx context.Context, i *AcceptInvitationByIDInput) (*AcceptInvitationByIDResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	invitationID := i.InvitationID
	type StaffInvitation struct {
		SchoolID   int64
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

	// Get staff invitation from ID
	if err := tx.QueryRowContext(ctx, `
		SELECT si.school_id, si.sent_by_user
		FROM staff_invitations si
		JOIN users u ON u.email = si.user_email
		WHERE si.id = $1
		AND u.id = $2
		AND si.expires_at > NOW()
		AND si.status = 'pending'
		FOR UPDATE
	`, invitationID, userID).Scan(&inv.SchoolID, &inv.SentByUser); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
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
	`, id, inv.SchoolID, userID, inv.SentByUser)
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

	// Log invitation action
	if err := schoollogs.StoreSchoolLog(inv.SchoolID, userID, schoollogs.ActionStaffInvitationDelete, schoollogs.TypeOther, "Staff invitation accepted", "{user} accepted a staff invitation.", tx, ctx, s.Sf, "{user} accepted staff invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Log action
	if err := schoollogs.StoreSchoolLog(inv.SchoolID, userID, schoollogs.ActionStaffCreate, schoollogs.TypeCreate, "Staff member added", "{user} joined the school.", tx, ctx, s.Sf, "{user} was added as a staff member from invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit TX
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	return &AcceptInvitationByIDResult{NoRowsAffected: noRowsAffected}, nil
}
