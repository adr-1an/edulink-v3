package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func coursePostRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create course post
	r.Post("/", h.CreatePost)

	// List posts
	r.Get("/", h.ListPosts)

	return r
}

func postRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// Update post
		r.Patch("/", h.UpdatePost)

		// Delete post
		r.Delete("/", h.DeletePost)

		// View post
		r.Get("/", h.ViewPost)

		// Init file upload
		r.Post("/upload", h.InitPostAttachmentUpload)
	})

	// Post attachments
	r.Mount("/attachments", postAttachmentRoutes(h))

	return r
}

func postAttachmentRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{attachmentID}", func(r chi.Router) {
		// Delete attachment
		r.Delete("/", h.DeletePostAttachment)
	})

	return r
}
