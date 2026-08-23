package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func UploadRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{objectID}", func(r chi.Router) {
		// Complete file upload
		r.Post("/", h.CompleteUploadHandler)
	})

	return r
}
