package portal

import (
	"app/internal/handlers/portal/students"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func SubmissionRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Submission attachments
	r.Mount("/attachments", SubmissionAttachmentRoutes(db, s3))

	r.Route("/{submissionID}", func(r chi.Router) {
		// Init attachment upload
		r.Post("/attachments", func(w http.ResponseWriter, r *http.Request) {
			students.InitSubmissionAttachmentUpload(w, r, db, sf, s3)
		})

		// Delete a draft attachment
		r.Delete("/attachments/{attachmentID}", func(w http.ResponseWriter, r *http.Request) {
			students.DeleteSubmissionAttachmentHandler(w, r, db, s3)
		})

		// Finish submission creation
		r.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
			students.CompleteSubmissionCreationHandler(w, r, db)
		})
	})

	return r
}
