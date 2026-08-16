package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func CoursePostRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Create course post
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.CreatePostHandler(w, r, db, sf)
	})

	// List posts
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.ListPostsHandler(w, r, db, s3)
	})

	return r
}

func PostRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID}", func(r chi.Router) {
		// Update post
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.UpdatePostHandler(w, r, db, sf)
		})

		// Delete post
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeletePostHandler(w, r, db, sf)
		})

		// View post
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.ViewPostHandler(w, r, db, s3)
		})

		// Init file upload
		r.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			staff2.InitPostAttachmentUploadHandler(w, r, db, sf, s3)
		})
	})

	// Post attachments
	r.Mount("/attachments", AttachmentRoutes(db, s3))

	return r
}

func AttachmentRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{attachmentID}", func(r chi.Router) {
		// Delete attachment
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeletePostAttachmentHandler(w, r, db, s3)
		})
	})

	return r
}
