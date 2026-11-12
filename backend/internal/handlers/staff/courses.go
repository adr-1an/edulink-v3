package staff

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"app/internal/application/staff/courses"
	"app/internal/auth"

	"github.com/go-chi/chi/v5"
)

type CreateCoursePayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (h *Handler) CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var p CreateCoursePayload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Staff.Courses.CreateCourse(ctx, &courses.CreateCourseInput{InvalidPayload: decodeErr != nil, UserID: userID, GradeID: gradeID, Data: courses.CreateCourseData{Name: p.Name, Description: p.Description, Color: p.Color}})
	if err != nil {
		switch {
		case errors.Is(err, courses.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, courses.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, courses.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, courses.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
}

type UpdateCoursePayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (h *Handler) UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {
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

	var p UpdateCoursePayload

	// Decode
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	decodeErr := dec.Decode(&p)
	res, err := h.App.Staff.Courses.UpdateCourse(ctx, &courses.UpdateCourseInput{InvalidPayload: decodeErr != nil, UserID: userID, CourseID: courseID, Data: courses.UpdateCourseData{Name: p.Name, Description: p.Description, Color: p.Color}})
	if err != nil {
		switch {
		case errors.Is(err, courses.ErrBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, courses.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, courses.ErrUnprocessableEntity):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, courses.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteCourseHandler(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.App.Staff.Courses.DeleteCourse(ctx, &courses.DeleteCourseInput{UserID: userID, CourseID: courseID})
	if err != nil {
		switch {
		case errors.Is(err, courses.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if res.AffectedRows == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

type ListCoursesCourse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func mapListCoursesCourse(v courses.ListCoursesCourse) ListCoursesCourse {
	out := ListCoursesCourse{
		ID:          v.ID,
		Name:        v.Name,
		Description: v.Description,
		Color:       v.Color,
	}
	return out
}

func (h *Handler) ListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := auth.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	gradeID, err := strconv.ParseInt(chi.URLParam(r, "gradeID"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := h.App.Staff.Courses.ListCourses(ctx, &courses.ListCoursesInput{UserID: userID, GradeID: gradeID})
	if err != nil {
		switch {
		case errors.Is(err, courses.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	var mappedCourses []ListCoursesCourse
	if res.Courses != nil {
		mappedCourses = make([]ListCoursesCourse, 0, len(res.Courses))
		for _, item := range res.Courses {
			mappedCourses = append(mappedCourses, mapListCoursesCourse(item))
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access":  res.Access,
		"courses": mappedCourses,
	})
}
