package staff

import (
	helpers2 "app/internal/helpers"
	"app/internal/helpers/staff"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/sony/sonyflake/v2"
)

type TwoFactorChallengePurpose string

const (
	ChallengePurposeLogin          TwoFactorChallengePurpose = "login"
	ChallengePurposeSchoolDeletion TwoFactorChallengePurpose = "schoolDeletion"
)

func (p TwoFactorChallengePurpose) String() string {
	return string(p)
}

func verifyTfa(
	w http.ResponseWriter,
	db *sql.DB,
	ctx context.Context,
	userID int64,
	p challengeCompletionPayload,
	challengeID int64,
	tokenHash []byte,
) (bool, error) {
	var secretKey []byte
	if err := db.QueryRowContext(ctx, `
			SELECT totp_secret FROM users WHERE id = $1
		`, userID).Scan(&secretKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return false, err
	}

	appKey, err := helpers2.LoadAppEncKey()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}
	secret, err := helpers2.DecryptString(secretKey, appKey)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}

	if !totp.Validate(p.Code, secret) {
		w.WriteHeader(http.StatusUnauthorized)
		return false, err
	}

	var consumedID int64
	if err := db.QueryRowContext(ctx, `
		DELETE FROM two_factor_challenges
		WHERE id = $1
		AND token_hash = $2
		AND expires_at > NOW()
		RETURNING id
		`, challengeID, tokenHash).Scan(&consumedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return false, err
	}

	return true, nil
}

type challengeCompletionPayload struct {
	Code           string `json:"code"`
	ChallengeToken string `json:"challengeToken"`
}

func CompleteChallenge(w http.ResponseWriter, r *http.Request, db *sql.DB, sf *sonyflake.Sonyflake) {
	ctx := r.Context()
	var p challengeCompletionPayload

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p.Code = strings.TrimSpace(p.Code)

	if len(p.Code) != 6 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tokenHash := helpers2.MakeHash256(p.ChallengeToken)

	// Get challenge info
	var id int64
	var userID int64
	var purpose TwoFactorChallengePurpose
	var dbMetadata json.RawMessage
	var expiresAt time.Time

	if err := db.QueryRowContext(ctx, `
		SELECT id, user_id, purpose, metadata, expires_at
		FROM two_factor_challenges
		WHERE token_hash = $1
	`, tokenHash).Scan(&id, &userID, &purpose, &dbMetadata, &expiresAt); err != nil {
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

	if expiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": staff_helpers.ErrorCodeExpiredToken,
		})
		return
	}

	switch purpose {
	case ChallengePurposeLogin:
		var meta LoginChallengeMetadata

		if err := json.Unmarshal(dbMetadata, &meta); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ok, err := verifyTfa(w, db, ctx, userID, p, id, tokenHash)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		CompleteLogin(CompleteLoginPayload{
			W:   w,
			DB:  db,
			Ctx: ctx,
			Meta: LoginChallengeMetadata{
				UserID:       meta.UserID,
				StayLoggedIn: meta.StayLoggedIn,
				IP:           meta.IP,
				UserAgent:    meta.UserAgent,
			},
		})

	case ChallengePurposeSchoolDeletion:
		var meta SchoolDeletionChallengeMetadata

		if err := json.Unmarshal(dbMetadata, &meta); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ok, err := verifyTfa(w, db, ctx, userID, p, id, tokenHash)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		CompleteSchoolDeletion(CompleteSchoolDeletionPayload{
			W:   w,
			DB:  db,
			Sf:  sf,
			Ctx: ctx,
			Meta: SchoolDeletionChallengeMetadata{
				UserID:   meta.UserID,
				SchoolID: meta.SchoolID,
			},
		})

	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

// Challenge metadata types

type LoginChallengeMetadata struct {
	UserID       int64   `json:"userId"`
	StayLoggedIn bool    `json:"stayLoggedIn"`
	IP           *string `json:"ip"`
	UserAgent    string  `json:"userAgent"`
}

type SchoolDeletionChallengeMetadata struct {
	UserID   int64 `json:"userId"`
	SchoolID int64 `json:"schoolId"`
}
