package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func GradeRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{gradeID}", func(r chi.Router) {
		// Update grade
		r.Patch("/", h.UpdateGradeHandler)

		// Delete grade
		r.Delete("/", h.DeleteGradeHandler)

		// Courses
		r.Mount("/courses", GradeCourseRoutes(h))
	})

	return r
}
