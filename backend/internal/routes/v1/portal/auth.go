// Package portal
package portal

import (
	handlers "app/internal/handlers/portal"

	"github.com/go-chi/chi/v5"
)

func authRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Login
	r.Post("/login", h.Login)

	// Token check
	r.Get("/", h.CheckToken)

	// Logout
	r.Delete("/", h.Logout)

	// Activate account
	r.Post("/activate/{token}", h.ActivateAccount)

	// Change password
	r.Put("/password", h.UpdatePassword)

	return r
}
