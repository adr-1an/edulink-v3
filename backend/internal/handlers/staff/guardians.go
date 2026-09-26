package staff

import (
	"app/internal/application/staff/guardians"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type guardiansPayload struct {
	Name           string    `json:"name"`
	LastName       string    `json:"lastName"`
	Email          string    `json:"email"`
	Phone          *string   `json:"phone"`
	Notes          *string   `json:"notes"`
	DateOfBirth    time.Time `json:"dateOfBirth"`
	AccountEnabled bool      `json:"accountEnabled"`
	Password       *string   `json:"password"`
}

func (h *Handler) CreateGuardianHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staff_helpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Payload
	var p guardiansPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Guardians.CreateGuardian(ctx, &guardians.CreateGuardianInput{
		UserID:   userID,
		SchoolID: schoolID,
		Guardian: guardians.CreateGuardian{
			Name:           p.Name,
			LastName:       p.LastName,
			Email:          p.Email,
			Phone:          p.Phone,
			Notes:          p.Notes,
			DateOfBirth:    p.DateOfBirth,
			AccountEnabled: p.AccountEnabled,
			Password:       p.Password,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, guardians.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, guardians.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, guardians.ErrNoRowsInserted):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id": res.InsertedID,
	})
}

func (h *Handler) ListGuardiansHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Guardians.ListGuardians(ctx, &guardians.ListGuardiansInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, guardians.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	type guardian struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		LastName       string    `json:"lastName"`
		Email          string    `json:"email"`
		Phone          *string   `json:"phone"`
		Notes          *string   `json:"notes"`
		DateOfBirth    time.Time `json:"dateOfBirth"`
		AccountEnabled bool      `json:"accountEnabled"`
		AccountActive  bool      `json:"accountActive"`
	}
	guardianList := make([]guardian, 0, len(res.Guardians))

	for _, g := range res.Guardians {
		guardianList = append(guardianList, guardian{
			ID:             strconv.FormatInt(g.ID, 10),
			Name:           g.Name,
			LastName:       g.LastName,
			Email:          g.Email,
			Phone:          g.Phone,
			Notes:          g.Notes,
			DateOfBirth:    g.DateOfBirth,
			AccountEnabled: g.AccountEnabled,
			AccountActive:  g.AccountActive,
		})
	}

	access, err := schools.GetAllUserPermissions(ctx, h.DB, userID, schoolID)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"guardians": guardianList,
		"access":    access,
	})
}

func (h *Handler) UpdateGuardianHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	guardianID, err := strconv.ParseInt(chi.URLParam(r, "guardianID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p guardiansPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.App.Staff.Guardians.UpdateGuardian(ctx, &guardians.UpdateGuardianInput{
		UserID:     userID,
		GuardianID: guardianID,
		Guardian: guardians.UpdateGuardian{
			Name:           p.Name,
			LastName:       p.LastName,
			Email:          p.Email,
			Phone:          p.Phone,
			Notes:          p.Notes,
			DateOfBirth:    p.DateOfBirth,
			AccountEnabled: p.AccountEnabled,
			Password:       p.Password,
		},
	}); err != nil {
		switch {
		case errors.Is(err, guardians.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, guardians.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteGuardianHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	guardianID, err := strconv.ParseInt(chi.URLParam(r, "guardianID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.Guardians.DeleteGuardian(ctx, &guardians.DeleteGuardianInput{
		UserID:     userID,
		GuardianID: guardianID,
	}); err != nil {
		switch {
		case errors.Is(err, guardians.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
