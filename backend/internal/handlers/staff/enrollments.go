package staff

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"app/internal/application/staff/enrollments"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type EnrollmentPayload struct {
	StudentID string `json:"studentId"`
}

func (h *Handler) AssignCourseStudent(w http.ResponseWriter, r *http.Request) {
	h.setCourseEnrollment(w, r, true)
}

func (h *Handler) RemoveCourseStudent(w http.ResponseWriter, r *http.Request) {
	h.setCourseEnrollment(w, r, false)
}

func (h *Handler) setCourseEnrollment(w http.ResponseWriter, r *http.Request, assign bool) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p EnrollmentPayload

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	err = h.App.Staff.Enrollments.SetEnrollment(ctx, &enrollments.SetEnrollmentInput{InvalidPayload: decodeErr != nil, UserID: userID, CourseID: courseID, Data: enrollments.EnrollmentData{StudentID: p.StudentID}, Assign: assign})
	if err != nil {
		switch {
		case errors.Is(err, enrollments.ErrBadRequest):
			log.Println(decodeErr)
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, enrollments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, enrollments.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if assign {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

type ListCourseStudentsStudent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Email    string `json:"email"`
	Assigned bool   `json:"assigned"`
}

func mapListCourseStudentsStudent(v enrollments.ListCourseStudentsStudent) ListCourseStudentsStudent {
	out := ListCourseStudentsStudent{
		ID:       v.ID,
		Name:     v.Name,
		LastName: v.LastName,
		Email:    v.Email,
		Assigned: v.Assigned,
	}
	return out
}

func (h *Handler) ListCourseStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Enrollments.ListCourseStudents(ctx, &enrollments.ListCourseStudentsInput{UserID: userID, CourseID: courseID})
	if err != nil {
		switch {
		case errors.Is(err, enrollments.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedStudents []ListCourseStudentsStudent
	if res.Students != nil {
		mappedStudents = make([]ListCourseStudentsStudent, 0, len(res.Students))
		for _, item := range res.Students {
			mappedStudents = append(mappedStudents, mapListCourseStudentsStudent(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":   res.Access,
		"students": mappedStudents,
	})
}
