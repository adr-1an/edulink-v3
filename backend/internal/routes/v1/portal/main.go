package portal

import (
	"app/internal/handlers/portal/students"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func MainPortalRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Auth routes
	r.Mount("/auth", AuthRoutes(db))

	// Get profile
	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		students.GetProfileHandler(w, r, db)
	})

	// Courses
	r.Mount("/courses", CourseRoutes(db, s3))

	// Posts
	r.Mount("/posts", PostRoutes(db, s3))

	// Assignments
	r.Mount("/assignments", CourseAssignmentRoutes(db, sf, s3))

	// Submissions
	r.Mount("/submissions", SubmissionRoutes(db, sf, s3))

	return r
}
