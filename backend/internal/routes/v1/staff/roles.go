package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func roleRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List role permissions
	r.Get("/permissions", h.ListPermissions)

	// List roles
	r.Get("/", h.ListRoles)

	// Role-specific
	r.Route("/{roleID}", func(r chi.Router) {
		// Update role
		r.Patch("/", h.UpdateRole)

		// Delete role
		r.Delete("/", h.DeleteRole)

		// Role permissions
		r.Mount("/permissions", rolePermissionRoutes(h))
	})

	return r
}

func schoolRoleRoutes(h *handlers.Handler) chi.Router {
	r := roleRoutes(h)

	r.Post("/", h.CreateRole)

	return r
}

func memberRoleRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List roles
	r.Get("/", h.ListStaffRoles)

	r.Route("/{roleID}", func(r chi.Router) {
		// Add role
		r.Post("/", h.AddStaffRole)

		// Remove role
		r.Delete("/", h.RemoveStaffRole)
	})

	return r
}

func rolePermissionRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// Set permission
	r.Put("/", h.SetStaffRolePermission)

	return r
}
