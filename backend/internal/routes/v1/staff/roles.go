package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func RoleRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
) chi.Router {
	r := chi.NewRouter()

	// List role permissions
	r.Get("/permissions", func(w http.ResponseWriter, r *http.Request) {
		h.ListPermissionsHandler(w, r, db)
	})

	// List roles
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListRolesHandler(w, r, db)
	})

	// Role-specific
	r.Route("/{roleID}", func(r chi.Router) {
		// Update role
		r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
			h.UpdateRoleHandler(w, r, db, sf)
		})

		// Delete role
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			h.DeleteRoleHandler(w, r, db, sf)
		})

		// Role permissions
		r.Mount("/permissions", RolePermissionsRouter(db, sf))
	})

	return r
}

func SchoolRoleRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := RoleRoutes(db, sf)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		h.CreateRoleHandler(w, r, db, sf)
	})

	return r
}

func StaffRoleRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	// List roles
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListStaffRolesHandler(w, r, db)
	})

	r.Route("/{roleID}", func(r chi.Router) {
		// Add role
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			h.AddStaffRoleHandler(w, r, db, sf)
		})

		// Remove role
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			h.RemoveStaffRoleHandler(w, r, db, sf)
		})
	})

	return r
}
