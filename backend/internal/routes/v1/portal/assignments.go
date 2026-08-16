package portal

import (
	"app/internal/handlers/portal"
	students2 "app/internal/handlers/portal/students"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func CourseAssignmentRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// List all student assignments
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		students2.ListAllAssignmentsHandler(w, r, db, s3)
	})

	r.Mount("/", AssignmentRoutes(db, sf))

	return r
}

func AssignmentRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	r.Route("/{assignmentID}", func(r chi.Router) {
		// Begin assignment submission
		r.Post("/submissions", func(w http.ResponseWriter, r *http.Request) {
			students2.CreateAssignmentSubmissionHandler(w, r, db, sf)
		})
	})

	return r
}

func SubmissionAttachmentRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Complete attachment upload
	r.Post("/{objectID}", func(w http.ResponseWriter, r *http.Request) {
		portal.CompleteUploadHandler(w, r, db, s3)
	})

	return r
}
