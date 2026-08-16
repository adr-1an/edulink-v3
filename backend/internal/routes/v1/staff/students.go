package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func SchoolStudentRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	// Create student
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		h.CreateStudentHandler(w, r, db, sf)
	})

	// List students
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListStudentHandler(w, r, db)
	})

	// Import students
	r.Post("/import", func(w http.ResponseWriter, r *http.Request) {
		h.ImportStudentHandler(w, r, db, sf)
	})

	return r
}

func StudentRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	r.Route("/{studentID}", func(r chi.Router) {
		// Update student
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			h.UpdateStudentHandler(w, r, db, sf)
		})

		// Delete student
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			h.DeleteStudentHandler(w, r, db, sf)
		})

		// View student profile
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			h.ViewStudentHandler(w, r, db)
		})
	})

	return r
}
