package portal

import (
	portalhandlers "app/internal/handlers/portal"

	"github.com/go-chi/chi/v5"
)

func PostRoutes(h *portalhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// View post
		r.Get("/", h.ViewPostHandler)
	})

	return r
}
