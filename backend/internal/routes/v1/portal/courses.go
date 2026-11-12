package portal

import (
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func courseRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()

	// List courses
	r.Get("/", h.ListCourses)

	r.Route("/{courseID}", func(r chi.Router) {
		// Course dashboard
		r.Get("/", h.ViewCourseDashboard)
	})

	return r
}
