package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func SchoolRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
	s3 *minio.Client,
) chi.Router {
	r := chi.NewRouter()

	// List schools
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.SchoolListHandler(w, r, db)
	})

	// Create school
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.CreateSchoolHandler(w, r, db, sf)
	})

	// School-specific
	r.Route("/{schoolID}", func(r chi.Router) {
		// View school dashboard
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.ViewSchoolDashboardHandler(w, r, db)
		})

		// Update school
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.UpdateSchoolHandler(w, r, db, sf)
		})

		// Delete school
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			staff2.DeleteSchoolHandler(w, r, db, sf)
		})

		// Staff
		r.Mount("/staff", SchoolStaffMemberRoutes(db, sf, s3))

		// Create academic year
		r.Post("/academic-years", func(w http.ResponseWriter, r *http.Request) {
			staff2.CreateAcademicYearHandler(w, r, db, sf)
		})

		// List academic years
		r.Get("/academic-years", func(w http.ResponseWriter, r *http.Request) {
			staff2.ListAcademicYearsHandler(w, r, db)
		})

		// List grades
		r.Get("/grades", func(w http.ResponseWriter, r *http.Request) {
			staff2.ListGradesHandler(w, r, db)
		})

		// Promote school
		r.Post("/promote", func(w http.ResponseWriter, r *http.Request) {
			staff2.SchoolPromotionHandler(w, r, db, sf)
		})

		// Clear active academic year
		r.Put("/academic-years", func(w http.ResponseWriter, r *http.Request) {
			staff2.ClearAcademicYearHandler(w, r, db, sf)
		})

		// Roles
		r.Mount("/roles", SchoolRoleRoutes(db, sf))

		// List logs
		r.Get("/logs", func(w http.ResponseWriter, r *http.Request) {
			staff2.ListSchoolLogsHandler(w, r, db)
		})

		// List staff invitations
		r.Get("/staff-invitations", func(w http.ResponseWriter, r *http.Request) {
			staff2.ListSchoolInvitationsHandler(w, r, db)
		})

		// Leave school (for staff members)
		r.Delete("/leave", func(w http.ResponseWriter, r *http.Request) {
			staff2.LeaveSchoolStaffHandler(w, r, db, sf)
		})

		// Students
		r.Mount("/students", SchoolStudentRoutes(db, sf))
	})

	return r
}
