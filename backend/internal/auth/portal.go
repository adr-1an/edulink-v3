// Package auth
package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"app/internal/cryptoutil"
	"app/internal/portalusers"
)

func ParseToken(w http.ResponseWriter, r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("missing authorization header")
	}

	token = strings.TrimPrefix(token, "Bearer ")

	return token, nil
}

func TokenToUID(w http.ResponseWriter, r *http.Request, db *sql.DB, ctx context.Context, accType portalusers.AccType) (int64, error) {
	token, err := ParseToken(w, r)
	if err != nil {
		return 0, err
	}

	tokenHash := cryptoutil.MakeHash256(token)

	var query string
	switch accType {
	case portalusers.AccTypeEither:
		query = `
SELECT ps.portal_user_id
FROM portal_sessions ps
JOIN portal_users pu
ON pu.id = ps.portal_user_id
WHERE ps.token_hash = $1
AND pu.account_type <> $2
AND ps.last_used_at > NOW() - INTERVAL '7 days'
AND pu.account_enabled = true
AND pu.account_active = true
`
	default:
		query = `
SELECT ps.portal_user_id
FROM portal_sessions ps
JOIN portal_users pu
ON pu.id = ps.portal_user_id
WHERE ps.token_hash = $1
AND pu.account_type = $2
AND ps.last_used_at > NOW() - INTERVAL '7 days'
AND pu.account_enabled = true
AND pu.account_active = true
`
	}

	var id int64
	if err := db.QueryRowContext(ctx, query, tokenHash, accType).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE portal_sessions SET last_used_at = NOW() WHERE token_hash = $1
	`, tokenHash); err != nil {
		log.Println(err)
		// don't return, no point
	}

	return id, nil
}
