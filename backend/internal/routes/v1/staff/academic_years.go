package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func academicYearRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{yearID}", func(r chi.Router) {
		// Delete year
		r.Delete("/", h.DeleteAcademicYear)

		// Create grade
		r.Post("/grades", h.CreateGrade)
	})

	return r
}

// registerSchoolAcademicYears registers the school-scoped collection and promotion actions.
func registerSchoolAcademicYears(r chi.Router, h *handlers.Handler) {
	r.Post("/academic-years", h.CreateAcademicYear)
	r.Get("/academic-years", h.ListAcademicYears)
	r.Put("/academic-years", h.ClearAcademicYear)
	r.Post("/promote", h.PromoteSchool)
}
