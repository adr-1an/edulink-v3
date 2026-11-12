package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func profileRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Get profile data
	r.Get("/", h.GetProfile)

	// Update profile
	r.Patch("/", h.UpdateProfile)

	// Send email change link
	r.Post("/email", h.SendEmailChange)

	// Change email
	r.Put("/email/{token}", h.UpdateEmail)

	// Change password
	r.Put("/password", h.UpdatePassword)

	// Upload profile picture
	r.Post("/profile-picture", h.UploadProfilePicture)

	// Remove profile picture
	r.Delete("/profile-picture", h.ClearProfilePicture)

	return r
}
