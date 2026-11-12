package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func schoolRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List schools
	r.Get("/", h.ListSchools)

	// Create school
	r.Post("/", h.CreateSchool)

	// School-specific
	r.Route("/{schoolID}", func(r chi.Router) {
		// View school dashboard
		r.Get("/", h.ViewSchool)

		// Update school
		r.Patch("/", h.UpdateSchool)

		// Delete school
		r.Delete("/", h.DeleteSchool)

		// Staff
		r.Mount("/staff", schoolMemberRoutes(h))

		registerSchoolAcademicYears(r, h)

		// List grades
		r.Get("/grades", h.ListGrades)

		// Roles
		r.Mount("/roles", schoolRoleRoutes(h))

		// List logs
		r.Get("/logs", h.ListSchoolLogs)

		// List staff invitations
		r.Get("/staff-invitations", h.ListSchoolInvitations)

		// Leave school (for staff members)
		r.Delete("/leave", h.LeaveSchool)

		// Students
		r.Mount("/students", schoolStudentRoutes(h))

		// Guardians
		r.Mount("/guardians", schoolGuardianRoutes(h))
	})

	return r
}
