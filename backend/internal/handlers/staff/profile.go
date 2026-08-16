package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-chi/chi/v5"
	"github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
	"github.com/wneessen/go-mail"
)

func GetProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Get user data
	type pfp struct {
		PresignedURL string `json:"presignedUrl"`
		Filename     string `json:"fileName"`
		ContentType  string `json:"contentType"`
	}

	type User struct {
		Pfp             *pfp      `json:"profilePicture"`
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		Phone           string    `json:"phone"`
		TwoFactorStatus string    `json:"twoFactorStatus"`
		PublicProfile   bool      `json:"publicProfile"`
		InvDisabled     bool      `json:"staffInvitationsDisabled"`
		UpdatedAt       time.Time `json:"updatedAt"`
		CreatedAt       time.Time `json:"createdAt"`
	}
	var user User

	var bucketName sql.NullString
	var objKey sql.NullString
	var filename sql.NullString
	var contentType sql.NullString

	if err = db.QueryRowContext(ctx, `
		SELECT
		    so.bucket_name, so.object_key, so.original_file_name, so.declared_content_type,
		    u.id, u.name, u.email, u.phone, u.two_factor_status, u.public_profile, u.staff_invitations_disabled, u.updated_at, u.created_at
		FROM users u
		LEFT JOIN user_profile_pictures pfps
		ON u.id = pfps.user_id
		LEFT JOIN storage_objects so
		ON so.id = pfps.storage_object_id
		WHERE u.id = $1
		`, userID).
		Scan(
			&bucketName,
			&objKey,
			&filename,
			&contentType,
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.TwoFactorStatus,
			&user.PublicProfile,
			&user.InvDisabled,
			&user.UpdatedAt,
			&user.CreatedAt,
		); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if filename.Valid && contentType.Valid && bucketName.Valid {
		url, err := s3.PresignedGetObject(ctx, bucketName.String, objKey.String, 15*time.Minute, nil)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user.Pfp = &pfp{
			PresignedURL: url.String(),
			Filename:     filename.String,
			ContentType:  contentType.String,
		}
	}

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": user,
	})
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Name          string `json:"name"`
		Phone         string `json:"phone"`
		PublicProfile bool   `json:"publicProfile"`
		InvDisabled   bool   `json:"staffInvitationsDisabled"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)

	// Validate
	if len(p.Name) > 64 || (len(p.Phone) < 3 && p.Phone != "") || len(p.Phone) > 64 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Update
	if _, err = db.ExecContext(ctx, `
		UPDATE users
		SET name = $1, phone = $2, public_profile = $3, staff_invitations_disabled = $4, updated_at = NOW()
		WHERE id = $5
	`, p.Name, p.Phone, p.PublicProfile, p.InvDisabled, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SendEmailChangeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		NewEmail string `json:"newEmail"`
		Password string `json:"password"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.NewEmail = strings.TrimSpace(strings.ToLower(p.NewEmail))

	// Validate
	if p.NewEmail == "" || len(p.NewEmail) < 5 || len(p.NewEmail) > 254 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Get current password hash & check for email conflict
	var passwordHash string
	var emailConflict bool
	err = db.QueryRowContext(ctx, `
		WITH conflict AS (
		    SELECT EXISTS (SELECT 1 FROM users WHERE email = $2) AS email_exists
		)
		SELECT u.password_hash, c.email_exists
		FROM users u, conflict c
		WHERE u.id = $1
	`, userID, p.NewEmail).Scan(&passwordHash, &emailConflict)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if emailConflict {
		// Fake 204
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Compare passwords
	match, err := argon2id.ComparePasswordAndHash(p.Password, passwordHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeIncorrectPassword,
		})
		return
	}

	// Generate verification token
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	// Store token
	var userEmail string
	if err := db.QueryRowContext(ctx, `
		WITH inserted AS (
		    INSERT INTO verification_tokens (token_hash, user_id, purpose, email_change_new_email)
		   	VALUES ($1, $2, $3, $4)
		    RETURNING user_id
		)
		SELECT u.email
		FROM users u
		JOIN inserted i ON u.id = i.user_id
	`, tokenHash, userID, staff_helpers.TokenPurposeEmailChange, p.NewEmail).Scan(&userEmail); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send verification email
	verURL := fmt.Sprintf("%s/auth/verify-change/%s", os.Getenv("FRONTEND_URL"), token)
	msg := fmt.Sprintf("Click this link to finish changing your email: %s", verURL)
	email := helpers2.Mail{
		To:          p.NewEmail,
		Subject:     "Verify your Email",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceHigh,
		Body:        msg,
	}
	if err := helpers2.SendMail(email); err != nil {
		log.Println(err)
		// Don't return - token was already stored
	}

	w.WriteHeader(http.StatusNoContent)
}

func EmailUpdateHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get token from URL
	token := chi.URLParam(r, "token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	// Get old email
	var oldEmail string
	var newEmail string
	if err := db.QueryRowContext(ctx, `
		SELECT u.email, vt.email_change_new_email
		FROM verification_tokens vt
		JOIN users u ON u.id = vt.user_id
		WHERE token_hash = $1
	`, tokenHash).Scan(&oldEmail, &newEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Delete token & update user
	var res string
	err := db.QueryRowContext(ctx, `
		WITH token AS (
		    SELECT user_id, email_change_new_email
		    FROM verification_tokens
		    WHERE token_hash = $1
				AND purpose = $2
				AND created_at >= NOW() - INTERVAL '24 hours'
		),
		email_taken AS (
		    SELECT 1
		    FROM users u
		    JOIN token t ON true
		    WHERE u.email = t.email_change_new_email
		    	AND u.id <> t.user_id
		),
		deleted AS (
		    DELETE FROM verification_tokens
			WHERE token_hash = $1
			  	AND purpose = $2
		        AND EXISTS (SELECT 1 FROM token)
		        AND NOT EXISTS (SELECT 1 FROM email_taken)
		    RETURNING user_id
		),
		updated AS (
		    UPDATE users u
		   	SET email = t.email_change_new_email
		   	FROM token t
	   		JOIN deleted d ON d.user_id = t.user_id
		   	RETURNING 1
		)
		SELECT CASE 
			WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
			WHEN EXISTS (SELECT 1 FROM email_taken) THEN 'email_taken'
			ELSE 'invalid_token'
		END AS result
	`, tokenHash, staff_helpers.TokenPurposeEmailChange).Scan(&res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Match cases
	switch res {
	case "updated":
		w.WriteHeader(http.StatusNoContent)
	case "email_taken":
		w.WriteHeader(http.StatusConflict)
	case "invalid_token":
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if res == "updated" {
		msg := fmt.Sprintf(`Your account email has been changed.
Old email: %s
New email: %s`, oldEmail, newEmail)
		emailOld := helpers2.Mail{
			To:          oldEmail,
			Subject:     "Email changed",
			ContentType: mail.TypeTextPlain,
			Importance:  mail.ImportanceNormal,
			Body:        msg,
		}
		emailNew := helpers2.Mail{
			To:          newEmail,
			Subject:     "Email changed",
			ContentType: mail.TypeTextPlain,
			Importance:  mail.ImportanceNormal,
			Body:        msg,
		}

		go func() {
			if err := helpers2.SendMail(emailOld); err != nil {
				log.Println(err)
			}
			if err := helpers2.SendMail(emailNew); err != nil {
				log.Println(err)
			}
		}()
	}

	return
}

func PasswordChangeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Password    string `json:"password"`
		NewPassword string `json:"newPassword"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate
	if len(p.NewPassword) < 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Compare passwords
	var uName string
	var uEmail string
	var currentHash string
	if err := db.QueryRowContext(ctx, `
		SELECT name, email, password_hash FROM users WHERE id = $1
	`, userID).Scan(&uName, &uEmail, &currentHash); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(p.Password, currentHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Delete all sessions
	if _, err := db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW(), revoke_note = 'Password change'
		WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate hash
	newHash, err := argon2id.CreateHash(p.NewPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Change password
	if _, err := db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, newHash, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send email notif
	appName := os.Getenv("APP_NAME")
	msg := fmt.Sprintf(`Hello %s,

Your %s account password has been changed.
If this wasn't you, review your account security immediately.
`, uName, appName)
	email := helpers2.Mail{
		To:          uEmail,
		Subject:     "Password changed",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceNormal,
		Body:        msg,
	}
	go func() {
		if err := helpers2.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}

func UploadPfpHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	type payload struct {
		Name                string `json:"name"`
		DeclaredSize        int64  `json:"declaredSize"`
		DeclaredContentType string `json:"declaredContentType"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.DeclaredSize <= 0 || p.DeclaredSize > 5*1024*1024 /* 5MB */ {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
		"image/gif":  true,
	}

	if !allowed[p.DeclaredContentType] {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		log.Println("S3_BUCKET environment variable not set.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s := helpers2.UploadService{
		Db:         db,
		S3:         s3,
		Ctx:        ctx,
		BucketName: bucketName,
	}

	id, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pfpID, err := sf.NextID()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completionToken, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(completionToken)

	objKey := fmt.Sprintf("%s/%d", helpers2.UploadCategoryUserProfilePics, id)

	if err := s.StoreStorageObjectsRow(id, tokenHash, userID, p.Name, p.DeclaredSize, p.DeclaredContentType, objKey); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO user_profile_pictures (id, user_id, storage_object_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		SET storage_object_id = EXCLUDED.storage_object_id
	`, pfpID, userID, id); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url, err := s3.PresignedPutObject(ctx, bucketName, objKey, 5*time.Minute)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":              strconv.FormatInt(id, 10),
		"completionToken": completionToken,
		"url":             url.String(),
	})
}

func ClearPfpHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, s3 *minio.Client) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// delete from db
	var objKey string
	var bucketName string
	if err = db.QueryRowContext(ctx, `
		WITH oid AS (
		    DELETE FROM user_profile_pictures WHERE user_id = $1 RETURNING storage_object_id
		)
		DELETE FROM storage_objects so USING oid WHERE id = oid.storage_object_id RETURNING so.object_key, so.bucket_name
	`, userID).Scan(&objKey, &bucketName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// delete from s3
	if err := s3.RemoveObject(ctx, bucketName, objKey, minio.RemoveObjectOptions{}); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
