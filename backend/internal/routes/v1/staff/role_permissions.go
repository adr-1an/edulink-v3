package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func RolePermissionsRouter(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Set permission
	r.Put("/", h.SetStaffRolePermissionHandler)

	return r
}
