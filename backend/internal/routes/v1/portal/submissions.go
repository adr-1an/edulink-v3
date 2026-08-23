package portal

import (
	portalhandlers "app/internal/handlers/portal"
	"app/internal/handlers/portal/students"

	"github.com/go-chi/chi/v5"
)

func SubmissionRoutes(h *students.Handler, portalHandler *portalhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Submission attachments
	r.Mount("/attachments", SubmissionAttachmentRoutes(portalHandler))

	r.Route("/{submissionID}", func(r chi.Router) {
		// Init attachment upload
		r.Post("/attachments", h.InitSubmissionAttachmentUpload)

		// Delete a draft attachment
		r.Delete("/attachments/{attachmentID}", h.DeleteSubmissionAttachmentHandler)

		// Finish submission creation
		r.Post("/submit", h.CompleteSubmissionCreationHandler)
	})

	return r
}
