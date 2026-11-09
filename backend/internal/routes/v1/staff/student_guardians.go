package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func StudentGuardianRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List student guardians
	r.Get("/", h.ListStudentGuardians)

	r.Route("/{guardianID}", func(r chi.Router) {
		// Assign guardian to student
		r.Put("/", h.AssignGuardianToStudent)

		// Unassign guardian from student
		r.Delete("/", h.UnassignGuardianFromStudent)
	})

	return r
}
