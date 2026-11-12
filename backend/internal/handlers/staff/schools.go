package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	authService "app/internal/application/staff/auth"
	schoolService "app/internal/application/staff/schools"
	"app/internal/auth"
	"app/internal/responsecodes"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListSchools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Search params

	// showDeleted := r.URL.Query().Get("showDeleted") == "true"
	// Not in use for now

	res, err := h.App.Staff.Schools.ListSchools(ctx, &schoolService.ListSchoolsInput{
		UserID: userID,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type school struct {
		ID         string `json:"id"`
		OwnerID    string `json:"ownerId"`
		Name       string `json:"name"`
		RegionCode string `json:"regionCode"`
	}
	var schoolList []school

	if res != nil {
		for _, s := range res.Schools {
			idStr := strconv.FormatInt(s.ID, 10)
			ownerIDStr := strconv.FormatInt(s.OwnerID, 10)

			schoolList = append(schoolList, school{
				ID:         idStr,
				OwnerID:    ownerIDStr,
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

func (h *Handler) CreateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
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
				"code": responsecodes.ErrorCodeInvalidName,
			})

		case errors.Is(err, schoolService.ErrInvalidRegionCode):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": responsecodes.ErrorCodeInvalidRegionCode,
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

func (h *Handler) UpdateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
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
				"code": responsecodes.ErrorCodeInvalidName,
			})

		case errors.Is(err, schoolService.ErrInvalidRegionCode):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": responsecodes.ErrorCodeInvalidRegionCode,
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

func (h *Handler) DeleteSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Authenticate
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Parse school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Auth.RequestSchoolDeletion(ctx, &authService.RequestSchoolDeletionInput{UserID: userID, SchoolID: schoolID})
	if err != nil {
		if errors.Is(err, authService.ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.Challenge != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"twoFactorChallenge": map[string]any{"token": res.Challenge.Token, "purpose": res.Challenge.Purpose, "expiresAt": res.Challenge.ExpiresAt}})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ViewSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
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
		"access": res.Access,
		"school": res.School,
	})
}
