package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func SchoolStudentRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Create student
	r.Post("/", h.CreateStudentHandler)

	// List students
	r.Get("/", h.ListStudentsHandler)

	// Import students
	r.Post("/import", h.ImportStudentHandler)

	return r
}

func StudentRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{studentID}", func(r chi.Router) {
		// Update student
		r.Patch("/", h.UpdateStudentHandler)

		// Delete student
		r.Delete("/", h.DeleteStudentHandler)

		// View student profile
		r.Get("/", h.ViewStudentHandler)

		// Upload profile picture
		r.Post("/profile-picture", h.UploadStudentPfpHandler)

		// Delete profile picture
		r.Delete("/profile-picture", h.RemoveStudentPfpHandler)

		// Student guardians
		r.Mount("/guardians", StudentGuardianRoutes(h))
	})

	return r
}
