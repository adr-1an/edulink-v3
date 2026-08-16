package staff

import (
	"app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func UploadRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{objectID}", func(r chi.Router) {
		// Complete file upload
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			staff.CompleteUploadHandler(w, r, db, s3)
		})
	})

	return r
}
