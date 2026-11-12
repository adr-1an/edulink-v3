package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

// authRoutes - Main authentication routes for this app.
func authRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Token check
	r.Get("/", h.CheckToken)

	// Check registration token
	r.Get("/register/{token}", h.CheckRegistrationToken)

	// Send registration link
	r.Post("/register", h.SendRegistrationLink)

	// Registration
	r.Post("/register/{token}", h.Register)

	// Login
	r.Post("/login", h.Login)

	// Send password reset email
	r.Post("/reset", h.SendPasswordResetEmail)

	// Password reset
	r.Put("/reset/{token}", h.ResetPassword)

	// Logout
	r.Delete("/logout", h.Logout)

	// Enable 2FA
	r.Post("/two-factor", h.EnableTwoFactor)

	// Confirm 2FA
	r.Put("/two-factor", h.VerifyTwoFactor)

	// Disable 2FA
	r.Delete("/two-factor", h.DisableTwoFactor)

	// Disable 2FA via recovery code
	r.Delete("/two-factor/recovery", h.RecoverTwoFactor)

	// 2FA challenge
	r.Post("/two-factor/challenge", h.CompleteTwoFactorChallenge)

	return r
}
