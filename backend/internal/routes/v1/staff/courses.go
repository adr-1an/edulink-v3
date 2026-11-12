package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func gradeCourseRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create course
	r.Post("/", h.CreateCourse)

	// List courses
	r.Get("/", h.ListCourses)

	return r
}

func courseRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{courseID}", func(r chi.Router) {
		// Update course
		r.Patch("/", h.UpdateCourse)

		// Delete course
		r.Delete("/", h.DeleteCourse)

		// Posts
		r.Mount("/posts", coursePostRoutes(h))

		registerCourseStudents(r, h)

		// Assignments
		r.Mount("/assignments", courseAssignmentRoutes(h))
	})

	return r
}
