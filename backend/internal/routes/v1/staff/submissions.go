package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func AssignmentSubmissionRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.ListAssignmentSubmissionsHandler)

	return r
}

func SubmissionRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{submissionID}", func(r chi.Router) {
		// View submission
		r.Get("/", h.ViewSubmissionHandler)

		// Return submission
		r.Delete("/return", h.ReturnSubmissionHandler)

		// Permanently delete a returned submission
		r.Delete("/", h.DeleteReturnedSubmissionHandler)

		// Grade submission
		r.Post("/grade", h.GradeSubmissionHandler)

		// Remove submission score/grade
		r.Delete("/grade", h.ClearSubmissionScoreHandler)
	})

	return r
}
