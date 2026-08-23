package staff

import (
	staffhandlers "app/internal/handlers/staff"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GradeCourseRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create course
	r.Post("/", h.CreateCourseHandler)

	// List courses
	r.Get("/", h.ListCoursesHandler)

	return r
}

func CourseRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{courseID}", func(r chi.Router) {
		// Update course
		r.Patch("/", h.UpdateCourseHandler)

		// Delete course
		r.Delete("/", h.DeleteCourseHandler)

		// Posts
		r.Mount("/posts", CoursePostRoutes(h))

		// Assign student to course
		r.Post("/students", func(w http.ResponseWriter, r *http.Request) {
			h.AddOrRemoveCourseStudentHandler(w, r, true)
		})

		// Remove student from course
		r.Delete("/students", func(w http.ResponseWriter, r *http.Request) {
			h.AddOrRemoveCourseStudentHandler(w, r, false)
		})

		// List course students
		r.Get("/students", h.ListCourseStudentsHandler)

		// Assignments
		r.Mount("/assignments", CourseAssignmentRoutes(h))
	})

	return r
}
