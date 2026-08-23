package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func CoursePostRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create course post
	r.Post("/", h.CreatePostHandler)

	// List posts
	r.Get("/", h.ListPostsHandler)

	return r
}

func PostRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// Update post
		r.Patch("/", h.UpdatePostHandler)

		// Delete post
		r.Delete("/", h.DeletePostHandler)

		// View post
		r.Get("/", h.ViewPostHandler)

		// Init file upload
		r.Post("/upload", h.InitPostAttachmentUploadHandler)
	})

	// Post attachments
	r.Mount("/attachments", AttachmentRoutes(h))

	return r
}

func AttachmentRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{attachmentID}", func(r chi.Router) {
		// Delete attachment
		r.Delete("/", h.DeletePostAttachmentHandler)
	})

	return r
}
