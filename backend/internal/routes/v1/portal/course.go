package portal

import (
	"app/internal/handlers/portal/students"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

func CourseRoutes(db *sql.DB, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// List courses
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		students.ListCoursesHandler(w, r, db)
	})

	r.Route("/{courseID}", func(r chi.Router) {
		// Course dashboard
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			students.CourseDashboardHandler(w, r, db, s3)
		})
	})

	return r
}
