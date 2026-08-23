package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func AcademicYearRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{yearID}", func(r chi.Router) {
		// Delete year
		r.Delete("/", h.DeleteAcademicYearHandler)

		// Create grade
		r.Post("/grades", h.CreateGradeHandler)
	})

	return r
}
