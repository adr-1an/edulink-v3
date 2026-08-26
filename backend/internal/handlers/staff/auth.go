package staff

import (
	"app/internal/application/staff/auth"
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/go-chi/chi/v5"
	"github.com/matoous/go-nanoid/v2"
	"github.com/pquerna/otp/totp"
	"github.com/wneessen/go-mail"
)

// SendRegistrationLinkHandler generates a registration link, and if the provided email isn't taken,
// sends a registration email.
func (h *Handler) SendRegistrationLinkHandler(w http.ResponseWriter, r *http.Request) {
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

	err := h.App.Staff.Auth.SendRegistrationLink(ctx, &auth.SendRegistrationLinkInput{
		Email: p.Email,
	})

	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)

	case errors.Is(err, auth.ErrUnprocessableEntity):
		w.WriteHeader(http.StatusUnprocessableEntity)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// CheckRegistrationTokenHandler checks whether the given registration token is valid
// and returns the email associated with it.
func (h *Handler) CheckRegistrationTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")

	res, err := h.App.Staff.Auth.CheckRegistrationToken(ctx, &auth.CheckRegistrationTokenInput{
		Token: token,
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"email": res.Email,
	})
}

// RegistrationHandler - Main handler for user registration.
// Validation rules:
// Name - min 1, max 128
// Phone - min 3, max 32
// Password - min 8, no max
func (h *Handler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
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

	token := chi.URLParam(r, "token")

	err := h.App.Staff.Auth.Register(ctx, &auth.RegisterInput{
		Name:              p.Name,
		Phone:             p.Phone,
		Password:          p.Password,
		RegistrationToken: token,
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, auth.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginHandler - Main handler for user login.
// Validation rules:
// Email - min 5, max 254
// Password - min 8, no max
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.App.Staff.Auth.Login(ctx, &auth.LoginInput{
		Email:        p.Email,
		Password:     p.Password,
		StayLoggedIn: p.StayLoggedIn,
		IP:           ip,
		UserAgent:    userAgent,
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)

		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)
		}

		return
	}

	switch res.Type {
	case auth.LoginResultSuccess:
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": res.Token,
		})

	case auth.LoginResultTwoFactor:
		_ = json.NewEncoder(w).Encode(map[string]any{
			"twoFactorChallenge": map[string]any{
				"token":     res.TwoFactorChallenge.Token,
				"purpose":   res.TwoFactorChallenge.Purpose,
				"expiresAt": res.TwoFactorChallenge.ExpiresAt,
			},
		})
	}
}

// TokenCheckHandler checks whether the given token is valid and not expired.
// It will only return a 204 if the token is valid.
func (h *Handler) TokenCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token & hash
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeNoToken,
		})
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")

	valid, err := h.App.Staff.Auth.CheckToken(ctx, &auth.CheckTokenInput{
		Token: token,
	})

	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthorized):
			w.WriteHeader(http.StatusUnauthorized)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendPasswordResetEmailHandler sends a password reset link ("forgot password") to the user's email address.
func (h *Handler) SendPasswordResetEmailHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := h.DB.QueryRowContext(ctx, `
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
	_, err := h.DB.ExecContext(ctx, `
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
	_, err = h.DB.ExecContext(ctx, `
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
func (h *Handler) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
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
	err := h.DB.QueryRowContext(ctx, `
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
	_, err = h.DB.ExecContext(ctx, `
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
	_, err = h.DB.ExecContext(ctx, `
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
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tokenHash, err := staff_helpers.TokenToHash(w, r)
	if err != nil {
		return
	}

	// Delete token
	res, err := h.DB.ExecContext(ctx, `
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

func (h *Handler) Enable2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
	if err != nil {
		return
	}

	// Check if 2FA is already enabled
	var email string
	var status string
	if err := h.DB.QueryRowContext(ctx, `
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

	if status == auth.TwoFAStatusEnabled {
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
	if _, err := h.DB.ExecContext(ctx, `
		UPDATE users
		SET totp_secret = $1,
		    two_factor_status = $2
		WHERE id = $3
	`, encSecret, auth.TwoFAStatusPending, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"url": url,
	})
}

func (h *Handler) Verify2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
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

	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var totpSecret []byte
	if err := tx.QueryRowContext(ctx, `
		SELECT totp_secret FROM users WHERE id = $1 AND two_factor_status = $2
	`, userID, auth.TwoFAStatusPending).Scan(&totpSecret); err != nil {
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
	`, auth.TwoFAStatusEnabled, userID); err != nil {
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

func (h *Handler) Disable2faHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
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
	if err := h.DB.QueryRowContext(ctx, `
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
	if _, err := h.DB.ExecContext(ctx, `
		UPDATE users
		SET two_factor_status = $1,
		    totp_secret = NULL
		WHERE id = $2
	`, auth.TwoFAStatusDisabled, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete recovery codes
	if _, err := h.DB.ExecContext(ctx, `
		DELETE FROM two_factor_recovery_codes
		WHERE user_id = $1
	`, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RecoverTwoFactorHandler(w http.ResponseWriter, r *http.Request) {
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

	tx, err := h.DB.BeginTx(ctx, nil)
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
	`, auth.TwoFAStatusDisabled, userID, auth.TwoFAStatusEnabled); err != nil {
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
