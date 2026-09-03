package portal

import (
	portalhandlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func MainPortalRoutes(h *portalhandlers.Handler, studentHandler *students.Handler) chi.Router {
	r := chi.NewRouter()

	// Auth routes
	r.Mount("/auth", AuthRoutes(h))

	// Get profile
	r.Get("/profile", h.GetProfileHandler)

	// Courses
	r.Mount("/courses", CourseRoutes(studentHandler))

	// Posts
	r.Mount("/posts", PostRoutes(h))

	// Assignments
	r.Mount("/assignments", CourseAssignmentRoutes(studentHandler))

	// Submissions
	r.Mount("/submissions", SubmissionRoutes(studentHandler, h))

	return r
}
