package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func GradeCourseRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	// Create course
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.CreateCourseHandler(w, r, db, sf)
	})

	// List courses
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.ListCoursesHandler(w, r, db)
	})

	return r
}

func CourseRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	r.Route("/{courseID}", func(r chi.Router) {
		// Update course
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.UpdateCourseHandler(w, r, db, sf)
		})

		// Delete course
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeleteCourseHandler(w, r, db, sf)
		})

		// Posts
		r.Mount("/posts", CoursePostRoutes(db, sf, s3))

		// Assign student to course
		r.Post("/students", func(w http.ResponseWriter, r *http.Request) {
			staff2.AddOrRemoveCourseStudentHandler(w, r, db, sf, true)
		})

		// Remove student from course
		r.Delete("/students", func(w http.ResponseWriter, r *http.Request) {
			staff2.AddOrRemoveCourseStudentHandler(w, r, db, sf, false)
		})

		// List course students
		r.Get("/students", func(w http.ResponseWriter, r *http.Request) {
			staff2.ListCourseStudentsHandler(w, r, db)
		})

		// Assignments
		r.Mount("/assignments", CourseAssignmentRoutes(db, sf))
	})

	return r
}
