package staff

import (
	handlers "app/internal/handlers/staff"
	"github.com/go-chi/chi/v5"
)

func registerCourseStudents(r chi.Router, h *handlers.Handler) {
	r.Get("/students", h.ListCourseStudents)
	r.Post("/students", h.AssignCourseStudent)
	r.Delete("/students", h.RemoveCourseStudent)
}
