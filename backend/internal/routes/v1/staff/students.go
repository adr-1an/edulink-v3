package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func schoolStudentRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create student
	r.Post("/", h.CreateStudent)

	// List students
	r.Get("/", h.ListStudents)

	// Import students
	r.Post("/import", h.ImportStudents)

	return r
}

func studentRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{studentID}", func(r chi.Router) {
		// Update student
		r.Patch("/", h.UpdateStudent)

		// Delete student
		r.Delete("/", h.DeleteStudent)

		// View student profile
		r.Get("/", h.ViewStudent)

		// Upload profile picture
		r.Post("/profile-picture", h.UploadStudentProfilePicture)

		// Delete profile picture
		r.Delete("/profile-picture", h.ClearStudentProfilePicture)

		// Student guardians
		r.Mount("/guardians", studentGuardianRoutes(h))
	})

	return r
}

func studentGuardianRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List student guardians
	r.Get("/", h.ListStudentGuardians)

	r.Route("/{guardianID}", func(r chi.Router) {
		// Assign guardian to student
		r.Put("/", h.AssignGuardianToStudent)

		// Unassign guardian from student
		r.Delete("/", h.UnassignGuardianFromStudent)
	})

	return r
}
