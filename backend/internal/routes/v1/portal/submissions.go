package portal

import (
	handlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func submissionRoutes(h *students.Handler, portalHandler *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Submission attachments
	r.Mount("/attachments", submissionUploadRoutes(portalHandler))

	r.Route("/{submissionID}", func(r chi.Router) {
		// Init attachment upload
		r.Post("/attachments", h.InitSubmissionAttachmentUpload)

		// Delete a draft attachment
		r.Delete("/attachments/{attachmentID}", h.DeleteSubmissionAttachment)

		// Finish submission creation
		r.Post("/submit", h.CompleteSubmissionCreation)
	})

	return r
}

func submissionUploadRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Complete attachment upload
	r.Post("/{objectID}", h.CompleteUpload)

	return r
}
