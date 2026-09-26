package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func SchoolRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List schools
	r.Get("/", h.SchoolListHandler)

	// Create school
	r.Post("/", h.CreateSchoolHandler)

	// School-specific
	r.Route("/{schoolID}", func(r chi.Router) {
		// View school dashboard
		r.Get("/", h.ViewSchoolDashboardHandler)

		// Update school
		r.Patch("/", h.UpdateSchoolHandler)

		// Delete school
		r.Delete("/", h.DeleteSchoolHandler)

		// Staff
		r.Mount("/staff", SchoolStaffMemberRoutes(h))

		// Create academic year
		r.Post("/academic-years", h.CreateAcademicYearHandler)

		// List academic years
		r.Get("/academic-years", h.ListAcademicYearsHandler)

		// List grades
		r.Get("/grades", h.ListGradesHandler)

		// Promote school
		r.Post("/promote", h.SchoolPromotionHandler)

		// Clear active academic year
		r.Put("/academic-years", h.ClearAcademicYearHandler)

		// Roles
		r.Mount("/roles", SchoolRoleRoutes(h))

		// List logs
		r.Get("/logs", h.ListSchoolLogsHandler)

		// List staff invitations
		r.Get("/staff-invitations", h.ListSchoolInvitationsHandler)

		// Leave school (for staff members)
		r.Delete("/leave", h.LeaveSchoolStaffHandler)

		// Students
		r.Mount("/students", SchoolStudentRoutes(h))

		// Guardians
		r.Mount("/guardians", SchoolGuardianRoutes(h))
	})

	return r
}
