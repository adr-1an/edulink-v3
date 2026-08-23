package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func StaffInvitationRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List user invitations
	r.Get("/", h.ListUserInvitationsHandler)

	// Reject invitation by ID
	r.Post("/by-id/{invitationID}/reject", h.RejectStaffInvitationByIDHandler)

	// Accept invitation by ID
	r.Post("/by-id/{invitationID}/accept", h.AcceptStaffInvitationByIDHandler)

	// View invitation
	r.Get("/{token}", h.ViewInvitationHandler)

	// Reject invitation
	r.Post("/{token}/reject", h.RejectStaffInvitationHandler)

	// Accept invitation
	r.Post("/{token}/accept", h.AcceptStaffInvitationHandler)

	// Cancel invitation
	r.Post("/{invitationID}/cancel", h.CancelSchoolInvitationHandler)

	return r
}
