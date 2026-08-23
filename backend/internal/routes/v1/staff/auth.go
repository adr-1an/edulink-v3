package staff

import (
	staff2 "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

// AuthRoutes - Main authentication routes for this app.
func AuthRoutes(h *staff2.Handler) chi.Router {
	r := chi.NewRouter()

	// Token check
	r.Get("/", h.TokenCheckHandler)

	// Check registration token
	r.Get("/register/{token}", h.CheckRegistrationTokenHandler)

	// Send registration link
	r.Post("/register", h.SendRegistrationLinkHandler)

	// Registration
	r.Post("/register/{token}", h.RegistrationHandler)

	// Login
	r.Post("/login", h.LoginHandler)

	// Send password reset email
	r.Post("/reset", h.SendPasswordResetEmailHandler)

	// Password reset
	r.Put("/reset/{token}", h.PasswordResetHandler)

	// Logout
	r.Delete("/logout", h.LogoutHandler)

	// Enable 2FA
	r.Post("/two-factor", h.Enable2faHandler)

	// Confirm 2FA
	r.Put("/two-factor", h.Verify2faHandler)

	// Disable 2FA
	r.Delete("/two-factor", h.Disable2faHandler)

	// Disable 2FA via recovery code
	r.Delete("/two-factor/recovery", h.RecoverTwoFactorHandler)

	// 2FA challenge
	r.Post("/two-factor/challenge", h.CompleteChallenge)

	return r
}
