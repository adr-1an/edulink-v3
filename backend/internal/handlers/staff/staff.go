package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	mail2 "net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
	"github.com/wneessen/go-mail"
)

func ListUserInvitationsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Invitation struct
	type User struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type School struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}

	type Invitation struct {
		ID        string    `json:"id"`
		School    School    `json:"school"`
		SentBy    User      `json:"sentBy"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	var invitations []Invitation

	// Get data from DB
	rows, err := db.QueryContext(ctx, `
		WITH uid AS (
		    SELECT email FROM users WHERE id = $1
		)
		SELECT
		    si.id, s.id, s.name, s.region_code, u.id, u.name,
		    u.email, si.status, si.created_at, si.expires_at
		FROM staff_invitations si
		JOIN schools s ON si.school_id = s.id
		JOIN users u ON si.sent_by_user = u.id
		WHERE si.user_email = (SELECT email FROM uid)
		AND si.status = 'pending'
		AND si.expires_at > NOW()
	`, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	// Scan rows
	for rows.Next() {
		var i Invitation

		if err := rows.Scan(
			&i.ID,
			&i.School.ID,
			&i.School.Name,
			&i.School.RegionCode,
			&i.SentBy.ID,
			&i.SentBy.Name,
			&i.SentBy.Email,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		invitations = append(invitations, i)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return invitations
	_ = json.NewEncoder(w).Encode(map[string]any{
		"invitations": invitations,
	})
}

func ListSchoolInvitationsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !schools.Can(schools.PermissionSchoolInviteList, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type invite struct {
		ID        string    `json:"id"`
		UserEmail string    `json:"userEmail"`
		SentBy    user      `json:"addedBy"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	var invites []invite

	rows, err := db.QueryContext(ctx, `
		SELECT
			u.id, u.name, u.email,
			i.id, i.user_email, i.status, i.created_at, i.expires_at
		FROM staff_invitations i
		JOIN users u
			ON u.id = i.sent_by_user
		WHERE i.school_id = $1
		AND status = 'pending'
	`, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var i invite

		if err := rows.Scan(
			&i.SentBy.ID,
			&i.SentBy.Name,
			&i.SentBy.Email,
			&i.ID,
			&i.UserEmail,
			&i.Status,
			&i.CreatedAt,
			&i.ExpiresAt,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		invites = append(invites, i)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":      access,
		"invitations": invites,
	})
}

func CancelSchoolInvitationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	invID, err := strconv.ParseInt(chi.URLParam(r, "invitationID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var schoolID int64
	var invEmail string
	if err := db.QueryRowContext(ctx, `
		SELECT school_id, user_email FROM staff_invitations WHERE id = $1
	`, invID).Scan(&schoolID, &invEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !schools.Can(schools.PermissionSchoolInviteCancel, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE staff_invitations
		SET status = 'rejected'
		WHERE id = $1
	`, invID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("{user} canceled the staff invitation sent to %s.", invEmail)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationCancel, schools.TypeEdit, "Staff invitation canceled", "{user} canceled a staff invitation.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ViewInvitationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// No auth required for this route, token is enough to view

	// Get token from URL & hash it
	token := chi.URLParam(r, "token")
	tokenHash := helpers2.MakeHash256(token)

	// Invitation struct
	type StaffInvitation struct {
		ID          string    `json:"id"`
		SentToEmail string    `json:"sentToEmail"`
		SchoolName  string    `json:"schoolName"`
		SentByName  string    `json:"sentByName"`
		SentByEmail string    `json:"sentByEmail"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"createdAt"`
		ExpiresAt   time.Time `json:"expiresAt"`
	}
	var inv StaffInvitation

	// Get invitation data
	if err := db.QueryRowContext(ctx, `
		SELECT
		    si.id, si.user_email, s.name, u.name, u.email, si.status, si.created_at, si.expires_at
		FROM staff_invitations si
		JOIN schools s ON s.id = si.school_id
		LEFT JOIN users u ON u.id = si.sent_by_user
		WHERE si.token_hash = $1
	`, tokenHash).
		Scan(
			&inv.ID,
			&inv.SentToEmail,
			&inv.SchoolName,
			&inv.SentByName,
			&inv.SentByEmail,
			&inv.Status,
			&inv.CreatedAt,
			&inv.ExpiresAt,
		); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Check expiry
	if inv.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusGone)
		return
	}

	// Return invitation
	_ = json.NewEncoder(w).Encode(map[string]any{
		"invitation": inv,
	})
}

func SendStaffInvitationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolIDStr := chi.URLParam(r, "schoolID")
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	allowed := schools.Can(schools.PermissionStaffCreate, userID, schoolID, ctx, db)
	if !allowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Payload
	type Payload struct {
		Email      string `json:"email"`
		Importance string `json:"importance"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse & validate
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.Importance = strings.ToLower(strings.TrimSpace(p.Importance))

	addr, err := mail2.ParseAddress(p.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidEmail,
		})
		return
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
	default:
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidMailImportance,
		})
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if memberConflict {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeStaffMemberEmailConflict,
		})
		return
	}

	// If the owner is trying to add themselves, create the staff member instantly instead of sending an invitation.
	if ownerAddingSelf && selfEmail == p.Email {
		staffID, err := sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO school_staff (id, school_id, user_id, added_by_user, created_at)
			VALUES ($1, $2, $3, $3, NOW())
		`, staffID, schoolID, userID); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	if invitationConflict {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvitationEmailConflict,
		})
		return
	}

	// Get the user's name and privacy settings
	var userName string
	var invitesOff bool
	if err := tx.QueryRowContext(ctx, `
		SELECT name, staff_invitations_disabled FROM users WHERE email = $1
	`, p.Email).Scan(&userName, &invitesOff); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if invitesOff {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeTargetPrivacyRestricted,
		})
		return
	}

	// Generate invitation token
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	// Generate invitation ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert staff invitation row
	res, err := tx.ExecContext(ctx, `
		INSERT INTO staff_invitations (id, school_id, token_hash, user_email, sent_by_user)
		VALUES ($1, $2, $3, $4, $5)
	`, id, schoolID, tokenHash, p.Email, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if p.Importance == "" {
		p.Importance = "normal"
	}
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationCreate, schools.TypeCreate, "Staff invitation sent", "{user} invited '"+p.Email+"' to join the school.", tx, ctx, sf, "{user} sent a staff invitation to '"+p.Email+"' with "+p.Importance+" importance."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send email
	verificationLink := fmt.Sprintf("%s/app/staff-invitations/%s", os.Getenv("FRONTEND_URL"), token)
	if userName != "" {
		userName = " " + userName
	}
	msg := fmt.Sprintf(`Hello%s,
You've been invited to join a school as a staff member. Click the link below to view the invitation:
%s`, userName, verificationLink)
	email := helpers2.Mail{
		To:          p.Email,
		Subject:     "EduLink staff invitation",
		ContentType: mail.TypeTextPlain,
		Importance:  importance,
		Body:        msg,
	}

	go func() {
		if err := helpers2.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}

func AcceptStaffInvitationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get token from URL & hash it
	token := chi.URLParam(r, "token")
	tokenHash := helpers2.MakeHash256(token)

	// Invitation struct
	type StaffInvitation struct {
		SchoolID   int64
		UserEmail  string
		SentByUser int64
	}
	var inv StaffInvitation

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Get user ID from email
	var staffUserID int64
	// Redundant ID check is intentional
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM users WHERE email = $1 AND id = $2
	`, inv.UserEmail, userID).Scan(&staffUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// The user is trying to accept an invitation which is not meant for their account.
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert staff member row
	res, err := tx.ExecContext(ctx, `
		INSERT INTO school_staff (id, school_id, user_id, added_by_user)
		VALUES ($1, $2, $3, $4)
	`, id, inv.SchoolID, staffUserID, inv.SentByUser)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Delete invitation
	res, err = tx.ExecContext(ctx, `
		DELETE FROM staff_invitations
		WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log invitation action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffInvitationDelete, schools.TypeDelete, "Staff invitation accepted", "{user} accepted a staff invitation.", tx, ctx, sf, "{user} accepted the staff invitation sent to '"+inv.UserEmail+"'."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffCreate, schools.TypeCreate, "Staff member added", "{user} joined the school staff_helpers.", tx, ctx, sf, "{user} was added as a staff member after accepting an invitation."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit TX
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RejectStaffInvitationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get invitation token & hash it
	token := chi.URLParam(r, "token")
	tokenHash := helpers2.MakeHash256(token)

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Compare IDs
	if userID != invUserID {
		// The user is trying to reject an invitation which is not meant for their account.
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Update invitation status to rejected
	res, err := tx.ExecContext(ctx, `
		UPDATE staff_invitations
		SET status = 'rejected'
		WHERE token_hash = $1
	`, tokenHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationEdit, schools.TypeEdit, "Staff invitation rejected", "{user} rejected a staff invitation.", tx, ctx, sf, "{user} rejected a staff invitation by token."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AcceptStaffInvitationByIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get invitation ID
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Invitation struct
	type StaffInvitation struct {
		SchoolID   int64
		SentByUser int64
	}
	var inv StaffInvitation

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert staff member row
	res, err := tx.ExecContext(ctx, `
		INSERT INTO school_staff (id, school_id, user_id, added_by_user)
		VALUES ($1, $2, $3, $4)
	`, id, inv.SchoolID, userID, inv.SentByUser)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Delete invitation
	res, err = tx.ExecContext(ctx, `
		DELETE FROM staff_invitations
		WHERE id = $1
	`, invitationID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log invitation action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffInvitationDelete, schools.TypeOther, "Staff invitation accepted", "{user} accepted a staff invitation.", tx, ctx, sf, "{user} accepted staff invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(inv.SchoolID, userID, schools.ActionStaffCreate, schools.TypeCreate, "Staff member added", "{user} joined the school staff_helpers.", tx, ctx, sf, "{user} was added as a staff member from invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit TX
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RejectStaffInvitationByIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get invitation ID
	invitationIDStr := chi.URLParam(r, "invitationID")
	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Compare IDs
	if userID != invUserID {
		// The user is trying to reject an invitation which is not meant for their account.
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Update invitation status to rejected
	res, err := tx.ExecContext(ctx, `
		UPDATE staff_invitations
		SET status = 'rejected'
		WHERE id = $1
	`, invitationID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	// Log action
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffInvitationEdit, schools.TypeOther, "Staff invitation rejected", "{user} rejected a staff invitation.", tx, ctx, sf, "{user} rejected staff invitation ID "+strconv.FormatInt(invitationID, 10)+"."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ListStaffMembersHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get school ID
	schoolIDStr := chi.URLParam(r, "schoolID")
	schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permissions
	if !schools.Can(schools.PermissionStaffView, userID, schoolID, ctx, db) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type UserSummary struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Email  string  `json:"email"`
		PfpURL *string `json:"profilePictureURL"`
	}

	type StaffRole struct {
		ID       string `json:"id"`
		Position int    `json:"position"`
		Name     string `json:"name"`
		Color    string `json:"color"`
	}

	type Staff struct {
		ID        string      `json:"id"`
		UserID    string      `json:"userId"`
		User      UserSummary `json:"user"`
		AddedBy   UserSummary `json:"addedBy"`
		CreatedAt time.Time   `json:"createdAt"`
		Roles     []StaffRole `json:"roles"`
	}

	// Get staff members
	rows, err := db.QueryContext(ctx, `
		SELECT
		    s.id,
		    s.user_id,
		    u.id, u.name, u.email,
		    added_by.id, added_by.name, added_by.email,
		    s.created_at,
		    sr.id,
		    sr.position,
		    sr.name,
		    sr.color,
		    
		    so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type
		FROM school_staff s
		JOIN users u ON s.user_id = u.id
		JOIN users added_by ON s.added_by_user = added_by.id
		LEFT JOIN staff_role_members srm
		ON srm.staff_id = s.id
		LEFT JOIN staff_roles sr
		ON srm.role_id = sr.id
		LEFT JOIN user_profile_pictures upp ON upp.user_id = u.id
		LEFT JOIN storage_objects so ON so.id = upp.storage_object_id AND status = $2
		WHERE s.school_id = $1
		ORDER BY s.created_at, s.id, sr.position DESC
	`, schoolID, helpers2.StatusDone)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = rows.Close() }()

	staffByID := make(map[string]*Staff)
	staffOrder := make([]string, 0)

	for rows.Next() {
		var (
			staffID   string
			userID    string
			user      UserSummary
			addedBy   UserSummary
			createdAt time.Time

			roleID       sql.NullString
			rolePosition sql.NullInt64
			roleName     sql.NullString
			roleColor    sql.NullString

			bucketName  *string
			objectKey   *string
			filename    *string
			contentType *string
		)

		if err := rows.Scan(
			&staffID,
			&userID,
			&user.ID,
			&user.Name,
			&user.Email,
			&addedBy.ID,
			&addedBy.Name,
			&addedBy.Email,
			&createdAt,
			&roleID,
			&rolePosition,
			&roleName,
			&roleColor,
			&bucketName,
			&objectKey,
			&filename,
			&contentType,
		); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		staff, exists := staffByID[staffID]
		if !exists {
			// Profile pic
			if bucketName != nil && objectKey != nil && filename != nil && contentType != nil {
				// Generate presigned URL
				url, err := s3.PresignedGetObject(ctx, *bucketName, *objectKey, 15*time.Minute, nil)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				pfpURL := url.String()
				user.PfpURL = &pfpURL
			}

			staff = &Staff{
				ID:        staffID,
				UserID:    userID,
				User:      user,
				AddedBy:   addedBy,
				CreatedAt: createdAt,
				Roles:     make([]StaffRole, 0),
			}

			staffByID[staffID] = staff
			staffOrder = append(staffOrder, staffID)
		}

		if roleID.Valid {
			staff.Roles = append(staff.Roles, StaffRole{
				ID:       roleID.String,
				Position: int(rolePosition.Int64),
				Name:     roleName.String,
				Color:    roleColor.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sMembers := make([]Staff, 0, len(staffOrder))
	for _, id := range staffOrder {
		sMembers = append(sMembers, *staffByID[id])
	}

	// Get user access
	access, err := schools.GetAllUserPermissions(ctx, db, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"staff":  sMembers,
	})
}

func DeleteStaffMemberHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	targetStaffID, err := strconv.ParseInt(chi.URLParam(r, "staffID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var schoolID int64
	var targetName string
	var targetEmail string

	err = tx.QueryRowContext(ctx, `
		SELECT ss.school_id, u.name, u.email
		FROM school_staff ss
		JOIN users u ON u.id = ss.user_id
		WHERE ss.id = $1
		FOR UPDATE
	`, targetStaffID).Scan(&schoolID, &targetName, &targetEmail)
	if errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var isOwner bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM schools WHERE id = $1 AND owner_id = $2)
	`, schoolID, userID).Scan(&isOwner); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !isOwner {
		var actorStaffID int64

		err = tx.QueryRowContext(ctx, `
			SELECT id
			FROM school_staff
			WHERE school_id = $1
			  AND user_id = $2
			FOR UPDATE
		`, schoolID, userID).Scan(&actorStaffID)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if actorStaffID == targetStaffID {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var targetHighestPos int

		err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(r.position), -1)
		FROM staff_role_members rm
		JOIN staff_roles r
			ON r.id = rm.role_id
		WHERE rm.staff_id = $1
		  AND r.school_id = $2
	`, targetStaffID, schoolID).Scan(&targetHighestPos)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var allowed bool

		err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM staff_role_members rm
			JOIN staff_roles r
				ON r.id = rm.role_id
			JOIN staff_role_permissions rp
				ON rp.role_id = r.id
			WHERE rm.staff_id = $1
			  AND r.school_id = $2
			  AND r.position > $3
			  AND rp.permission = $4
		)
		`,
			actorStaffID,
			schoolID,
			targetHighestPos,
			schools.PermissionStaffDelete,
		).Scan(&allowed)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM school_staff
		WHERE id = $1
		  AND school_id = $2
	`, targetStaffID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionStaffDelete, schools.TypeDelete, "Staff member removed", "{user} removed "+targetName+" from the staff_helpers.", tx, ctx, sf, "{user} removed staff member "+targetName+" <"+targetEmail+">."); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func LeaveSchoolStaffHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var staffRowExists bool
	var userName string
	var userEmail string
	if err := tx.QueryRowContext(ctx, `
		WITH staff_exists AS (
		    SELECT EXISTS (
				SELECT 1 FROM school_staff ss
				JOIN schools s ON ss.school_id = s.id
				WHERE ss.school_id = $2
				AND ss.user_id = $1
				AND s.owner_id <> $1
			) AS exists
		)
		SELECT staff_exists.exists, users.name, users.email
		FROM staff_exists
		JOIN users ON users.id = $1
		`, userID, schoolID).Scan(&staffRowExists, &userName, &userEmail); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !staffRowExists {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM school_staff WHERE user_id = $1 AND school_id = $2
	`, userID, schoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	details := fmt.Sprintf("%s <%s> left the school's staff members.", userName, userEmail)
	if err := schools.StoreSchoolLog(schoolID, userID, schools.ActionSchoolLeave, schools.TypeDelete, "User left", "{user} left the school staff_helpers.", tx, ctx, sf, details); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := staff_helpers.RowsAffectedOr500(res, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
