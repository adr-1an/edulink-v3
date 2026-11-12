package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func uploadRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{objectID}", func(r chi.Router) {
		// Complete file upload
		r.Post("/", h.CompleteUpload)
	})

	return r
}
