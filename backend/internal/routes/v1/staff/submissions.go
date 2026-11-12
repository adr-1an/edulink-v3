package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func assignmentSubmissionRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListAssignmentSubmissions)

	return r
}

func submissionRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{submissionID}", func(r chi.Router) {
		// View submission
		r.Get("/", h.ViewSubmission)

		// Return submission
		r.Delete("/return", h.ReturnSubmission)

		// Permanently delete a returned submission
		r.Delete("/", h.DeleteReturnedSubmission)

		// Grade submission
		r.Post("/grade", h.GradeSubmission)

		// Remove submission score/grade
		r.Delete("/grade", h.ClearSubmissionScore)
	})

	return r
}
