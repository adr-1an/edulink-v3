package portal

import (
	handlers "app/internal/handlers/portal"

	"github.com/go-chi/chi/v5"
)

func postRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// View post
		r.Get("/", h.ViewPost)
	})

	return r
}
