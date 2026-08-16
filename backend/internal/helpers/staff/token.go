package staff_helpers

import (
	"app/internal/helpers"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// TokenToUID is a handler helper function used to return the ID of the user who owns the token provided
// in the request.
func TokenToUID(w http.ResponseWriter, r *http.Request, db *sql.DB, ctx context.Context) (int64, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": ErrorCodeNoToken,
		})
		return 0, errors.New("missing token")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	tokenHash := helpers.MakeHash256(token)

	var uid int64
	err := db.QueryRowContext(ctx, `
		SELECT user_id FROM sessions
		WHERE token_hash = $1
		AND expires_at > NOW()
		AND revoked_at IS NULL
	`, tokenHash).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return 0, err
	}

	return uid, nil
}

// TokenToHash is a handler helper function used to get the token from the request and return
// it hashed (SHA256).
func TokenToHash(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": ErrorCodeNoToken,
		})
		return nil, errors.New("missing token")
	}
	token = strings.TrimPrefix(token, "Bearer ")

	return helpers.MakeHash256(token), nil
}
