package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/academicyears"
	"app/internal/auth"
	"app/internal/responsecodes"

	"github.com/go-chi/chi/v5"
)

// PromoteSchool promotes all the grades in the currently active Academic Year.
// The promotion process is roughly:
// - Creating a new Academic Year
// - Optionally copying grades, courses, student/teacher/parent accounts etc.
// - If grades are selected for copying, their level value can optionally be bumped by 1.
func (h *Handler) PromoteSchool(w http.ResponseWriter, r *http.Request) {
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

	// -- 2. Payload & options --
	type Payload struct {
		NewAcademicYear struct {
			From int `json:"from"`
			To   int `json:"to"`
		} `json:"newAcademicYear"`
		Opts struct {
			ActivateAfterPromotion bool `json:"activateAfterPromotion"`
			TransferGrades         bool `json:"transferGrades"`
			PromoteGradeLevels     bool `json:"promoteGradeLevels"`
			// TODO: Uncomment features below once they're added
			// TransferCourses bool `json:"transferCourses"`
			// TransferStudents       bool `json:"transferStudents"`
			// TransferTeachers bool `json:"transferTeachers"`
			// TransferParents bool `json:"transferParents"`
		} `json:"options"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.AcademicYears.PromoteSchool(ctx, &academicyears.PromoteSchoolInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		SchoolID:       schoolID,
		Promotion: academicyears.PromoteSchool{NewAcademicYear: struct {
			From int
			To   int
		}{From: p.NewAcademicYear.From, To: p.NewAcademicYear.To}, Opts: struct {
			ActivateAfterPromotion bool
			TransferGrades         bool
			PromoteGradeLevels     bool
		}{ActivateAfterPromotion: p.Opts.ActivateAfterPromotion, TransferGrades: p.Opts.TransferGrades, PromoteGradeLevels: p.Opts.PromoteGradeLevels}},
	})
	if err != nil {
		switch {
		case errors.Is(err, academicyears.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, academicyears.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, academicyears.ErrNoActiveYear):
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeNoActiveYear})
		case errors.Is(err, academicyears.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, academicyears.ErrAcademicYearConflict):
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeAcademicYearConflict})
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
