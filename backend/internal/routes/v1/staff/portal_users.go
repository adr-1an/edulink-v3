package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func PortalUserRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{id}", func(r chi.Router) {
		// Send account activation email
		r.Post("/activation/send", h.SendPortalUserAccountVerificationEmail)
	})

	return r
}
