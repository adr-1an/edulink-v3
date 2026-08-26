package staff

import (
	"app/internal/application/staff/auth"
	schoolService "app/internal/application/staff/schools"
	"app/internal/helpers"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (h *Handler) SchoolListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
	if err != nil {
		return
	}

	// Search params

	// showDeleted := r.URL.Query().Get("showDeleted") == "true"
	// Not in use for now

	res, err := h.App.Staff.Schools.ListSchools(ctx, &schoolService.ListSchoolsInput{
		UserID: userID,
	})

	type school struct {
		ID         string `json:"id"`
		OwnerID    string `json:"ownerId"`
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}
	var schoolList []school

	if res != nil {
		for _, s := range res.Schools {
			schoolList = append(schoolList, school{
				ID:         s.ID,
				OwnerID:    s.OwnerID,
				Name:       s.Name,
				RegionCode: s.RegionCode,
			})
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schools": schoolList,
	})
}

func (h *Handler) CreateSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
	if err != nil {
		return
	}

	// Payload
	type Payload struct {
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.Schools.CreateSchool(ctx, &schoolService.CreateSchoolInput{
		UserID: userID,
		School: schoolService.CreateSchool{
			Name:       p.Name,
			RegionCode: p.RegionCode,
		},
	})

	if err != nil {
		switch {
		case errors.Is(err, schoolService.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)

		case errors.Is(err, schoolService.ErrInvalidName):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidName,
			})

		case errors.Is(err, schoolService.ErrInvalidRegionCode):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidRegionCode,
			})

		case errors.Is(err, schoolService.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
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

	// Payload
	type Payload struct {
		Name                 string `json:"name"`
		RegionCode           string `json:"regionCode"`
		ActiveAcademicYearID *int64 `json:"activeAcademicYearId"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.App.Staff.Schools.UpdateSchool(ctx, &schoolService.UpdateSchoolInput{
		UserID:   userID,
		SchoolID: schoolID,
		School: schoolService.SchoolUpdate{
			Name:                 p.Name,
			RegionCode:           p.RegionCode,
			ActiveAcademicYearID: p.ActiveAcademicYearID,
		},
	})

	if err != nil {
		switch {
		case errors.Is(err, schoolService.ErrNoPermission), errors.Is(err, schoolService.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		case errors.Is(err, schoolService.ErrInvalidName):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidName,
			})

		case errors.Is(err, schoolService.ErrInvalidRegionCode):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staff_helpers.ErrorCodeInvalidRegionCode,
			})

		case errors.Is(err, schoolService.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authenticate
	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
	if err != nil {
		return
	}

	// Parse school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Permission check
	if !schools.Can(
		schools.SchoolOwner,
		userID,
		schoolID,
		ctx,
		h.DB,
	) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check whether 2FA is enabled
	var tfa string

	err = h.DB.QueryRowContext(ctx, `
		SELECT two_factor_status
		FROM users
		WHERE id = $1
	`, userID).Scan(&tfa)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if tfa == auth.TwoFAStatusEnabled {
		token, err := gonanoid.New(128)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenHash := helpers.MakeHash256(token)

		id, err := h.App.Staff.Auth.Sf.NextID()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		meta := auth.SchoolDeletionChallengeMetadata{
			UserID:   userID,
			SchoolID: schoolID,
		}

		metadata, err := json.Marshal(meta)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(15 * time.Minute)

		_, err = h.DB.ExecContext(ctx, `
			INSERT INTO two_factor_challenges (
				id,
				user_id,
				purpose,
				token_hash,
				expires_at,
				metadata
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			id,
			userID,
			auth.ChallengePurposeSchoolDeletion,
			tokenHash,
			expiresAt,
			metadata,
		)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"twoFactorChallenge": map[string]any{
				"token":     token,
				"purpose":   auth.ChallengePurposeSchoolDeletion,
				"expiresAt": expiresAt,
			},
		})

		return
	}

	// No 2FA required -> delete immediately
	err = h.App.Staff.Schools.DeleteSchool(
		ctx,
		&schoolService.DeleteSchoolInput{
			UserID:   userID,
			SchoolID: schoolID,
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ViewSchoolDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUID(w, r, h.DB, ctx)
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

	res, err := h.App.Staff.Schools.ViewSchool(ctx, &schoolService.SchoolViewInput{
		UserID:   userID,
		SchoolID: schoolID,
	})

	if err != nil {
		switch {
		case errors.Is(err, schoolService.ErrNoPermission):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"school": res.School,
	})
}
