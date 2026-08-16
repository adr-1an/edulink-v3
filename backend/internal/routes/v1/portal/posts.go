package portal

import (
	"app/internal/handlers/portal"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func PostRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// View post
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			portal.ViewPostHandler(w, r, db, s3)
		})
	})

	return r
}
