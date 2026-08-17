package portal

import (
	"app/internal/handlers/portal"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func AuthRoutes(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	// Login
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		portal.LoginHandler(w, r, db)
	})

	// Token check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		portal.TokenCheckHandler(w, r, db)
	})

	// Logout
	r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
		portal.LogoutHandler(w, r, db)
	})

	// Activate account
	r.Post("/activate/{token}", func(w http.ResponseWriter, r *http.Request) {
		portal.AccountActivationHandler(w, r, db)
	})

	return r
}
