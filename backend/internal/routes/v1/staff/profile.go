package staff

import (
	staffHandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func ProfileRoutes(h *staffHandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Get profile data
	r.Get("/", h.GetProfileHandler)

	// Update profile
	r.Patch("/", h.UpdateProfileHandler)

	// Send email change link
	r.Post("/email", h.SendEmailChangeHandler)

	// Change email
	r.Put("/email/{token}", h.EmailUpdateHandler)

	// Change password
	r.Put("/password", h.PasswordChangeHandler)

	// Upload profile picture
	r.Post("/profile-picture", h.UploadPfpHandler)

	// Remove profile picture
	r.Delete("/profile-picture", h.ClearPfpHandler)

	return r
}
