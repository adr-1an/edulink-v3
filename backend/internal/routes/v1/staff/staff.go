package staff

import (
	staffhandlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func SchoolStaffMemberRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List staff members
	r.Get("/", h.ListStaffMembersHandler)

	// Invitations
	r.Route("/invitations", func(r chi.Router) {
		// Create & send invitation
		r.Post("/", h.SendStaffInvitationHandler)
	})

	return r
}

func StaffMemberRoutes(h *staffhandlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{staffID}", func(r chi.Router) {
		// Delete staff member
		r.Delete("/", h.DeleteStaffMemberHandler)

		// Staff roles
		r.Mount("/roles", StaffRoleRoutes(h))
	})

	return r
}
