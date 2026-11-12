package members

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	mail2 "net/mail"
	"os"
	"strings"

	"app/internal/cryptoutil"
	"app/internal/mailer"
	"app/internal/permissions"
	"app/internal/schoollogs"

	"github.com/matoous/go-nanoid/v2"
	"github.com/wneessen/go-mail"
)

type SendInvitation struct {
	Email      string
	Importance string
}
type SendInvitationInput struct {
	InvalidPayload bool
	UserID         int64
	SchoolID       int64
	Invitation     SendInvitation
}
type SendInvitationResult struct {
	StaffCreated   bool
	NoRowsAffected bool
}

func (s *Service) SendInvitation(ctx context.Context, i *SendInvitationInput) (*SendInvitationResult, error) {
	var noRowsAffected bool
	var affectedRows int64

	userID := i.UserID
	schoolID := i.SchoolID
	p := i.Invitation
	// Check permissions
	allowed := permissions.Can(permissions.PermissionStaffCreate, userID, schoolID, ctx, s.DB)
	if !allowed {
		return nil, ErrForbidden
	}

	if i.InvalidPayload {
		return nil, ErrBadRequest
	}

	// Parse & validate
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.Importance = strings.ToLower(strings.TrimSpace(p.Importance))

	addr, err := mail2.ParseAddress(p.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}
	p.Email = addr.Address

	var importance mail.Importance
	switch p.Importance {
	case mail.ImportanceLow.String():
		importance = mail.ImportanceLow
	case mail.ImportanceNormal.String():
		importance = mail.ImportanceNormal
	case mail.ImportanceHigh.String():
		importance = mail.ImportanceHigh
	case mail.ImportanceNonUrgent.String():
		importance = mail.ImportanceNonUrgent
	case mail.ImportanceUrgent.String():
		importance = mail.ImportanceUrgent
	case "normal":
		importance = mail.ImportanceNormal
	default:
		return nil, ErrInvalidMailImportance
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Check for email conflicts
	var invitationConflict bool
	var memberConflict bool
	var ownerAddingSelf bool
	var selfEmail string
	if err := tx.QueryRowContext(ctx, `
		WITH invitation_conflict AS (
		    SELECT EXISTS (
		        SELECT 1 FROM staff_invitations
			 	WHERE user_email = $1
			  	AND school_id = $2
			 	AND created_at >= NOW() - INTERVAL '24 HOURS'	
		    ) AS exists
		),
		member_conflict AS (
		    SELECT EXISTS (
		        SELECT 1 FROM school_staff ss
			 	JOIN users u ON ss.user_id = u.id
		    	WHERE u.email = $1
		    	AND ss.school_id = $2
		    ) AS exists
		),
		adding_self AS (
		    SELECT EXISTS (
		        SELECT 1 FROM schools WHERE owner_id = $3 AND id = $2
		    ) AS exists
		),
		self_email AS (
		    SELECT email FROM users WHERE id = $3
		)
		SELECT
		    (SELECT exists FROM invitation_conflict),
		   (SELECT exists FROM member_conflict),
		   (SELECT exists FROM adding_self),
		   (SELECT email FROM self_email)
	`, p.Email, schoolID, userID).Scan(&invitationConflict, &memberConflict, &ownerAddingSelf, &selfEmail); err != nil {
		log.Println(err)
		return nil, err
	}

	if memberConflict {
		return nil, ErrStaffMemberEmailConflict
	}

	// If the owner is trying to add themselves, create the staff member instantly instead of sending an invitation.
	if ownerAddingSelf && selfEmail == p.Email {
		staffID, err := s.Sf.NextID()
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO school_staff (id, school_id, user_id, added_by_user, created_at)
			VALUES ($1, $2, $3, $3, NOW())
		`, staffID, schoolID, userID); err != nil {
			log.Println(err)
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			log.Println(err)
			return nil, err
		}

		return &SendInvitationResult{StaffCreated: true, NoRowsAffected: noRowsAffected}, nil
	}

	if invitationConflict {
		return nil, ErrInvitationEmailConflict
	}

	// Get the user's name and privacy settings
	var userName string
	var invitesOff bool
	if err := tx.QueryRowContext(ctx, `
		SELECT name, staff_invitations_disabled FROM users WHERE email = $1
	`, p.Email).Scan(&userName, &invitesOff); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		return nil, err
	}

	if invitesOff {
		return nil, ErrTargetPrivacyRestricted
	}

	// Generate invitation token
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	tokenHash := cryptoutil.MakeHash256(token)

	// Generate invitation ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Insert staff invitation row
	res, err := tx.ExecContext(ctx, `
		INSERT INTO staff_invitations (id, school_id, token_hash, user_email, sent_by_user)
		VALUES ($1, $2, $3, $4, $5)
	`, id, schoolID, tokenHash, p.Email, userID)
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
	if p.Importance == "" {
		p.Importance = "normal"
	}
	if err := schoollogs.StoreSchoolLog(schoolID, userID, schoollogs.ActionStaffInvitationCreate, schoollogs.TypeCreate, "Staff invitation sent", "{user} invited '"+p.Email+"' to join the school.", tx, ctx, s.Sf, "{user} sent a staff invitation to '"+p.Email+"' with "+p.Importance+" importance."); err != nil {
		log.Println(err)
		return nil, err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, err
	}

	// Send email
	verificationLink := fmt.Sprintf("%s/app/staff-invitations/%s", os.Getenv("FRONTEND_URL"), token)
	if userName != "" {
		userName = " " + userName
	}
	msg := fmt.Sprintf(`Hello%s,
You've been invited to join a school as a staff member. Click the link below to view the invitation:
%s`, userName, verificationLink)
	email := mailer.Mail{
		To:          p.Email,
		Subject:     "EduLink staff invitation",
		ContentType: mail.TypeTextPlain,
		Importance:  importance,
		Body:        msg,
	}

	go func() {
		if err := mailer.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	return &SendInvitationResult{NoRowsAffected: noRowsAffected}, nil
}
