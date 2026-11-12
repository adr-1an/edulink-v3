package portal

import (
	handlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func Routes(h *handlers.Handler, studentHandler *students.Handler) chi.Router {
	r := chi.NewRouter()

	// Auth routes
	r.Mount("/auth", authRoutes(h))

	// Get profile
	r.Get("/profile", h.GetProfile)

	// Courses
	r.Mount("/courses", courseRoutes(studentHandler))

	// Posts
	r.Mount("/posts", postRoutes(h))

	// Assignments
	r.Mount("/assignments", assignmentRoutes(studentHandler))

	// Submissions
	r.Mount("/submissions", submissionRoutes(studentHandler, h))

	return r
}
