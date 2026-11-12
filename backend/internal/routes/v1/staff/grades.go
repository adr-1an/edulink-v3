package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func gradeRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{gradeID}", func(r chi.Router) {
		// Update grade
		r.Patch("/", h.UpdateGrade)

		// Delete grade
		r.Delete("/", h.DeleteGrade)

		// Courses
		r.Mount("/courses", gradeCourseRoutes(h))
	})

	return r
}
