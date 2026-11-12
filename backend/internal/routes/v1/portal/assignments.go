package portal

import (
	"app/internal/handlers/portal/students"
	"github.com/go-chi/chi/v5"
)

func assignmentRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListAllAssignments)
	r.Mount("/", assignmentDetailRoutes(h))
	return r
}

func assignmentDetailRoutes(h *students.Handler) chi.Router {
	r := chi.NewRouter()
	r.Route("/{assignmentID}", func(r chi.Router) {
		r.Post("/submissions", h.CreateAssignmentSubmission)
	})
	return r
}
