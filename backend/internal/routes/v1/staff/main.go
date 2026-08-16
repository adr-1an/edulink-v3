package staff

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func MainStaffRoutes(db *sql.DB, sf *sonyflake.Sonyflake, s3 *minio.Client) chi.Router {
	r := chi.NewRouter()

	// Auth routes
	r.Mount("/auth", AuthRoutes(db, sf))

	// Profile routes
	r.Mount("/profile", ProfileRoutes(db, sf, s3))

	// School routes
	r.Mount("/schools", SchoolRoutes(db, sf, s3))

	// Staff invitation routes
	r.Mount("/staff-invitations", StaffInvitationRoutes(db, sf))

	// Academic year routes
	r.Mount("/academic-years", AcademicYearRoutes(db, sf))

	// Grade routes
	r.Mount("/grades", GradeRoutes(db, sf))

	// Role routes
	r.Mount("/roles", RoleRoutes(db, sf))

	// Course routes
	r.Mount("/courses", CourseRoutes(db, sf, s3))

	// Staff member routes
	r.Mount("/staff-members", StaffMemberRoutes(db, sf))

	// Course post routes
	r.Mount("/course-posts", PostRoutes(db, sf, s3))

	// Students
	r.Mount("/students", StudentRoutes(db, sf))

	// Assignments
	r.Mount("/assignments", AssignmentRoutes(db, sf))

	// Uploads
	r.Mount("/uploads", UploadRoutes(db, s3))

	// Submissions
	r.Mount("/submissions", SubmissionRoutes(db, s3))

	return r
}
