package staff

import (
	"app/internal/application/staff/portal_users"
	"app/internal/helpers/staff"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) SendPortalUserAccountVerificationEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := staff_helpers.TokenToUserID(ctx, w, r, h.DB)
	if err != nil {
		return
	}

	portalUserID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.App.Staff.PortalUsers.SendActivationEmail(ctx, &portal_users.SendActivationEmailInput{
		UserID:       userID,
		PortalUserID: portalUserID,
	}); err != nil {
		switch {
		case errors.Is(err, portal_users.ErrForbidden):
			w.WriteHeader(http.StatusForbidden)
		case errors.Is(err, portal_users.ErrAccountDisabled):
			w.WriteHeader(http.StatusUnprocessableEntity)
		case errors.Is(err, portal_users.ErrAccountAlreadyActive):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
