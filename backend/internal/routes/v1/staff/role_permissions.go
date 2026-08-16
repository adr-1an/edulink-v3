package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func RolePermissionsRouter(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	// Set permission
	r.Put("/", func(w http.ResponseWriter, r *http.Request) {
		h.SetStaffRolePermissionHandler(w, r, db, sf)
	})

	return r
}
