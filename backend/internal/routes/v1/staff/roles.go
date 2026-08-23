package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func RoleRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List role permissions
	r.Get("/permissions", h.ListPermissionsHandler)

	// List roles
	r.Get("/", h.ListRolesHandler)

	// Role-specific
	r.Route("/{roleID}", func(r chi.Router) {
		// Update role
		r.Patch("/", h.UpdateRoleHandler)

		// Delete role
		r.Delete("/", h.DeleteRoleHandler)

		// Role permissions
		r.Mount("/permissions", RolePermissionsRouter(h))
	})

	return r
}

func SchoolRoleRoutes(h *staffhandlers.Handler) chi.Router {
	r := RoleRoutes(h)

	r.Post("/", h.CreateRoleHandler)

	return r
}

func StaffRoleRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List roles
	r.Get("/", h.ListStaffRolesHandler)

	r.Route("/{roleID}", func(r chi.Router) {
		// Add role
		r.Post("/", h.AddStaffRoleHandler)

		// Remove role
		r.Delete("/", h.RemoveStaffRoleHandler)
	})

	return r
}
