package staff

import (
	handlers "app/internal/handlers/staff"

	"github.com/go-chi/chi/v5"
)

func schoolMemberRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	// List staff members
	r.Get("/", h.ListStaffMembers)

	// Invitations
	r.Route("/invitations", func(r chi.Router) {
		// Create & send invitation
		r.Post("/", h.SendStaffInvitation)
	})

	return r
}

func memberRoutes(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route("/{staffID}", func(r chi.Router) {
		// Delete staff member
		r.Delete("/", h.DeleteStaffMember)

		// Staff roles
		r.Mount("/roles", memberRoleRoutes(h))
	})

	return r
}
