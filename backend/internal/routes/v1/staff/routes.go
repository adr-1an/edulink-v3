package staff

import (
	handlers "app/internal/handlers/staff"
	"github.com/go-chi/chi/v5"
)

// Routes builds the staff API router.
func Routes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Mount("/auth", authRoutes(h))
	r.Mount("/profile", profileRoutes(h))
	r.Mount("/uploads", uploadRoutes(h))

	r.Mount("/schools", schoolRoutes(h))
	r.Mount("/academic-years", academicYearRoutes(h))
	r.Mount("/grades", gradeRoutes(h))
	r.Mount("/courses", courseRoutes(h))

	r.Mount("/staff-invitations", invitationRoutes(h))
	r.Mount("/staff-members", memberRoutes(h))
	r.Mount("/roles", roleRoutes(h))

	r.Mount("/students", studentRoutes(h))
	r.Mount("/guardians", guardianRoutes(h))
	r.Mount("/portal-users", portalUserRoutes(h))

	r.Mount("/course-posts", postRoutes(h))
	r.Mount("/assignments", assignmentRoutes(h))
	r.Mount("/submissions", submissionRoutes(h))

	return r
}
