package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func AcademicYearRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/{yearID}", func(r chi.Router) {
		// Delete year
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeleteAcademicYearHandler(w, r, db, sf)
		})

		// Create grade
		r.Post("/grades", func(w http.ResponseWriter, r *http.Request) {
			staff2.CreateGradeHandler(w, r, db, sf)
		})
	})

	return r
}
