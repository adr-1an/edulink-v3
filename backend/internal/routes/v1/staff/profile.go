package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func ProfileRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Get profile data
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { h.GetProfileHandler(w, r, db, s3) })

	// Update profile
	r.Patch("/", func(w http.ResponseWriter, r *http.Request) { h.UpdateProfileHandler(w, r, db) })

	// Send email change link
	r.Post("/email", func(w http.ResponseWriter, r *http.Request) { h.SendEmailChangeHandler(w, r, db) })

	// Change email
	r.Put("/email/{token}", func(w http.ResponseWriter, r *http.Request) { h.EmailUpdateHandler(w, r, db) })

	// Change password
	r.Put("/password", func(w http.ResponseWriter, r *http.Request) { h.PasswordChangeHandler(w, r, db) })

	// Upload profile picture
	r.Post("/profile-picture", func(w http.ResponseWriter, r *http.Request) { h.UploadPfpHandler(w, r, db, sf, s3) })

	// Remove profile picture
	r.Delete("/profile-picture", func(w http.ResponseWriter, r *http.Request) { h.ClearPfpHandler(w, r, db, s3) })

	return r
}
