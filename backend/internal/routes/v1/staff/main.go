package staff

import (
	"app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func MainStaffRoutes(h *staff.Handler) chi.Router {
	r := chi.NewRouter()

	// Auth routes
	r.Mount("/auth", AuthRoutes(h))

	// Profile routes
	r.Mount("/profile", ProfileRoutes(h))

	// School routes
	r.Mount("/schools", SchoolRoutes(h))

	// Staff invitation routes
	r.Mount("/staff-invitations", StaffInvitationRoutes(h))

	// Academic year routes
	r.Mount("/academic-years", AcademicYearRoutes(h))

	// Grade routes
	r.Mount("/grades", GradeRoutes(h))

	// Role routes
	r.Mount("/roles", RoleRoutes(h))

	// Course routes
	r.Mount("/courses", CourseRoutes(h))

	// Staff member routes
	r.Mount("/staff-members", UserStaffMemberRoutes(h))

	// Course post routes
	r.Mount("/course-posts", PostRoutes(h))

	// Students
	r.Mount("/students", StudentRoutes(h))

	// Assignments
	r.Mount("/assignments", AssignmentRoutes(h))

	// Uploads
	r.Mount("/uploads", UploadRoutes(h))

	// Submissions
	r.Mount("/submissions", SubmissionRoutes(h))

	// Guardian routes
	r.Mount("/guardians", GuardianRoutes(h))

	// General portal user routes
	r.Mount("/portal-users", PortalUserRoutes(h))

	return r
}
