package portal

import (
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func CourseRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()

	// List courses
	r.Get("/", h.ListCoursesHandler)

	r.Route("/{courseID}", func(r chi.Router) {
		// Course dashboard
		r.Get("/", h.CourseDashboardHandler)
	})

	return r
}
