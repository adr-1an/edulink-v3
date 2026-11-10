package staff

import (
	"app/internal/application/staff/students"
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) AssignGuardianToStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	guardianID, err := strconv.ParseInt(chi.URLParam(r, "guardianID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.Students.AssignGuardian(ctx, &students.AssignGuardianInput{
		UserID:     userID,
		StudentID:  studentID,
		GuardianID: guardianID,
	}); err != nil {
		switch {
		case errors.Is(err, students.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnassignGuardianFromStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	guardianID, err := strconv.ParseInt(chi.URLParam(r, "guardianID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.Students.UnassignGuardian(ctx, &students.UnassignGuardianInput{
		UserID:     userID,
		StudentID:  studentID,
		GuardianID: guardianID,
	}); err != nil {
		switch {
		case errors.Is(err, students.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListStudentGuardians(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staffhelpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	studentID, err := strconv.ParseInt(chi.URLParam(r, "studentID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Students.ListGuardians(ctx, &students.ListGuardiansInput{
		UserID:    userID,
		StudentID: studentID,
	})
	if err != nil {
		switch {
		case errors.Is(err, students.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	type guardian struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		LastName    string  `json:"lastName"`
		DateOfBirth string  `json:"dateOfBirth"`
		Email       string  `json:"email"`
		Phone       *string `json:"phone"`
	}
	var guardians []guardian

	for _, g := range res.Guardians {
		guardians = append(guardians, guardian{
			ID:          strconv.FormatInt(g.ID, 10),
			Name:        g.Name,
			LastName:    g.LastName,
			DateOfBirth: g.DateOfBirth,
			Email:       g.Email,
			Phone:       g.Phone,
		})
	}

	access, err := schools.GetAllUserPermissions(ctx, h.DB, userID, res.SchoolID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":    access,
		"guardians": guardians,
		"schoolId":  res.SchoolID,
	})
}
