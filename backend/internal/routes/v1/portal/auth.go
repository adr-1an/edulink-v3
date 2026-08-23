package portal

import (
	portalhandlers "app/internal/handlers/portal"

	"github.com/go-chi/chi/v5"
)

func AuthRoutes(h *portalhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Login
	r.Post("/login", h.LoginHandler)

	// Token check
	r.Get("/", h.TokenCheckHandler)

	// Logout
	r.Delete("/", h.LogoutHandler)

	// Activate account
	r.Post("/activate/{token}", h.AccountActivationHandler)

	return r
}
