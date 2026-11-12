package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func portalUserRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{id}", func(r chi.Router) {
		// Send account activation email
		r.Post("/activation/send", h.SendPortalUserActivationEmail)
	})

	return r
}
