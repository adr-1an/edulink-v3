package staff

import (
	h "app/internal/handlers/staff"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sony/sonyflake/v2"
)

func StaffInvitationRoutes(
	db *sql.DB,
	sf *sonyflake.Sonyflake,
) chi.Router {
	r := chi.NewRouter()

	// List user invitations
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		h.ListUserInvitationsHandler(w, r, db)
	})

	// Reject invitation by ID
	r.Post("/by-id/{invitationID}/reject", func(w http.ResponseWriter, r *http.Request) {
		h.RejectStaffInvitationByIDHandler(w, r, db, sf)
	})

	// Accept invitation by ID
	r.Post("/by-id/{invitationID}/accept", func(w http.ResponseWriter, r *http.Request) {
		h.AcceptStaffInvitationByIDHandler(w, r, db, sf)
	})

	// View invitation
	r.Get("/{token}", func(w http.ResponseWriter, r *http.Request) {
		h.ViewInvitationHandler(w, r, db)
	})

	// Reject invitation
	r.Post("/{token}/reject", func(w http.ResponseWriter, r *http.Request) {
		h.RejectStaffInvitationHandler(w, r, db, sf)
	})

	// Accept invitation
	r.Post("/{token}/accept", func(w http.ResponseWriter, r *http.Request) {
		h.AcceptStaffInvitationHandler(w, r, db, sf)
	})

	// Cancel invitation
	r.Post("/{invitationID}/cancel", func(w http.ResponseWriter, r *http.Request) {
		h.CancelSchoolInvitationHandler(w, r, db, sf)
	})

	return r
}
