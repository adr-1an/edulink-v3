package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func courseAssignmentRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create assignment
	r.Post("/", h.CreateAssignment)

	// List assignments
	r.Get("/", h.ListAssignments)

	return r
}

func assignmentRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{assignmentID}", func(r chi.Router) {
		// Update assignment
		r.Patch("/", h.UpdateAssignment)

		// Delete assignment
		r.Delete("/", h.DeleteAssignment)

		// Submissions
		r.Mount("/submissions", assignmentSubmissionRoutes(h))
	})

	return r
}
