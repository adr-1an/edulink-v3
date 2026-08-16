package portal

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/portal"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/netip"
	"strings"

	"github.com/alexedwards/argon2id"
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

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	var passwordHash string
	var loginAllowed bool
	if err := tx.QueryRowContext(ctx, `
		SELECT id, COALESCE(password_hash, ''), account_enabled FROM portal_users WHERE email = $1
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

	token, err := gonanoid.New(128)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenHash := helpers2.MakeHash256(token)

	var ptrForwardedFor *string
	var ptrSourceIP *string

	xForwardedFor := r.Header.Get("X-Forwarded-For")
	sourceIP := r.RemoteAddr

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

	userAgent := r.Header.Get("User-Agent")

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO portal_sessions (token_hash, portal_user_id, x_forwarded_for, source_ip, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, tokenHash, userID, ptrForwardedFor, ptrSourceIP, userAgent); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"token": token,
	})
}

func TokenCheckHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := r.Context()

	userID, err := portal.TokenToUID(w, r, db, ctx, helpers2.AccTypeEither)
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
	tokenHash := helpers2.MakeHash256(token)

	if _, err := db.ExecContext(ctx, `
		DELETE FROM portal_sessions WHERE token_hash = $1
	`, tokenHash); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
