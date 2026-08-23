package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func CourseAssignmentRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create assignment
	r.Post("/", h.CreateAssignmentHandler)

	// List assignments
	r.Get("/", h.ListAssignmentsHandler)

	return r
}

func AssignmentRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{assignmentID}", func(r chi.Router) {
		// Update assignment
		r.Patch("/", h.UpdateAssignmentHandler)

		// Delete assignment
		r.Delete("/", h.DeleteAssignmentHandler)

		// Submissions
		r.Mount("/submissions", AssignmentSubmissionRoutes(h))
	})

	return r
}
