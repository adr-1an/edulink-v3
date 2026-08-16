package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func AssignmentSubmissionRoutes(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.ListAssignmentSubmissionsHandler(w, r, db)
	})

	return r
}

func SubmissionRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{submissionID}", func(r chi.Router) {
		// View submission
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.ViewSubmissionHandler(w, r, db, s3)
		})

		// Return submission
		r.Delete("/return", func(w http.ResponseWriter, r *http.Request) {
			staff2.ReturnSubmissionHandler(w, r, db)
		})

		// Permanently delete a returned submission
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeleteReturnedSubmissionHandler(w, r, db, s3)
		})

		// Grade submission
		r.Post("/grade", func(w http.ResponseWriter, r *http.Request) {
			staff2.GradeSubmissionHandler(w, r, db)
		})

		// Remove submission score/grade
		r.Delete("/grade", func(w http.ResponseWriter, r *http.Request) {
			staff2.ClearSubmissionScoreHandler(w, r, db)
		})
	})

	return r
}
