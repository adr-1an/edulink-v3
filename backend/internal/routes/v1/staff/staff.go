package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
	"github.com/sony/sonyflake/v2"
)

func SchoolStaffMemberRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
	s3 *minio.Client,
) chi.Router {
	r := chi.NewRouter()

	// List staff members
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListStaffMembersHandler(w, r, db, s3)
	})

	// Invitations
	r.Route("/invitations", func(r chi.Router) {
		// Create & send invitation
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			h.SendStaffInvitationHandler(w, r, db, sf)
		})
	})

	return r
}

func StaffMemberRoutes(db *sql.DB, sf *sonyflake.Sonyflake) chi.Router {
	r := chi.NewRouter()

	r.Route("/{staffID}", func(r chi.Router) {
		// Delete staff member
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			h.DeleteStaffMemberHandler(w, r, db, sf)
		})

		// Staff roles
		r.Mount("/roles", StaffRoleRoutes(db, sf))
	})

	return r
}
