package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/grades"
	"app/internal/auth"
	"app/internal/responsecodes"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListGrades(w http.ResponseWriter, r *http.Request) {
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
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type Grade struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	res, err := h.App.Staff.Grades.ListGrades(ctx, &grades.ListGradesInput{UserID: userID, SchoolID: schoolID})
	if err != nil {
		switch {
		case errors.Is(err, grades.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	var grades []Grade
	if res.Grades != nil {
		grades = make([]Grade, 0, len(res.Grades))
	}
	for _, item := range res.Grades {
		grades = append(grades, Grade{ID: item.ID, Name: item.Name, Level: item.Level})
	}
	// Get user permissions
	access := res.Access

	// Return grades
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access": access,
		"grades": grades,
	})
}

func (h *Handler) CreateGrade(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	yearID, err := strconv.ParseInt(chi.URLParam(r, "yearID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Payload
	type Payload struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.Grades.CreateGrade(ctx, &grades.CreateGradeInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		YearID:         yearID,
		Grade:          grades.CreateGrade{Name: p.Name, Level: p.Level},
	})
	if err != nil {
		switch {
		case errors.Is(err, grades.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, grades.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, grades.ErrMissingLevelVar):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeMissingLevelVar})
		case errors.Is(err, grades.ErrInvalidName):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidName})
		case errors.Is(err, grades.ErrLevelOutOfRange):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeLevelOutOfRange})
		case errors.Is(err, grades.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.NoRowsAffected {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdateGrade(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Payload
	type Payload struct {
		Name  string `json:"name"`
		Level int    `json:"level"`
	}
	var p Payload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)

	res, err := h.App.Staff.Grades.UpdateGrade(ctx, &grades.UpdateGradeInput{
		InvalidPayload: decodeErr != nil,
		UserID:         userID,
		GradeID:        gradeID,
		Grade:          grades.UpdateGrade{Name: p.Name, Level: p.Level},
	})
	if err != nil {
		switch {
		case errors.Is(err, grades.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, grades.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, grades.ErrMissingLevelVar):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeMissingLevelVar})
		case errors.Is(err, grades.ErrInvalidName):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeInvalidName})
		case errors.Is(err, grades.ErrLevelOutOfRange):
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{"code": responsecodes.ErrorCodeLevelOutOfRange})
		case errors.Is(err, grades.ErrConflict):
			w.WriteHeader(http.StatusConflict)
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

func (h *Handler) DeleteGrade(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID
	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	// Get school ID
	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Grades.DeleteGrade(ctx, &grades.DeleteGradeInput{
		UserID:  userID,
		GradeID: gradeID,
	})
	if err != nil {
		switch {
		case errors.Is(err, grades.ErrForbidden):
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
