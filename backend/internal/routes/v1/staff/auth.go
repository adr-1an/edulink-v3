package staff

import (
	staff2 "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

// AuthRoutes - Main authentication routes for this app.
func AuthRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
) chi.Router {
	r := chi.NewRouter()

	// Token check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		staff2.TokenCheckHandler(w, r, db)
	})

	// Check registration token
	r.Get("/register/{token}", func(w http.ResponseWriter, r *http.Request) {
		staff2.CheckRegistrationTokenHandler(w, r, db)
	})

	// Send registration link
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		staff2.SendRegistrationLinkHandler(w, r, db)
	})

	// Registration
	r.Post("/register/{token}", func(w http.ResponseWriter, r *http.Request) {
		staff2.RegistrationHandler(w, r, db, sf)
	})

	// Login
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		staff2.LoginHandler(w, r, db, sf)
	})

	// Send password reset email
	r.Post("/reset", func(w http.ResponseWriter, r *http.Request) {
		staff2.SendPasswordResetEmailHandler(w, r, db)
	})

	// Password reset
	r.Put("/reset/{token}", func(w http.ResponseWriter, r *http.Request) {
		staff2.PasswordResetHandler(w, r, db)
	})

	// Logout
	r.Delete("/logout", func(w http.ResponseWriter, r *http.Request) {
		staff2.LogoutHandler(w, r, db)
	})

	// Enable 2FA
	r.Post("/two-factor", func(w http.ResponseWriter, r *http.Request) {
		staff2.Enable2faHandler(w, r, db)
	})

	// Confirm 2FA
	r.Put("/two-factor", func(w http.ResponseWriter, r *http.Request) {
		staff2.Verify2faHandler(w, r, db)
	})

	// Disable 2FA
	r.Delete("/two-factor", func(w http.ResponseWriter, r *http.Request) {
		staff2.Disable2faHandler(w, r, db)
	})

	// Disable 2FA via recovery code
	r.Delete("/two-factor/recovery", func(w http.ResponseWriter, r *http.Request) {
		staff2.RecoverTwoFactorHandler(w, r, db)
	})

	// 2FA challenge
	r.Post("/two-factor/challenge", func(w http.ResponseWriter, r *http.Request) {
		staff2.CompleteChallenge(w, r, db, sf)
	})

	return r
}
