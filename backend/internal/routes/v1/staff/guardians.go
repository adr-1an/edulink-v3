package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func schoolGuardianRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create guardian
	r.Post("/", h.CreateGuardian)

	// List guardians
	r.Get("/", h.ListGuardians)

	return r
}

func guardianRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{guardianID}", func(r chi.Router) {
		// Update guardian
		r.Patch("/", h.UpdateGuardian)

		// Delete guardian
		r.Delete("/", h.DeleteGuardian)

		// List students
		r.Get("/", h.ListGuardianStudents)
	})

	return r
}
