package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/academicyears"
	"app/internal/helpers/staff"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListAcademicYearsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
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

	res, err := h.App.Staff.AcademicYears.ListAcademicYears(ctx, &academicyears.ListAcademicYearsInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, academicyears.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// AC Year struct
	type AcYear struct {
		ID        string `json:"id"`
		StartYear int    `json:"startYear"`
		EndYear   int    `json:"endYear"`
		IsActive  bool   `json:"isActive"`
	}
	var years []AcYear

	for _, y := range res.AcademicYears {
		years = append(years, AcYear{
			ID:        strconv.FormatInt(y.ID, 10),
			StartYear: y.StartYear,
			EndYear:   y.EndYear,
			IsActive:  y.IsActive,
		})
	}

	// Get user access
	access := res.Access

	// Return data
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":        access,
		"academicYears": years,
	})
}

func (h *Handler) CreateAcademicYearHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
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
		AcademicYear struct {
			From int `json:"from"`
			To   int `json:"to"`
		} `json:"academicYear"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.AcademicYears.CreateAcademicYear(ctx, &academicyears.CreateAcademicYearInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		SchoolID:       schoolID,
		AcademicYear: academicyears.CreateAcademicYear{
			From: p.AcademicYear.From,
			To:   p.AcademicYear.To,
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, academicyears.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, academicyears.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, academicyears.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, academicyears.ErrAcademicYearConflict):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": staffhelpers.ErrorCodeAcademicYearConflict,
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeleteAcademicYearHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get year ID
	acYearIDStr := chi.URLParam(r, "yearID")
	acYearID, err := strconv.ParseInt(acYearIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.AcademicYears.DeleteAcademicYear(ctx, &academicyears.DeleteAcademicYearInput{
		UserID:         userID,
		AcademicYearID: acYearID,
	})
	if err != nil {
		switch {
		case errors.Is(err, academicyears.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, academicyears.ErrYearIsActive):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorYearIsActive})
		case errors.Is(err, academicyears.ErrYearHasGrades):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": staffhelpers.ErrorYearHasGrades})
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ClearAcademicYearHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	schoolID, err := strconv.ParseInt(chi.URLParam(r, "schoolID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.AcademicYears.ClearAcademicYear(ctx, &academicyears.ClearAcademicYearInput{
		UserID:   userID,
		SchoolID: schoolID,
	})
	if err != nil {
		switch {
		case errors.Is(err, academicyears.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
