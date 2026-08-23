package portal

import (
	portalhandlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func CourseAssignmentRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()

	// List all student assignments
	r.Get("/", h.ListAllAssignmentsHandler)

	r.Mount("/", AssignmentRoutes(h))

	return r
}

func AssignmentRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{assignmentID}", func(r chi.Router) {
		// Begin assignment submission
		r.Post("/submissions", h.CreateAssignmentSubmissionHandler)
	})

	return r
}

func SubmissionAttachmentRoutes(h *portalhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Complete attachment upload
	r.Post("/{objectID}", h.CompleteUploadHandler)

	return r
}
