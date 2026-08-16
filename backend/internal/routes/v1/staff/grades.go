package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func GradeRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/{gradeID}", func(r chi.Router) {
		// Update grade
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			h.UpdateGradeHandler(w, r, db, sf)
		})

		// Delete grade
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			h.DeleteGradeHandler(w, r, db, sf)
		})

		// Courses
		r.Mount("/courses", GradeCourseRoutes(db, sf))
	})

	return r
}
