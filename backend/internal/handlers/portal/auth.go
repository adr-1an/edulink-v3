package portal

import (
	"app/internal/helpers"
	"app/internal/helpers/portal"
	staffhelpers "app/internal/helpers/staff"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type payload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validatePayload(p payload) error {
	if p.Email == "" || len(p.Email) < 5 || len(p.Email) > 254 {
		return errors.New("invalid payload")
	}

	return nil
}

func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	var p payload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.Email = strings.TrimSpace(strings.ToLower(p.Email))

	if err := validatePayload(p); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var userID int64
	var passwordHash string
	var loginAllowed bool
	if err := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(password_hash, ''), account_enabled
		FROM portal_users
		WHERE email = $1 AND account_active = true
	`, p.Email).Scan(&userID, &passwordHash, &loginAllowed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !loginAllowed {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	match, err := argon2id.ComparePasswordAndHash(p.Password, passwordHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !match {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	CompleteLogin(completeLoginPayload{
		W:      w,
		R:      r,
		DB:     db,
		Ctx:    ctx,
		UserID: userID,
	})
}

type completeLoginPayload struct {
	W      http.ResponseWriter
	R      *http.Request
	DB     *sql.DB
	Ctx    context.Context
	UserID int64
}

func CompleteLogin(p completeLoginPayload) {
	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers.MakeHash256(token)

	var ptrForwardedFor *string
	var ptrSourceIP *string

	xForwardedFor := p.R.Header.Get("X-Forwarded-For")
	sourceIP := p.R.RemoteAddr

	if _, err := netip.ParseAddr(xForwardedFor); err != nil {
		xForwardedFor = ""
	}

	if _, err := netip.ParseAddr(sourceIP); err != nil {
		sourceIP = ""
	}

	if xForwardedFor != "" {
		ptrForwardedFor = &xForwardedFor
	}

	if sourceIP != "" {
		ptrSourceIP = &sourceIP
	}

	userAgent := p.R.Header.Get("User-Agent")

	if _, err := p.DB.ExecContext(p.Ctx, `
		INSERT INTO portal_sessions (token_hash, portal_user_id, x_forwarded_for, source_ip, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, tokenHash, p.UserID, ptrForwardedFor, ptrSourceIP, userAgent); err != nil {
		log.Println(err)
		p.W.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(p.W).Encode(map[string]any{
		"token": token,
	})
}

func TokenCheckHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, db, ctx, helpers.AccTypeEither)
	if err != nil {
		return
	}

	var accType string
	if err := db.QueryRowContext(ctx, `
		SELECT account_type FROM portal_users WHERE id = $1
	`, userID).Scan(&accType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": map[string]any{
			"id":          userID,
			"accountType": accType,
		},
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	token, err := portal.ParseToken(w, r)
	if err != nil {
		return
	}
	tokenHash := helpers.MakeHash256(token)

	if _, err := db.ExecContext(ctx, `
		DELETE FROM portal_sessions WHERE token_hash = $1
	`, tokenHash); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func AccountActivationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	activationToken := chi.URLParam(r, "token")
	if activationToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorCodeNoToken,
		})
		return
	}
	tokenHash := helpers.MakeHash256(activationToken)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Get user ID by token hash
	var userID int64
	var expiresAt time.Time
	if err := tx.QueryRowContext(ctx, `
		DELETE FROM portal_account_activation_tokens
		WHERE token_hash = $1
		RETURNING portal_user_id, expires_at
	`, tokenHash).Scan(&userID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staffhelpers.ErrorCodeInvalidToken,
			})
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if expiresAt.Before(time.Now()) || expiresAt.Equal(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staffhelpers.ErrorCodeExpiredToken,
		})
		return
	}

	type payload struct {
		NewPassword string `json:"newPassword"`
	}
	var p payload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(p.NewPassword) < 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Hash new password
	passHash, err := argon2id.CreateHash(p.NewPassword, argon2id.DefaultParams)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Update user
	if _, err := tx.ExecContext(ctx, `
		UPDATE portal_users
		SET
		    password_hash = $1,
		    account_active = true
		WHERE id = $2
	`, passHash, userID); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Login
	CompleteLogin(completeLoginPayload{
		W:      w,
		R:      r,
		DB:     db,
		Ctx:    ctx,
		UserID: userID,
	})
}
