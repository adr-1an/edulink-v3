package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func SchoolGuardianRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create guardian
	r.Post("/", h.CreateGuardianHandler)

	// List guardians
	r.Get("/", h.ListGuardiansHandler)

	return r
}

func GuardianRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{guardianID}", func(r chi.Router) {
		// Update guardian
		r.Patch("/", h.UpdateGuardianHandler)

		// Delete guardian
		r.Delete("/", h.DeleteGuardianHandler)
	})

	return r
}
