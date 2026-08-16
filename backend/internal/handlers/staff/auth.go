package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-chi/chi/v5"
	"github.com/matoous/go-nanoid/v2"
	"github.com/pquerna/otp/totp"
	"github.com/sony/sonyflake/v2"
	"github.com/wneessen/go-mail"
)

// SendRegistrationLinkHandler generates a registration link, and if the provided email isn't taken,
// sends a registration email.
func SendRegistrationLinkHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Email string `json:"email"`
	}
	var p Payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))

	// Validate
	if p.Email == "" || len(p.Email) < 5 || len(p.Email) > 254 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fmt.Println(err)
		}
	}()

	// Check for email conflict
	var emailConflict bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM users u
		 	WHERE u.email = $1
		)
	`, p.Email).Scan(&emailConflict); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if emailConflict {
		// Email is already taken, fake ok response
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check if an active token already exists
	var validTokenExists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM registration_tokens rt
		 	WHERE rt.email = $1
		 	AND rt.expires_at > NOW()
		)
	`, p.Email).Scan(&validTokenExists); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if validTokenExists {
		// A valid token already exists, fake ok response
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Generate token & hash
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	// Store registration token
	res, err := db.ExecContext(ctx, `
		INSERT INTO registration_tokens (token_hash, email, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
		ON CONFLICT (email) DO UPDATE
		    SET
		        token_hash = EXCLUDED.token_hash,
		        expires_at = EXCLUDED.expires_at,
		        created_at = NOW()
		WHERE registration_tokens.expires_at <= NOW()
	`, tokenHash, p.Email)
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
		// Valid token with this email exists, fake ok response
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Send email
	frontendURL := os.Getenv("FRONTEND_URL")
	appName := os.Getenv("APP_NAME")
	registrationURL := fmt.Sprintf("%s/auth/register/%s", frontendURL, token)
	msg := fmt.Sprintf("Hi! Let's complete your %s registration: %s", appName, registrationURL)
	email := helpers2.Mail{
		To:          p.Email,
		Subject:     "Complete your registration",
		Body:        msg,
		Importance:  mail.ImportanceHigh,
		ContentType: mail.TypeTextPlain,
	}

	// Send email
	go func() {
		if err := helpers2.SendMail(email); err != nil {
			log.Println(err)
		}
	}()

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckRegistrationTokenHandler checks whether the given registration token is valid
// and returns the email associated with it.
func CheckRegistrationTokenHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	tokenHash := helpers2.MakeHash256(token)

	// Get the email from the token
	var email string
	if err := db.QueryRowContext(ctx, `
		SELECT email
		FROM registration_tokens
		WHERE token_hash = $1
		AND expires_at > NOW()
	`, tokenHash).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"email": email,
	})
}

// RegistrationHandler - Main handler for user registration.
// Validation rules:
// Name - min 1, max 128
// Phone - min 3, max 32
// Password - min 8, no max
func RegistrationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)

	// Validate

	// Name
	if p.Name == "" ||
		len(p.Name) < 1 ||
		len(p.Name) > 128 ||
		// Password
		p.Password == "" ||
		len(p.Password) < 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Phone
	if p.Phone != "" {
		if len(p.Phone) < 3 || len(p.Phone) > 32 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	}

	// Get token & hash it
	token := chi.URLParam(r, "token")
	tokenHash := helpers2.MakeHash256(token)

	// Start tx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() {
		err = tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Println(err)
		}
	}()

	// Get email from token
	var email string
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM registration_tokens
		WHERE token_hash = $1
		AND expires_at > NOW()
		RETURNING email
	`, tokenHash).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Generate ID
	id, err := sf.NextID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Hash password
	passwordHash, err := argon2id.CreateHash(p.Password, argon2id.DefaultParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Store user
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, name, email, phone, password_hash)
		VALUES ($1, $2, $3, $4, $5)
	`, id, p.Name, email, p.Phone, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Commit tx
	if err = tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginHandler - Main handler for user login.
// Validation rules:
// Email - min 5, max 254
// Password - min 8, no max
func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()

	// Get user agent & IP
	userAgent := r.Header.Get("User-Agent")
	forwardedFor := r.Header.Get("X-Forwarded-For")
	var ip *string
	if forwardedFor != "" {
		ip = &forwardedFor
	}

	// Payload
	type Payload struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		StayLoggedIn bool   `json:"stayLoggedIn"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Email = strings.TrimSpace(strings.ToLower(p.Email))

	// Validate
	if p.Email == "" || len(p.Email) < 5 || len(p.Email) > 254 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Get user password
	var userID int64
	var passwordHash string
	err := db.QueryRowContext(ctx, `
		SELECT id, password_hash FROM users WHERE email = $1
	`, p.Email).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	// Compare password and hash
	match, err := argon2id.ComparePasswordAndHash(p.Password, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the account requires a 2FA challenge
	var tfa string
	if err := db.QueryRowContext(ctx, `
		SELECT two_factor_status FROM users WHERE id = $1
	`, userID).Scan(&tfa); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if tfa == TwoFAStatusEnabled {
		// Generate challenge token
		token, err := gonanoid.New(128)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tokenHash := helpers2.MakeHash256(token)

		id, err := sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		meta := LoginChallengeMetadata{
			UserID:       userID,
			StayLoggedIn: p.StayLoggedIn,
			IP:           ip,
			UserAgent:    userAgent,
		}

		metadata, err := json.Marshal(meta)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(15 * time.Minute)
		if _, err := db.ExecContext(ctx, `
			INSERT INTO two_factor_challenges (id, user_id, purpose, token_hash, expires_at, metadata)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, id, userID, ChallengePurposeLogin, tokenHash, expiresAt, metadata); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return challenge token, purpose and expiry
		_ = json.NewEncoder(w).Encode(map[string]any{
			"twoFactorChallenge": map[string]any{
				"token":     token,
				"purpose":   ChallengePurposeLogin,
				"expiresAt": expiresAt,
			},
		})
		return
	}

	// If not, continue
	CompleteLogin(CompleteLoginPayload{
		W:   w,
		DB:  db,
		Ctx: ctx,
		Meta: LoginChallengeMetadata{
			UserID:       userID,
			StayLoggedIn: p.StayLoggedIn,
			IP:           ip,
			UserAgent:    userAgent,
		},
	})
}

type CompleteLoginPayload struct {
	W    http.ResponseWriter
	DB   *sql.DB
	Ctx  context.Context
	Meta LoginChallengeMetadata
}

func CompleteLogin(p CompleteLoginPayload) {
	// Generate session token & hash
	sessionToken, err := gonanoid.New(128)
	if err != nil {
		p.W.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	sessionTokenHash := helpers2.MakeHash256(sessionToken)

	// Store session token
	// If stayLoggedIn == true, expire session in 3 months
	// If stayLoggedIn == false, expire session in 24 hours
	now := time.Now()
	expiresAt := now
	if p.Meta.StayLoggedIn {
		// Expires in 3 months
		expiresAt = now.AddDate(0, 3, 0)
	} else {
		// Expires in 1 day
		expiresAt = now.AddDate(0, 0, 1)
	}

	_, err = p.DB.ExecContext(p.Ctx, `
		INSERT INTO sessions (token_hash, user_id, created_from_ip, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, sessionTokenHash, p.Meta.UserID, p.Meta.IP, p.Meta.UserAgent, expiresAt)
	if err != nil {
		p.W.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Return the session token
	p.W.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(p.W).Encode(map[string]any{
		"token": sessionToken,
	})
}

// TokenCheckHandler checks whether the given token is valid and not expired.
// It will only return a 204 if the token is valid.
func TokenCheckHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get token & hash
	tokenHash, err := staff_helpers.TokenToHash(w, r)
	if err != nil {
		return
	}

	// Check if token is valid
	var exists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM sessions
			WHERE token_hash = $1
			AND expires_at > NOW()
			AND revoked_at IS NULL
		)
	`, tokenHash).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidToken,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendPasswordResetEmailHandler sends a password reset link ("forgot password") to the user's email address.
func SendPasswordResetEmailHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Payload
	type Payload struct {
		Email string `json:"email"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse
	p.Email = strings.TrimSpace(strings.ToLower(p.Email))

	// Validate
	if len(p.Email) < 5 || len(p.Email) > 254 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidEmail,
		})
		return
	}

	// Get ID from email
	var userID int64
	if err := db.QueryRowContext(ctx, `
		SELECT id FROM users WHERE email = $1
	`, p.Email).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Fake 204
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	// Delete all previous password reset tokens
	_, err := db.ExecContext(ctx, `
		DELETE FROM verification_tokens
		WHERE user_id = $1
		AND purpose = $2
	`, userID, staff_helpers.TokenPurposePasswordReset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Generate token & hash
	resetToken, err := gonanoid.New(128)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	resetTokenHash := helpers2.MakeHash256(resetToken)

	// Store in DB
	_, err = db.ExecContext(ctx, `
		INSERT INTO verification_tokens (token_hash, user_id, purpose)
		VALUES ($1, $2, $3)
	`, resetTokenHash, userID, staff_helpers.TokenPurposePasswordReset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Send email
	frontendUrl := os.Getenv("FRONTEND_URL")
	resetURL := fmt.Sprintf("%s/auth/reset/%s", frontendUrl, resetToken)

	msgBody := fmt.Sprintf("Did you request a password reset? If so, here's your link: %s", resetURL)
	msg := helpers2.Mail{
		To:          p.Email,
		Subject:     "Password reset request",
		ContentType: mail.TypeTextPlain,
		Importance:  mail.ImportanceHigh,
		Body:        msgBody,
	}
	go func() {
		if err := helpers2.SendMail(msg); err != nil {
			log.Println(err)
			return
		}
	}()

	w.WriteHeader(http.StatusNoContent)
}

// PasswordResetHandler resets the user's password by using a reset token.
func PasswordResetHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	// Get reset token
	token := chi.URLParam(r, "token")
	if token == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeNoToken,
		})
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	// Payload
	type Payload struct {
		NewPassword string `json:"newPassword"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate
	if len(p.NewPassword) < 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Get ID
	var userID int64
	err := db.QueryRowContext(ctx, `
		DELETE FROM verification_tokens
	   		WHERE token_hash = $1
			AND created_at >= NOW() - INTERVAL '24 hours'
	   		AND purpose = $2
		RETURNING user_id
	`, tokenHash, staff_helpers.TokenPurposePasswordReset).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidToken,
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	// Revoke all sessions
	_, err = db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE user_id = $1
		AND revoked_at IS NULL
	`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Hash the new password
	newPasswdHash, err := argon2id.CreateHash(p.NewPassword, argon2id.DefaultParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Update password
	_, err = db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
	`, newPasswdHash, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LogoutHandler revokes the current token sent in the request.
func LogoutHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	tokenHash, err := staff_helpers.TokenToHash(w, r)
	if err != nil {
		return
	}

	// Delete token
	res, err := db.ExecContext(ctx, `
		UPDATE sessions
		SET revoked_at = NOW()
		WHERE token_hash = $1
		AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// Check rows affected
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeInvalidToken,
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

const (
	TwoFAStatusEnabled  string = "enabled"
	TwoFAStatusPending  string = "pending"
	TwoFAStatusDisabled string = "disabled"
)

func Enable2faHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	// Check if 2FA is already enabled
	var email string
	var status string
	if err := db.QueryRowContext(ctx, `
		SELECT email, two_factor_status FROM users WHERE id = $1
	`, userID).Scan(&email, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if status == TwoFAStatusEnabled {
		w.WriteHeader(http.StatusConflict)
		return
	}

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		log.Println("APP_NAME environment variable not set.")
		appName = "-"
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: email,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	secret := key.Secret()
	url := key.URL()

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Encrypt secret
	encSecret, err := helpers2.EncryptString(secret, appKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store secret
	if _, err := db.ExecContext(ctx, `
		UPDATE users
		SET totp_secret = $1,
		    two_factor_status = $2
		WHERE id = $3
	`, encSecret, TwoFAStatusPending, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"url": url,
	})
}

func Verify2faHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	type payload struct {
		TOTP string `json:"code"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.TOTP = strings.TrimSpace(p.TOTP)

	if len(p.TOTP) != 6 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var totpSecret []byte
	if err := tx.QueryRowContext(ctx, `
		SELECT totp_secret FROM users WHERE id = $1 AND two_factor_status = $2
	`, userID, TwoFAStatusPending).Scan(&totpSecret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	secret, err := helpers2.DecryptString(totpSecret, appKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !totp.Validate(p.TOTP, secret) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET two_factor_status = $1
		WHERE id = $2
	`, TwoFAStatusEnabled, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	codes := make(map[int]string)
	values := make([]string, 0, 8)
	args := make([]any, 0, 16)

	for i := 0; i < 8; i++ {
		code, err := helpers2.GenerateRecoveryCode()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Println(code)

		codeHash := helpers2.MakeHash256(code)
		codes[i] = code

		n := len(args)
		values = append(values, fmt.Sprintf("($%d, $%d)", n+1, n+2))
		args = append(args, codeHash, userID)
	}

	// Workaround to suppress false IDE SQL parsing errors
	query := `
	INSERT INTO two_factor_recovery_codes
		(recovery_code_hash, user_id)
	VALUES ($0, $0)
`
	query = strings.ReplaceAll(query, "($0, $0)", strings.Join(values, ","))

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"codes": codes,
	})
}

func Disable2faHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, db, ctx)
	if err != nil {
		return
	}

	type payload struct {
		TOTP     string `json:"code"`
		Password string `json:"password"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.TOTP = strings.TrimSpace(p.TOTP)

	if len(p.TOTP) != 6 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Get password hash
	var passHash string
	var totpSecret []byte
	if err := db.QueryRowContext(ctx, `
		SELECT password_hash, totp_secret FROM users WHERE id = $1
	`, userID).Scan(&passHash, &totpSecret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	match, err := argon2id.ComparePasswordAndHash(p.Password, passHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check TOTP
	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	secret, err := helpers2.DecryptString(totpSecret, appKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !totp.Validate(p.TOTP, secret) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Disable 2FA
	if _, err := db.ExecContext(ctx, `
		UPDATE users
		SET two_factor_status = $1,
		    totp_secret = NULL
		WHERE id = $2
	`, TwoFAStatusDisabled, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete recovery codes
	if _, err := db.ExecContext(ctx, `
		DELETE FROM two_factor_recovery_codes
		WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RecoverTwoFactorHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	type payload struct {
		RecoveryCode string `json:"recoveryCode"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	codeHash := helpers2.MakeHash256(p.RecoveryCode)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT user_id FROM two_factor_recovery_codes WHERE recovery_code_hash = $1
	`, codeHash).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidRecoveryCode,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Disable 2FA for the user
	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET totp_secret = NULL, two_factor_status = $1
		WHERE id = $2
		AND two_factor_status = $3
	`, TwoFAStatusDisabled, userID, TwoFAStatusEnabled); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete all the user's recovery codes
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM two_factor_recovery_codes WHERE user_id = $1
	`, userID); err != nil {
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
