package staff

import (
	"app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func CourseAssignmentRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	// Create assignment
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		staff.CreateAssignmentHandler(w, r, db, sf)
	})

	// List assignments
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff.ListAssignmentsHandler(w, r, db)
	})

	return r
}

func AssignmentRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	r.Route("/{assignmentID}", func(r chi.Router) {
		// Update assignment
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			staff.UpdateAssignmentHandler(w, r, db, sf)
		})

		// Delete assignment
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff.DeleteAssignmentHandler(w, r, db, sf)
		})

		// Submissions
		r.Mount("/submissions", AssignmentSubmissionRoutes(db))
	})

	return r
}
