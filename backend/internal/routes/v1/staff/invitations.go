package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func invitationRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List user invitations
	r.Get("/", h.ListUserInvitations)

	// Reject invitation by ID
	r.Post("/by-id/{invitationID}/reject", h.RejectStaffInvitationByID)

	// Accept invitation by ID
	r.Post("/by-id/{invitationID}/accept", h.AcceptStaffInvitationByID)

	// View invitation
	r.Get("/{token}", h.ViewInvitation)

	// Reject invitation
	r.Post("/{token}/reject", h.RejectStaffInvitation)

	// Accept invitation
	r.Post("/{token}/accept", h.AcceptStaffInvitation)

	// Cancel invitation
	r.Post("/{invitationID}/cancel", h.CancelInvitation)

	return r
}
